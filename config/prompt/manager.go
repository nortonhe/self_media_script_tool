package prompt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v3"
)

var tplFileRegex = regexp.MustCompile(`^(.+)_(v\d+)\.yaml$`)

type PromptManager struct {
	baseDir    string
	cache      map[string]*PromptTemplate // key: templateId:version
	logger     *zap.Logger
	mu         sync.RWMutex
	watcher    *fsnotify.Watcher
	ctx        context.Context
	cancel     context.CancelFunc
	uses   []PromptInfo // 保存当前主配置的uses列表，监听用
	debounceMu sync.Mutex
	debounceMap map[string]time.Time // 防抖，防止编辑器多次触发write事件
}

func NewPromptManager(baseDir string, logger *zap.Logger) *PromptManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &PromptManager{
		baseDir:    baseDir,
		logger:     logger,
		cache:      make(map[string]*PromptTemplate),
		ctx:        ctx,
		cancel:     cancel,
		debounceMap: make(map[string]time.Time),
	}
}

// PreLoadFromConfig 启动预加载，同时保存uses引用，用于热加载监听
func (m *PromptManager) PreLoadFromConfig(uses []PromptInfo, hotReload bool) error {
	m.uses = uses
	for _, ref := range uses {
		_, err := m.Load(ref.TemplateID, ref.Version)
		if err != nil {
			return fmt.Errorf("preload template %s:%s failed: %w", ref.TemplateID, ref.Version, err)
		}
	}

	// 如果开启热加载，启动文件监听器
	if hotReload {
		if err := m.startWatcher(); err != nil {
			return fmt.Errorf("start prompt watcher failed: %w", err)
		}
		m.logger.Info("prompt hot reload enabled")
	}
	return nil
}

func (m *PromptManager) startWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	m.watcher = watcher

	// 只监听uses里对应的模板文件，不是整个目录
	for _, ref := range m.uses {
		fileName := fmt.Sprintf("%s_%s.yaml", ref.TemplateID, ref.Version)
		fullPath := filepath.Join(m.baseDir, fileName)
		if err := watcher.Add(fullPath); err != nil {
			m.logger.Warn("add prompt file to watch failed", zap.String("path", fullPath), zap.Error(err))
		}
	}

	// 后台协程处理文件变更事件
	go m.watchLoop()
	return nil
}

func (m *PromptManager) watchLoop() {
	const debounceDuration = 200 * time.Millisecond
	for {
		select {
		case <-m.ctx.Done():
			return
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			// 只处理写入、保存事件
			if !event.Has(fsnotify.Write) {
				continue
			}
			filePath := event.Name

			m.debounceMu.Lock()
			now := time.Now()
			last, exist := m.debounceMap[filePath]
			if exist && now.Sub(last) < debounceDuration {
				m.debounceMu.Unlock()
				continue
			}
			m.debounceMap[filePath] = now
			m.debounceMu.Unlock()

			// 防抖延迟再加载
			go func(path string) {
				time.Sleep(debounceDuration)
				m.reloadSingleFile(path)
			}(filePath)

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			m.logger.Error("prompt watch error", zap.Error(err))
		}
	}
}

// reloadSingleFile 单个文件热重载，失败保留旧缓存
func (m *PromptManager) reloadSingleFile(fullPath string) {
	fileName := filepath.Base(fullPath)
	sub := tplFileRegex.FindStringSubmatch(fileName)
	if sub == nil {
		m.logger.Warn("invalid prompt file name, skip reload", zap.String("file", fileName))
		return
	}
	tplID, ver := sub[1], sub[2]

	// 尝试重新加载
	newTpl, err := m.loadRaw(tplID, ver)
	if err != nil {
		m.logger.Error("hot reload prompt failed, keep old template",
			zap.String("template_id", tplID),
			zap.String("version", ver),
			zap.Error(err))
		return
	}

	// 写入新缓存，原子替换
	cacheKey := fmt.Sprintf("%s:%s", tplID, ver)
	m.mu.Lock()
	m.cache[cacheKey] = newTpl
	m.mu.Unlock()

	m.logger.Info("prompt hot reload success", zap.String("template_id", tplID), zap.String("version", ver))
}

// loadRaw 底层加载逻辑，不读缓存，直接从文件读取并校验
func (m *PromptManager) loadRaw(tplID, ver string) (*PromptTemplate, error) {
	fileName := fmt.Sprintf("%s_%s.yaml", tplID, ver)
	fullPath := filepath.Join(m.baseDir, fileName)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read file err: %w", err)
	}
	var tpl PromptTemplate
	if err := yaml.Unmarshal(data, &tpl); err != nil {
		return nil, fmt.Errorf("yaml unmarshal err: %w", err)
	}
	sub := tplFileRegex.FindStringSubmatch(fileName)
	if sub == nil {
		return nil, fmt.Errorf("filename invalid")
	}
	parsedID, parsedVer := sub[1], sub[2]
	if parsedID != tplID || parsedVer != ver {
		return nil, fmt.Errorf("filename mismatch")
	}
	if tpl.TemplateID != tplID || tpl.Version != ver {
		return nil, fmt.Errorf("yaml inner id/version mismatch")
	}
	if err := tpl.PreCompile(); err != nil {
		return nil, fmt.Errorf("template compile failed: %w", err)
	}
	return &tpl, nil
}

// Load 优先读缓存，不存在则调用loadRaw加载并写入缓存
func (m *PromptManager) Load(tplID, ver string) (*PromptTemplate, error) {
	cacheKey := fmt.Sprintf("%s:%s", tplID, ver)
	m.mu.RLock()
	item, ok := m.cache[cacheKey]
	m.mu.RUnlock()
	if ok {
		return item, nil
	}

	newTpl, err := m.loadRaw(tplID, ver)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	m.cache[cacheKey] = newTpl
	m.mu.Unlock()
	return newTpl, nil
}

func (m *PromptManager) Get(tplID, ver string) (*PromptTemplate, error) {
	cacheKey := fmt.Sprintf("%s:%s", tplID, ver)
	m.mu.RLock()
	tpl, ok := m.cache[cacheKey]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("template %s:%s not loaded, check prompt.uses config", tplID, ver)
	}
	return tpl, nil
}

// Close 关闭监听器、释放资源，优雅退出调用
func (m *PromptManager) Close() error {
	m.cancel()
	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}
