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
	query   = graphql.NewObject(graphql.ObjectConfig{
		Name: "query",
		Fields: graphql.Fields{
			"qiniuToken": &graphql.Field{
				Type:        qiniuTokenType,
				Description: "获取上传图片的七牛云upload-token 链接：https://developer.qiniu.com/kodo/manual/1208/upload-token",
				Args:        qiniuTokenArgs,
				Resolve:     getQiniuToken,
			},
			"user": &graphql.Field{
				Type:        userType,
				Description: "获取用户信息",
			},
			"notice": &graphql.Field{
				Type:        noticeType,
				Description: "获取通知信息",
			},
			"group": &graphql.Field{
				Type:        groupType,
				Description: "获取圈子信息",
			},
			"template": &graphql.Field{
				Type:        templateType,
				Description: "获取模板信息",
			},
		},
	})
	mutation = graphql.NewObject(graphql.ObjectConfig{
		Name: "mutation",
		Fields: graphql.Fields{
			"createFeedback": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "创建反馈",
				Args:        createFeedbackArgs,
				Resolve:     createFeedback,
			},
			"createGroup": &graphql.Field{
				Type:        groupType,
				Description: "创建群组",
				Args:        createGroupArgs,
				Resolve:     createGroup,
			},
			"joinGroup": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "加入群组",
				Args:        codeArgs,
				Resolve:     joinGroup,
			},
			"leaveGroup": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "离开群组",
				Args:        codeArgs,
				Resolve:     leaveGroup,
			},
			"updateGroupMembers": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "更新群组成员, 按照权限限制: 创建者 > 管理员 > 成员, 如：管理员可以删除成员",
				Args:        updateGroupMembersArgs,
				Resolve:     updateGroupMembers,
			},
			"createNotices": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "创建提醒",
				Args:        noticesArgs,
				Resolve:     createNotices,
			},
			"updateNotice": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "更新提醒",
				Args:        noticeArgs,
				Resolve:     updateNotice,
			},
			"deleteNotice": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "删除提醒, 参数只需要id",
				Args:        noticeArgs,
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
