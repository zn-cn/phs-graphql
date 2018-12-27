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
	handler  *gh.Handler
	query    *graphql.Object
	mutation *graphql.Object
)

func init() {
	query = graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
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
		Name: "Mutation",
		Fields: graphql.Fields{
			"createGroup": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "创建群组",
			},
			"joinGroup": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "加入群组",
			},
			"leaveGroup": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "离开群组",
			},
		},
	})

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
		Pretty:   true,
		GraphiQL: graphiql,
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

	ctx := context.WithValue(context.Background(), "user", user)
	handler.ContextHandler(ctx, w, r)
}
