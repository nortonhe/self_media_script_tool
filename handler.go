package main

import (
	"net/http"
	"self_medie_script_tool/context"
	"self_medie_script_tool/model/ollama"

	"github.com/gin-gonic/gin"
)

func handler(c *gin.Context) {
	// 1. 设置 SSE 必需的响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Status(http.StatusOK)

	// 2. 构建服务上下文，提取请求参数和traceID
	svrCtx, err := context.BuildServerContext(c)
	if err != nil {
		// 错误已经在BuildServerContext里处理过响应了，这里直接返回
		return
	}

	// 3. 调用 Ollama 模型处理请求
	ollama.Handler(svrCtx)
}