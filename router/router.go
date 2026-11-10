package router

import (
	"MiniBlog/handlers"
	"MiniBlog/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	//创建一个*gin.Default,它内部已经预置了Logger（日志）和 Recovery（panic 恢复）
	//两个全局中间件。所有请求都会先经过这两个中间件。
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.GET("/articles", handlers.GetArticles)
		//:id是Gin的路径参数语法，表示这个是动态值
		api.GET("/articles/:id", handlers.GetArticleByID)
		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)
	}

	//需要登录才能访问的路由
	//auth 的完整前缀继承自父组 api，即 /api/v1
	auth := api.Group("")
	//给当前路由组挂载中间件
	//中间件的执行顺序遵循洋葱模型：
	//请求由外向内穿过中间件，到达 handler 后再由内向外返回。
	auth.Use(middleware.AuthMiddleware())
	//执行顺序为：gin.Default() 的 Logger
	//gin.Default() 的 Recovery
	//AuthMiddleware（验证 Token，失败则 c.Abort() 并直接返回 401）
	//handlers.CreatedArticle
	{
		auth.PUT("/articles/:id", handlers.UpdateArticle)
		auth.DELETE("/articles/:id", handlers.DeletArticle)
		auth.POST("/articles", handlers.CreatedArticle)
	}
	return r
}
