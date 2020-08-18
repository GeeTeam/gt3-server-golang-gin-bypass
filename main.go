package main

import (
	"github.com/gin-gonic/gin"
	"gt3-server-golang-gin-sdk/controllers"
)

func main() {
	// 启动轮询从geetest获取当前bypass状态
	go controllers.CheckBypassStatus()

	r := gin.Default()
	// 设置静态资源
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// 设置web路由
	r.GET("/register", controllers.FirstRegister)
	r.POST("/validate", controllers.SecondValidate)
	r.Run(":8000")
}
