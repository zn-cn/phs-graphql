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
	graphiql bool
)

func init() {
	if config.Conf.AppInfo.Env == "prod" {
		graphiql = false
	} else {
		graphiql = true
	}

	schemaConfig := graphql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	}
	schema, _ := graphql.NewSchema(schemaConfig)
	handler = gh.New(&gh.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: graphiql,
	})
}

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
