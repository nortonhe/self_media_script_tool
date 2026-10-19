package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"self_medie_script_tool/config"
	"self_medie_script_tool/logger"
	"self_medie_script_tool/svrcontext"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// openaiClient is a global variable to hold the OpenAI client instance.
var openaiClient *openai.Client

// Init initializes the OpenAI client with the provided configuration.
func Init(conf config.Config) {
	clientConfig := openai.DefaultConfig(conf.LLM.APIKey)
	clientConfig.BaseURL = conf.LLM.BaseURL
	openaiClient = openai.NewClientWithConfig(clientConfig)
}

// Handler handles requests and streams chat responses to the client.
func Handler(c *svrcontext.ServerContext) {
	if err := chat(c); err != nil {
		logger.TraceError(c.GetTraceID(), "处理请求失败", zap.Error(err))
		errorMessage, marshalErr := json.Marshal(map[string]string{"error": err.Error()})
		if marshalErr != nil {
			logger.TraceError(c.GetTraceID(), "编码错误响应失败", zap.Error(marshalErr))
			errorMessage = []byte(`{"error":"request failed"}`)
		}
		_, _ = fmt.Fprintf(c.GetGinContext().Writer, "data: %s\n\n", errorMessage)
		_, _ = fmt.Fprint(c.GetGinContext().Writer, "data: [DONE]\n\n")
		c.GetGinContext().Writer.Flush()
	}
}

func chat(c *svrcontext.ServerContext) error {
	request := buildRequest(c)
	// 继承gin请求ctx，再增加30s硬超时
	ctx, cancel := context.WithTimeout(c.GetGinContext().Request.Context(), 30*time.Second)
	defer cancel()

	if !request.Stream {
		response, err := openaiClient.CreateChatCompletion(ctx, request)
		if err != nil {
			return fmt.Errorf("call ollama chat failed: %w", err)
		}
		if len(response.Choices) == 0 {
			return errors.New("ollama chat returned no choices")
		}
		if err := writeContent(c, response.Choices[0].Message.Content); err != nil {
			return err
		}
		return finishResponse(c)
	}

	stream, err := openaiClient.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return fmt.Errorf("start ollama chat stream failed: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read ollama chat stream failed: %w", err)
		}

		logger.TraceDebug(c.GetTraceID(), "收到聊天流结果", zap.Any("result", response))
		for _, choice := range response.Choices {
			if err := writeContent(c, choice.Delta.Content); err != nil {
				return err
			}
		}
	}

	return finishResponse(c)
}

func buildRequest(c *svrcontext.ServerContext) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: c.GetConfig().LLM.Model,
		Temperature: c.GetConfig().LLM.Temperature,
		MaxTokens: c.GetConfig().LLM.MaxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: c.GetQuery(),
			},
		},
		Stream: c.GetConfig().LLM.Stream,
	}
}

func writeContent(c *svrcontext.ServerContext, content string) error {
	if content == "" {
		return nil
	}

	c.AppendResponse(content)
	if _, err := fmt.Fprintf(c.GetGinContext().Writer, "data: %s\n\n", content); err != nil {
		return fmt.Errorf("write stream data failed: %w", err)
	}
	c.GetGinContext().Writer.Flush()
	return nil
}

func finishResponse(c *svrcontext.ServerContext) error {
	if _, err := fmt.Fprint(c.GetGinContext().Writer, "data: [DONE]\n\n"); err != nil {
		return fmt.Errorf("write end signal failed: %w", err)
	}
	c.GetGinContext().Writer.Flush()
	c.SetResult()
	logger.TraceDebug(c.GetTraceID(), "最终结果", zap.String("result", c.GetResult()))
	return nil
}
