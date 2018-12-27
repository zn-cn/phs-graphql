package controller

import (
	"config"
	"constant"
	"model"
	"strings"
	"util/token"

	"github.com/graphql-go/graphql"
	"github.com/qiniu/api.v7/storage"
	uuid "github.com/satori/go.uuid"
)

var (
	qiniuTokenType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "QiniuToken",
		Description: "QiniuToken",
		Fields: graphql.Fields{
			"uploadToken": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "uploadToken",
			},
			"key": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "key",
			},
			"img": &graphql.Field{
				Type:        graphql.NewNonNull(imgType),
				Description: "img",
			},
		},
	})

	qiniuTokenArgs = graphql.FieldConfigArgument{
		"type": &graphql.ArgumentConfig{
			Description: "类型：1 -> 作业图片, 2 -> 圈子头像, 3 -> 反馈图片",
			Type:        graphql.NewNonNull(graphql.Int),
		},
		"suffix": &graphql.ArgumentConfig{
			Description: "后缀，如：.jpg",
			Type:        graphql.String,
		},
	}

	imgType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Img",
		Description: "Img",
		Fields: graphql.Fields{
			"url": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "url",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
			"microUrl": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "microUrl",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
		},
	})
)

func getQiniuToken(p graphql.ResolveParams) (interface{}, error) {
	tokenType := p.Args["type"].(int)
	suffix := p.Args["suffix"].(string)
	if suffix == "" {
		suffix = constant.ImgSuffix
	}

	imgID := uuid.NewV4().String()

	imgPrefix, ok := constant.ImgPrefix[tokenType]
	if !ok {
		writeTokenLog("GetQiniuImgUpToken", constant.ErrorMsgParamWrong, nil)
		return nil, constant.ErrorParamWrong
	}
	microImgPrefix := constant.ImgPrefixMicro[tokenType]
	keyToOverwrite := imgPrefix + imgID + suffix
	saveAsKey := microImgPrefix + imgID + suffix

	fop := constant.ImgOps + "|saveas/" + storage.EncodedEntry(config.Conf.Qiniu.Bucket, saveAsKey)
	persistentOps := strings.Join([]string{fop}, ";")
	upToken := token.GetCustomUpToken(keyToOverwrite, persistentOps, constant.TokenQiniuExpire)

	img := model.Img{
		URL:      constant.ImgURIPrefix + keyToOverwrite,
		MicroURL: constant.ImgURIPrefix + saveAsKey,
	}

	resData := map[string]interface{}{
		"uploadToken": upToken,
		"key":         keyToOverwrite,
		"img":         img,
	}
	return resData, nil
}

func writeTokenLog(funcName, errMsg string, err error) {
	writeLog("token.go", funcName, errMsg, err)
}
