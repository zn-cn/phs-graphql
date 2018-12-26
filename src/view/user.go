package view

import (
	"controller"

	"github.com/labstack/echo"
)

func InitUserView(group *echo.Group) {
	group.GET("/status/follow", controller.GetUserFollowStatus)
}
