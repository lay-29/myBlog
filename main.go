package main

import "github.com/gin-gonic/gin"

func main() {
	// 1. 创建 Gin 引擎（自带日志和 panic 恢复）
	r := gin.Default()
	// 2. 定义一个 GET API
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 3. 启动服务
	r.Run(":8080") // http://localhost:8080
}
