package svrcontext

import (
	"fmt"
	"net/http"
	"self_medie_script_tool/config"
	"self_medie_script_tool/config/prompt"
	"self_medie_script_tool/logger"
	"self_medie_script_tool/observability"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServerContext 服务全局上下文封装了gin.Context和配置对象，方便在handler里使用
type ServerContext struct {
	ginCtx *gin.Context
	conf   config.Config
	pm     *prompt.PromptManager
	metrics *observability.LLMMetrics

	request Request
	result  string
	traceID string
}

// Request represents the expected structure of the incoming request parameters.
type Request struct {
	ProductInfo   string `json:"ProductInfo"`   // 产品信息
	TargetUser    string `json:"TargetUser"`    // 目标用户
	VideoDuration string `json:"VideoDuration"` // 视频时长
	BloggerStyle  string `json:"BloggerStyle"`  // 博主风格
	ExtraReq      string `json:"ExtraReq"`      // 额外要求
}

// BuildServerContext 构建服务上下文，提取请求参数和traceID
func BuildServerContext(ginCtx *gin.Context, pm *prompt.PromptManager, metrics *observability.LLMMetrics) (*ServerContext, error) {
	svrCtx := &ServerContext{
		ginCtx:  ginCtx,
		conf:    config.Get(),
		pm:      pm,
		metrics: metrics,
		traceID: logger.GetTraceId(ginCtx.Request.Context()),
	}

	if err := ginCtx.ShouldBindJSON(&svrCtx.request); err != nil {
		ginCtx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return svrCtx, fmt.Errorf("bind request body: %w", err)
	}
	logger.TraceInfo(svrCtx.traceID, "收到查询请求", zap.Any("request", svrCtx.request))
	return svrCtx, nil
}

// AppendResponse 追加响应内容到ServerContext.result
func (s *ServerContext) AppendResponse(response string) {
	s.result += response
}

// GetResult 获取最终响应结果
func (s *ServerContext) GetResult() string {
	return s.result
}

// GetTraceID 获取当前请求的traceID
func (s *ServerContext) GetTraceID() string {
	return s.traceID
}

// GetGinContext 获取gin.Context
func (s *ServerContext) GetGinContext() *gin.Context {
	return s.ginCtx
}

// GetConfig 获取配置对象
func (s *ServerContext) GetConfig() config.Config {
	return s.conf
}

// GetRequest 获取从请求体解析得到的请求参数
func (s *ServerContext) GetRequest() Request {
	return s.request
}

// GetPromptManager 获取PromptManager实例
func (s *ServerContext) GetPromptManager() *prompt.PromptManager {
	return s.pm
}

func (s *ServerContext) GetMetrics() *observability.LLMMetrics {
	return s.metrics
}
