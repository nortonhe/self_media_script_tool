package svrcontext

import (
	"fmt"
	"net/http"
	"self_medie_script_tool/config"
	"self_medie_script_tool/logger"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServerContext 服务全局上下文封装了gin.Context和配置对象，方便在handler里使用
type ServerContext struct {
	ginCtx *gin.Context
	conf config.Config

	query string
	result string
	traceID string
}

// BuildServerContext 构建服务上下文，提取请求参数和traceID
func BuildServerContext(ginCtx *gin.Context) (*ServerContext, error) {
	svrCtx := &ServerContext{
		ginCtx: ginCtx,
		conf: config.Get(),
		traceID: logger.GetTraceId(ginCtx.Request.Context()),
		query: strings.TrimSpace(ginCtx.PostForm("query")),
	}

	if svrCtx.query == "" {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "query 参数不能为空"})
		return svrCtx, fmt.Errorf("query 参数不能为空")
	}
	logger.TraceInfo(svrCtx.traceID, "收到查询请求", zap.String("query", svrCtx.query))

	return svrCtx, nil
}

// AppendResponse 追加响应内容到ServerContext.result
func (s *ServerContext) AppendResponse(response string) {
	s.result += response
}

// SetResult 设置最终响应结果到ServerContext.result
func (s *ServerContext) SetResult() {
	s.ginCtx.Set(logger.ParamFinalResponse, s.result)
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

//	GetConfig 获取配置对象
func (s *ServerContext) GetConfig() config.Config {
	return s.conf
}

// GetQuery 获取请求的query参数
func (s *ServerContext) GetQuery() string {
	return s.query
}
