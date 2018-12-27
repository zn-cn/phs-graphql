package controller

import "github.com/graphql-go/graphql"

var (
	templateType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Template",
		Description: "Template",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.ID,
				Description: "id",
			},
			"type": &graphql.Field{
				Type:        graphql.Int,
				Description: "类型",
			},
			"status": &graphql.Field{
				Type:        graphql.Int,
				Description: "状态",
			},
			"creatorID": &graphql.Field{
				Type:        graphql.ID,
				Description: "创建者 unionid",
			},
			"createTime": &graphql.Field{
				Type:        graphql.Int,
				Description: "创建时间毫秒时间戳",
			},
			"notices": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "模板内容",
			},
		},
	})
)

func writeTemplateLog(funcName, errMsg string, err error) {
	writeLog("template.go", funcName, errMsg, err)
}
