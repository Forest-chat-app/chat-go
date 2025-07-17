package initialize

import (
	"chat-server/global"
	"chat-server/middleware"
	"chat-server/router"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRouter 初始化所有路由
func InitRouter() {
	global.CHAT_ROUTERS = gin.Default()

	// 全局中间件
	global.CHAT_ROUTERS.Use(middleware.Cors())

	// 使用带排除路径的JWT认证中间件，对所有路由进行认证
	global.CHAT_ROUTERS.Use(middleware.JWTAuth())

	// 添加Swagger路由 - 不需要认证
	global.CHAT_ROUTERS.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 初始化API v1路由组
	apiV1 := global.CHAT_ROUTERS.Group("/api/v1")

	// 初始化各个模块的路由
	router.RouterGroupApp.UserRouter.InitUserRouter(apiV1)
	router.RouterGroupApp.ChatRouter.InitChatRouter(apiV1)
}
