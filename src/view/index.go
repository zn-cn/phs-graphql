/*
The index view contains the index view and other view that uses `/` to begin
*/
package view

import (
	"github.com/labstack/echo"
)

func InitViewV1(group *echo.Group) {

	feedback := group.Group("/feedback")
	InitFeedbackView(feedback)

	groupView := group.Group("/group")
	InitGroupView(groupView)

	notice := group.Group("/notice")
	InitNoticeView(notice)

	template := group.Group("/template")
	InitTemplateView(template)

	user := group.Group("/user")
	InitUserView(user)
}
