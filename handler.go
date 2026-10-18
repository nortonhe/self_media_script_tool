package main

import (
	"self_medie_script_tool/model/ollama"

	"github.com/gin-gonic/gin"
)

func handler(c *gin.Context) {
	// 1. 设置 SSE 必需的响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 2. 调用 Ollama 模型处理请求
	ollama.Handler(c)
}