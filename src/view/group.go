package view

import (
	"controller"

	"github.com/labstack/echo"
)

func InitGroupView(group *echo.Group) {
	group.GET("/list", controller.GetGroups)
	group.POST("", controller.CreateGroup)
	group.POST("/action/join", controller.JoinGroup)
	group.POST("/action/leave", controller.LeaveGroup)
	group.PUT("/member/list", controller.UpdateGroupMembers)
	group.GET("/member/list", controller.GetGroupMembers)
	group.GET("/qrcode", controller.GetGroupQrcode)
}
