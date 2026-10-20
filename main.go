package main

import (
	"fmt"
	"self_medie_script_tool/config"
	"self_medie_script_tool/config/prompt"
	"self_medie_script_tool/logger"
	"self_medie_script_tool/logic"
	"self_medie_script_tool/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化配置（viper）
	if err := config.Init("./config/config.yaml", false); err != nil {
		panic("配置初始化失败:" + err.Error())
	}
	conf := config.Get()

	// 2. 初始化日志组件，传入配置里的log段（zap）
	logger.Init(conf.Log)
	defer logger.Sync() // 关键！退出刷日志

	// 初始化PromptManager
	pm := prompt.NewPromptManager(conf.Prompt.BaseDir, logger.ZapLogger)
	defer pm.Close() // 优雅关闭释放watcher

	// 预加载所有uses模板，自动根据hot_reload开关启动监听
	if err := pm.PreLoadFromConfig(conf.Prompt.Uses, conf.Prompt.HotReload); err != nil {
		panic(fmt.Sprintf("prompt preload failed: %v", err))
	}

	// 初始化OpenAI客户端
	logic.Init(conf)

	// 启动服务（gin）
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 顺序：trace中间件最先注册
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.GinZap())
	r.Use(middleware.RecoveryWithZap()) // 替换原生gin.Recovery()

	r.POST(conf.Server.Path, handler(pm))
	if err := r.Run(fmt.Sprintf(":%d", conf.Server.Port)); err != nil {
		panic("sse服务启动失败:" + err.Error())
	}
}
