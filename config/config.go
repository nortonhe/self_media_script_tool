package config

import (
	"errors"
	"fmt"
	"self_medie_script_tool/config/prompt"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// 全局对外只读实例，业务读取配置用这个
var GlobalConf *Config
var validate *validator.Validate
var once sync.Once
var mu sync.RWMutex // 热更新读写锁

// Config 总配置结构体，mapstructure标签是viper反序列化必须的
type Config struct {
	Server ServerConfig `mapstructure:"server" validate:"required"`
	LLM    LLMConfig    `mapstructure:"llm" validate:"required"`
	Prompt PromptConfig `mapstructure:"prompt"`
	Log    LogConfig    `mapstructure:"log" validate:"required"`
}

type ServerConfig struct {
	Path        string `mapstructure:"path" validate:"required"`
	Port        int    `mapstructure:"port" validate:"required,min=1,max=65535"`
}

type LLMConfig struct {
	APIKey string `mapstructure:"apikey" validate:"required"`
	BaseURL   string `mapstructure:"baseURL" validate:"required,url"`
}

type PromptConfig struct {
	BaseDir   string          `mapstructure:"baseDir"`
	HotReload bool            `mapstructure:"hot_reload"`
	Uses      []prompt.PromptInfo `mapstructure:"uses"`
}

type LogConfig struct {
	Level      string `mapstructure:"level" validate:"oneof=debug info warn error"` //日志级别
	File       string `mapstructure:"file"`          //日志文件路径，空则只输出stdout
	MaxSize    int    `mapstructure:"maxSize"`      //MB，单个文件最大大小
	MaxBackups int    `mapstructure:"maxBackups"`   //保留旧文件数量
	MaxAge     int    `mapstructure:"maxAge"`       //保留天数
	Compress   bool   `mapstructure:"compress"`     //是否gzip压缩旧日志
	Console    bool   `mapstructure:"console"` // 新增：是否输出控制台
}


// Init 初始化配置，入参：配置文件路径；enableWatch 是否开启热加载
func Init(configFile string, enableWatch bool) error {
	var err error
	once.Do(func() {
		validate = validator.New()
	})

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	// 开启环境变量自动覆盖
	v.AutomaticEnv()
	v.SetEnvPrefix("APP")        // 环境变量前缀 APP_
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // server.port => APP_SERVER_PORT

	// 读取配置文件
	if err = v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	// 反序列化到结构体
	var conf Config
	if err = v.Unmarshal(&conf); err != nil {
		return fmt.Errorf("unmarshal配置失败: %w", err)
	}

	// 结构体校验（必填项、范围校验）
	if err = validate.Struct(&conf); err != nil {
		var validationErrs []string
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, e := range ve {
				validationErrs = append(validationErrs, e.Field()+" "+e.Tag())
			}
		}
		return fmt.Errorf("配置校验失败: %v", validationErrs)
	}

	mu.Lock()
	GlobalConf = &conf
	mu.Unlock()

	// 开启热加载
	if enableWatch {
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			fmt.Printf("配置文件变更:%s\n", e.Name)
			var newConf Config
			if err := v.Unmarshal(&newConf); err != nil {
				fmt.Printf("配置重载失败: %v\n", err)
				return
			}
			if err := validate.Struct(&newConf); err != nil {
				fmt.Printf("重载后的配置校验失败: %v\n", err)
				return
			}
			// 原子替换
			mu.Lock()
			GlobalConf = &newConf
			mu.Unlock()
			fmt.Println("✅ 配置热更新成功")
		})
	}

	return nil
}

// Get 获取配置副本（可选，防止外部直接修改GlobalConf）
func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return *GlobalConf
}
