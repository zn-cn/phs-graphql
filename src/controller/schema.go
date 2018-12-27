package controller

import (
	"config"
	"constant"
	"context"
	"net/http"

	"github.com/graphql-go/graphql"
	gh "github.com/graphql-go/handler"
)

var (
	handler *gh.Handler

	query = graphql.NewObject(graphql.ObjectConfig{
		Name: "query",
		Fields: graphql.Fields{
			"health": &graphql.Field{
				Type:        graphql.String,
				Description: "判断健康情况",
				Resolve:     getHealth,
			},
			"qiniuToken": &graphql.Field{
				Args:        qiniuTokenArgs,
				Type:        qiniuTokenType,
				Description: "获取上传图片的七牛云upload-token 链接：https://developer.qiniu.com/kodo/manual/1208/upload-token",
				Resolve:     getQiniuToken,
			},
			"user": &graphql.Field{
				Args:        userArgs,
				Type:        userType,
				Description: "获取用户信息",
				Resolve:     getUserByUnionid,
			},
			"notice": &graphql.Field{
				Args:        idArgs,
				Type:        noticeType,
				Description: "获取通知信息",
				Resolve:     getNotice,
			},
			"notices": &graphql.Field{
				Args:        noticePageArgs,
				Type:        graphql.NewList(noticeType),
				Description: "获取通知列表",
				Resolve:     getNotices,
			},
			"group": &graphql.Field{
				Args:        codeArgs,
				Type:        groupType,
				Description: "获取圈子信息",
				Resolve:     getGroupByCode,
			},
			"template": &graphql.Field{
				Args:        idArgs,
				Type:        templateType,
				Description: "获取模板信息",
				Resolve:     getTemplate,
			},
		},
	})

	mutation = graphql.NewObject(graphql.ObjectConfig{
		Name: "mutation",
		Fields: graphql.Fields{
			"health": &graphql.Field{
				Type:        graphql.String,
				Description: "判断健康情况",
				Resolve:     getHealth,
			},
			"createFeedback": &graphql.Field{
				Args:        createFeedbackArgs,
				Type:        graphql.Boolean,
				Description: "创建反馈",
				Resolve:     createFeedback,
			},
			"createGroup": &graphql.Field{
				Args:        createGroupArgs,
				Type:        groupType,
				Description: "创建群组",
				Resolve:     createGroup,
			},
			"joinGroup": &graphql.Field{
				Args:        codeArgs,
				Type:        graphql.Boolean,
				Description: "加入群组",
				Resolve:     joinGroup,
			},
			"leaveGroup": &graphql.Field{
				Args:        codeArgs,
				Type:        graphql.Boolean,
				Description: "离开群组",
				Resolve:     leaveGroup,
			},
			"updateGroupMembers": &graphql.Field{
				Args:        updateGroupMembersArgs,
				Type:        graphql.Boolean,
				Description: "更新群组成员, 按照权限限制: 创建者 > 管理员 > 成员, 如：管理员可以删除成员",
				Resolve:     updateGroupMembers,
			},
			"createNotices": &graphql.Field{
				Args:        noticesArgs,
				Type:        graphql.Boolean,
				Description: "创建提醒",
				Resolve:     createNotices,
			},
			"updateNotice": &graphql.Field{
				Args:        noticeArgs,
				Type:        graphql.Boolean,
				Description: "更新提醒",
				Resolve:     updateNotice,
			},
			"deleteNotice": &graphql.Field{
				Args:        idArgs,
				Type:        graphql.Boolean,
				Description: "删除提醒",
				Resolve:     deleteNotice,
			},
		},
	})
)

func init() {
	schemaConfig := graphql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	}
	schema, _ := graphql.NewSchema(schemaConfig)

	graphiql := true
	if config.Conf.AppInfo.Env == "prod" {
		graphiql = false
	}

	handler = gh.New(&gh.Config{
		Schema:   &schema,
		GraphiQL: graphiql,
		Pretty:   graphiql,
	})
}

// Graphql Graphql handler
func Graphql(w http.ResponseWriter, r *http.Request) {
	// jwt
	token := r.Header.Get("Authorization")
	user, ok := validateJWT(token)
	if !ok {
		resJSONError(w, http.StatusUnauthorized, constant.ErrorMsgUnAuth)
		return
	}

	ctx := context.WithValue(context.Background(), constant.JWTContextKey, user)
	handler.ContextHandler(ctx, w, r)
}
