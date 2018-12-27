package controller

import (
	"config"
	"constant"
	"fmt"
	"model"
	"util"

	"github.com/graphql-go/graphql"
)

var (
	createFeedbackArgs = graphql.FieldConfigArgument{
		"content": &graphql.ArgumentConfig{
			Description: "内容",
			Type:        graphql.String,
		},
		"imgs": &graphql.ArgumentConfig{
			Type:        graphql.NewList(imgType),
			Description: "图片",
		},
		"contactWay": &graphql.ArgumentConfig{
			Type:        graphql.String,
			Description: "联系方式",
		},
	}
)

func createFeedback(p graphql.ResolveParams) (interface{}, error) {
	feedback := model.Feedback{}
	err := util.MapToJSONStruct(p.Args, &feedback)
	if err != nil {
		return false, err
	}
	// 提醒管理员有人反馈了
	go func() {
		imgHTML := "<img src=\"%s\"  alt=\"反馈图片\" />"
		content := fmt.Sprintf(constant.EmailFeedbackNotice, feedback.ContactWay, feedback.Content)
		for _, img := range feedback.Imgs {
			content += "<br />" + fmt.Sprintf(imgHTML, img.URL)
		}
		util.SendEmail("小灵通", "小灵通反馈", content, config.Conf.EmailInfo.To)
	}()

	return true, nil
}

func writeFeedbackLog(funcName, errMsg string, err error) {
	writeLog("feedback.go", funcName, errMsg, err)
}
