package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"self_medie_script_tool/context"
	"self_medie_script_tool/logger"
	"strings"

	"go.uber.org/zap"
)

// Handler 处理前端请求，流式返回 Ollama 中千问模型的响应
func Handler(c *context.ServerContext) {
	if err := callModel(c); err != nil {
		c.GetGinContext().Writer.WriteString("data: {\"error\":\"" + err.Error() + "\"}\n\n")
		c.GetGinContext().Writer.WriteString("data: [DONE]\n\n")
		c.GetGinContext().Writer.Flush()
		return
	}
}

// callModel 调用本地 Ollama 的模型进行流式处理
func callModel(c *context.ServerContext) error {
	payload := map[string]any{
		"model":  c.GetConfig().Ollama.Model,
		"prompt": c.GetQuery(),
		"stream": c.GetConfig().Ollama.Stream,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request body failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.GetConfig().Ollama.URL, bytes.NewReader(body))
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
		// 打印日志方便观察
		logger.TraceDebug(c.GetTraceID(), "推送结果", zap.Any("result", result))
		if result.Error != "" {
			return fmt.Errorf("ollama error: %s", result.Error)
		}

		if result.Response != "" {
			c.AppendResponse(result.Response)
			_, err = fmt.Fprintf(c.GetGinContext().Writer, "data: %s\n\n", result.Response)
			if err != nil {
				return fmt.Errorf("write stream data failed: %w", err)
			}
			c.GetGinContext().Writer.Flush()
		}
		if result.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read ollama stream failed: %w", err)
	}

	_, err = fmt.Fprint(c.GetGinContext().Writer, "data: [DONE]\n\n")
	if err != nil {
		return fmt.Errorf("write end signal failed: %w", err)
	}
	c.GetGinContext().Writer.Flush()
	c.SetResult()
	logger.TraceDebug(c.GetTraceID(), "最终结果", zap.String("result", c.GetResult()))
	return nil
}