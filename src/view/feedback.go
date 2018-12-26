package view

import (
	"controller"

	"github.com/labstack/echo"
)

// InitFeedbackView init feedback view
func InitFeedbackView(group *echo.Group) {
	group.POST("", controller.CreateFeedback)
}
