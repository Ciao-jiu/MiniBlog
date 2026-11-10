package main

import (
	"MiniBlog/database"
	"MiniBlog/router"
)

func main() {
	database.InitDB()
	database.InitRedis()
	//组装路由 (加载所有 handlers 和 middleware)
	r := router.SetupRouter()
	//修复 Gin 的安全警告：不信任任何代理头
	r.SetTrustedProxies(nil)
	r.Run(":8080")
}
