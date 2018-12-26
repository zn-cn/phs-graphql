package controller

import (
	"config"
	"context"
	"net/http"

	"github.com/graphql-go/graphql"
	gh "github.com/graphql-go/handler"
)

var (
	schema       graphql.Schema
	schemaConfig graphql.SchemaConfig
	queryType    *graphql.Object
	mutationType *graphql.Object
	graphiql     bool
)

func init() {
	if config.Conf.AppInfo.Env == "prod" {
		graphiql = false
	} else {
		graphiql = true
	}
	schema, _ = graphql.NewSchema(schemaConfig)
}

func Graphql(w http.ResponseWriter, r *http.Request) {
	// jwt
	token := r.Header.Get("Authorization")
	user, ok := validateJWT(token)
	if !ok {

	}

	h := gh.New(&gh.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: graphiql,
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, "user", user)
	h.ContextHandler(ctx, w, r)
}
