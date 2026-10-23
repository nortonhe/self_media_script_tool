package prompt

import (
	"fmt"
	"strings"
	"text/template"
)

// PromptInfo 用于在主配置中引用具体的prompt模板
type PromptInfo struct {
	TemplateID string `mapstructure:"templateID"`
	Version    string `mapstructure:"version"`
}

// PromptTemplate 代表一个prompt模板的结构体
type PromptTemplate struct {
	TemplateID         string `yaml:"template_id"`
	Version            string `yaml:"version"`
	Name               string `yaml:"name"`
	Desc               string `yaml:"desc"`

	// ========== LLM推理参数 ==========
	Model       string  `yaml:"model"`        // 模型名称
	Temperature float32 `yaml:"temperature"`  // 温度参数，控制生成文本的随机性，范围为 0.0 到 1.0
	TopP        float32 `yaml:"top_p"` 		  // top_p 参数，控制生成文本的多样性，范围为 0.0 到 1.0
	MaxTokens   int     `yaml:"max_tokens"`   // 最大生成 token 数量
	Stream      bool    `yaml:"stream"`       // 是否流式返回
	Timeout 	int `yaml:"timeout"`          // 请求超时时间，单位为秒

	SystemPrompt       string `yaml:"system_prompt"`
	UserPromptTemplate string `yaml:"user_prompt_template"`

	sysTpl  *template.Template
	userTpl *template.Template
}

// RenderSystem 渲染系统提示模板
func (t *PromptTemplate) RenderSystem(data any) (string, error) {
	var buf strings.Builder
	if err := t.sysTpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderUser 渲染用户提示模板
func (t *PromptTemplate) RenderUser(data any) (string, error) {
	var buf strings.Builder
	if err := t.userTpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// PreCompile 预编译模板，避免每次渲染时都重新解析模板
func (t *PromptTemplate) PreCompile() error {
	var err error
	// 自定义分隔符：左[[ 右]]
	t.sysTpl, err = template.New("sys").
		Delims("[[", "]]").
		Parse(t.SystemPrompt)
	if err != nil {
		return fmt.Errorf("compile system prompt template failed: %w", err)
	}

	t.userTpl, err = template.New("user").
		Delims("[[", "]]").
		Parse(t.UserPromptTemplate)
	return err
}
