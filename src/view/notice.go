package view

import (
	"controller"

	"github.com/labstack/echo"
)

func InitNoticeView(group *echo.Group) {
	group.PUT("", controller.UpdateNotice)
	group.DELETE("", controller.DeleteNotice)

	group.GET("/list", controller.GetNotices)
	group.POST("/list", controller.CreateNotices)
}
