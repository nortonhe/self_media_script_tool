package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"self_medie_script_tool/config"
	"strings"

	"github.com/gin-gonic/gin"
)

var model = "qwen2.5:7b"
var url = "http://127.0.0.1:11434/api/generate"

// Handler 处理前端请求，流式返回 Ollama 中千问模型的响应
func Handler(c *gin.Context) {
	query := strings.TrimSpace(c.PostForm("query"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query 参数不能为空"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Status(http.StatusOK)

	fmt.Printf("收到查询请求: %s\n", query)

	if err := callModel(query, c); err != nil {
		c.Writer.WriteString("data: {\"error\":\"" + err.Error() + "\"}\n\n")
		c.Writer.WriteString("data: [DONE]\n\n")
		c.Writer.Flush()
		return
	}
}

// callModel 调用本地 Ollama 的模型进行流式处理
func callModel(query string, c *gin.Context) error {
	conf := config.Get()
	payload := map[string]any{
		"model":  conf.Ollama.Model,
		"prompt": query,
		"stream": conf.Ollama.Stream,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request body failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, conf.Ollama.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("call ollama failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status code %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var result struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
			Error    string `json:"error"`
		}
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			return fmt.Errorf("decode ollama stream chunk failed: %w", err)
		}
		if result.Error != "" {
			return fmt.Errorf("ollama error: %s", result.Error)
		}

		// 打印日志方便观察
		fmt.Printf("推送结果: %s\n", result.Response)
		if result.Response != "" {
			_, err = fmt.Fprintf(c.Writer, "data: %s\n\n", result.Response)
			if err != nil {
				return fmt.Errorf("write stream data failed: %w", err)
			}
			c.Writer.Flush()
		}
		if result.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read ollama stream failed: %w", err)
	}

	_, err = fmt.Fprint(c.Writer, "data: [DONE]\n\n")
	if err != nil {
		return fmt.Errorf("write end signal failed: %w", err)
	}
	c.Writer.Flush()
	return nil
}