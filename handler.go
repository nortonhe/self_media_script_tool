package main

import (
	"net/http"
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

	// 2. 调用 Ollama 模型处理请求
	ollama.Handler(c)
}