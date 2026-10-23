package main

import (
	"encoding/json"
	"net/http"
	"self_medie_script_tool/logic"
	"self_medie_script_tool/observability"
	"self_medie_script_tool/svrcontext"

	"self_medie_script_tool/config/prompt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func handler(pm *prompt.PromptManager, metrics *observability.LLMMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 设置 SSE 必需的响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		// 2. 构建服务上下文，提取请求参数和traceID
		svrCtx, err := svrcontext.BuildServerContext(c, pm, metrics)
		if err != nil {
			// 错误已经在BuildServerContext里处理过响应了，这里直接返回
			return
		}

		// 3. 调用模型处理请求
		c.Status(http.StatusOK)
		logic.Handler(svrCtx)

		// http trace
		if httpSpan := trace.SpanFromContext(c.Request.Context()); httpSpan.IsRecording() {
			if requestBody, err := json.Marshal(svrCtx.GetRequest()); err == nil {
				httpSpan.SetAttributes(attribute.String("http.request.body", string(requestBody)))
			}
			httpSpan.SetAttributes(attribute.String("http.response.body", svrCtx.GetResult()))
		}
	}
}
