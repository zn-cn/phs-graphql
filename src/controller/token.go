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

var qiniuTokenTypeEnumType = graphql.NewEnum(graphql.EnumConfig{
	Name:        "qiniuTokenTypeEnum",
	Description: "七牛Token类型",
	Values: graphql.EnumValueConfigMap{
		"homework": &graphql.EnumValueConfig{
			Value:       constant.ImgTypeHomework,
			Description: "作业图片",
		},
		"head": &graphql.EnumValueConfig{
			Value:       constant.ImgTypeHead,
			Description: "头像",
		},
		"feedback": &graphql.EnumValueConfig{
			Value:       constant.ImgTypeFeedback,
			Description: "反馈图片",
		},
	},
})

var qiniuTokenArgs = graphql.FieldConfigArgument{
	"type": &graphql.ArgumentConfig{
		Description: "类型",
		Type:        graphql.NewNonNull(qiniuTokenTypeEnumType),
	},
	"suffix": &graphql.ArgumentConfig{
		Description:  "后缀，如：.jpg",
		Type:         graphql.String,
		DefaultValue: constant.ImgSuffix,
	},
}

var imgType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "img",
	Description: "img",
	Fields: graphql.Fields{
		"url": &graphql.Field{
			Type:        graphql.String,
			Description: "url",
		},
		"microUrl": &graphql.Field{
			Type:        graphql.String,
			Description: "microUrl",
		},
	},
})

var imgArgsType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name:        "imgArgs",
	Description: "imgArgs",
	Fields: graphql.InputObjectConfigFieldMap{
		"url": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "url",
		},
		"microUrl": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "microUrl",
		},
	},
})

var qiniuTokenType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "qiniuToken",
	Description: "qiniuToken",
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

func getQiniuToken(p graphql.ResolveParams) (interface{}, error) {
	tokenType := p.Args["type"].(int)
	suffix := p.Args["suffix"].(string)

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
