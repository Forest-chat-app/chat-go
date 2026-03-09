package router

import (
	"chat-server/api/v1"
	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

// InitUserRouter 初始化用户相关路由
func (s *UserRouter) InitUserRouter(apiV1 *gin.RouterGroup) {
	userGroup := apiV1.Group("/user")
	{
		userGroup.POST("/register", v1.ApiGroupApp.Register)
		userGroup.POST("/loginAccount", v1.ApiGroupApp.LoginAccount)
		userGroup.GET("/getUserInfo", v1.ApiGroupApp.GetUserInfo)
		userGroup.PUT("/updateUserProfile", v1.ApiGroupApp.UpdateUserProfile)
		userGroup.PUT("/updateUserPassword", v1.ApiGroupApp.UpdateUserPassword)
		userGroup.POST("/logout", v1.ApiGroupApp.Logout)
		userGroup.GET("/test", v1.ApiGroupApp.Test)
	}
}
