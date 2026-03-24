package admin

import (
	adminApi "chat-server/api/v1/admin"
	"chat-server/middleware"

	"github.com/gin-gonic/gin"
)

type AdminRouter struct{}

func (r *AdminRouter) InitAdminRouter(root *gin.RouterGroup) {
	// 管理端路由组
	adminGroup := root.Group("/admin")

	// 登录不需要认证
	adminGroup.POST("/login", adminApi.AdminAuthApiApp.Login)

	// 以下路由需要管理员 JWT 认证
	auth := adminGroup.Group("", middleware.AdminJWTAuth())
	{
		// 认证/个人信息
		auth.GET("/self", adminApi.AdminAuthApiApp.GetSelf)

		// 用户管理
		userGroup := auth.Group("/user")
		{
			userGroup.GET("/list", adminApi.AdminUserApiApp.ListUsers)
			userGroup.GET("/detail", adminApi.AdminUserApiApp.GetUserDetail)
			userGroup.PUT("/profile", adminApi.AdminUserApiApp.UpdateUserProfile)
			userGroup.PUT("/avatar", adminApi.AdminUserApiApp.UpdateUserAvatar)
			userGroup.PUT("/resetPassword", adminApi.AdminUserApiApp.ResetUserPassword)
			userGroup.DELETE("/ban", adminApi.AdminUserApiApp.BanUser)
		}

		// 房间管理
		roomGroup := auth.Group("/room")
		{
			roomGroup.GET("/list", adminApi.AdminRoomApiApp.ListRooms)
			roomGroup.GET("/detail", adminApi.AdminRoomApiApp.GetRoomDetail)
			roomGroup.POST("/create", adminApi.AdminRoomApiApp.CreateRoom)
			roomGroup.PUT("/info", adminApi.AdminRoomApiApp.UpdateRoomInfo)
			roomGroup.PUT("/avatar", adminApi.AdminRoomApiApp.UpdateRoomAvatar)
			roomGroup.DELETE("/delete", adminApi.AdminRoomApiApp.DeleteRoom)
			roomGroup.GET("/members", adminApi.AdminRoomApiApp.GetRoomMembers)
			roomGroup.DELETE("/members/remove", adminApi.AdminRoomApiApp.RemoveRoomMember)
		}

		// 角色管理
		roleGroup := auth.Group("/role")
		{
			roleGroup.GET("/list", adminApi.AdminRoleApiApp.ListRoles)
			roleGroup.POST("/create", adminApi.AdminRoleApiApp.CreateRole)
			roleGroup.PUT("/update", adminApi.AdminRoleApiApp.UpdateRole)
			roleGroup.DELETE("/delete", adminApi.AdminRoleApiApp.DeleteRole)
		}

		// 管理员管理（仅超级管理员）
		adminMgmtGroup := auth.Group("/admin-mgmt", middleware.SuperAdminOnly())
		{
			adminMgmtGroup.GET("/list", adminApi.AdminRoleApiApp.ListAdmins)
			adminMgmtGroup.POST("/create", adminApi.AdminRoleApiApp.CreateAdmin)
			adminMgmtGroup.DELETE("/delete", adminApi.AdminRoleApiApp.DeleteAdmin)
		}

		// 消息管理
		messageGroup := auth.Group("/message")
		{
			messageGroup.GET("/list", adminApi.AdminMessageApiApp.ListMessages)
			messageGroup.GET("/search", adminApi.AdminMessageApiApp.SearchMessages)
			messageGroup.DELETE("/delete", adminApi.AdminMessageApiApp.DeleteMessage)
		}
	}
}
