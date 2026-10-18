package main

import (
	"fmt"
	"self_medie_script_tool/conf"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := conf.Init("./conf/config.yaml", false); err != nil {
		panic(err)
	}
	c := conf.Get()

	// 启动服务
	r := gin.Default()
	r.POST(c.Server.Path, handler)
	if err := r.Run(fmt.Sprintf(":%d", c.Server.Port)); err != nil {
		panic(err)
	}
}