package main

import (
	"fmt"
	"self_medie_script_tool/config"
	"self_medie_script_tool/logger"

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

	// 启动服务（gin）
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(logger.GinZap())
	r.Use(logger.RecoveryWithZap()) // 替换原生gin.Recovery()

	r.POST(conf.Server.Path, handler)
	if err := r.Run(fmt.Sprintf(":%d", conf.Server.Port)); err != nil {
		panic("sse服务启动失败:" + err.Error())
	} else {
		logger.Sugar.Infof("sse服务启动成功，监听端口: %d", conf.Server.Port)
	}
}