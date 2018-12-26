package controller

import "github.com/graphql-go/graphql"

var (
	query = graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"qiniuToken": &graphql.Field{
				Type:        qiniuTokenType,
				Description: "获取上传图片的七牛云upload-token 链接：https://developer.qiniu.com/kodo/manual/1208/upload-token",
				Args:        qiniuTokenArgs,
				Resolve:     getQiniuToken,
			},
		},
	})

	mutation = graphql.NewObject(graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: graphql.Fields{},
	})
)
