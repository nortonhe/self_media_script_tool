package main

import (
	"fmt"
	"self_medie_script_tool/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := config.Init("./conf/config.yaml", false); err != nil {
		panic(err)
	}
	conf := config.Get()

	// 启动服务
	r := gin.Default()
	r.POST(conf.Server.Path, handler)
	if err := r.Run(fmt.Sprintf(":%d", conf.Server.Port)); err != nil {
		panic(err)
	}
}