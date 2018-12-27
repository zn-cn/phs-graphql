package controller

import (
	"constant"
	"model"

	"github.com/graphql-go/graphql"
)

var (
	userStatusEnumType = graphql.NewEnum(graphql.EnumConfig{
		Name:        "userStatusEnum",
		Description: "用户状态",
		Values: graphql.EnumValueConfigMap{
			"delete": &graphql.EnumValueConfig{
				Value:       constant.UserDeleteStatus,
				Description: "被删除",
			},
			"unFollow": &graphql.EnumValueConfig{
				Value:       constant.UserUnFollowStatus,
				Description: "没有关注公众号",
			},
			"follow": &graphql.EnumValueConfig{
				Value:       constant.UserFollowStatus,
				Description: "已经关注公众号的普通用户",
			},
		},
	})

	userType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "user",
		Description: "user",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.ID,
				Description: "id",
			},
			"status": &graphql.Field{
				Type:        userStatusEnumType,
				Description: "状态",
			},
			"openid": &graphql.Field{
				Type:        graphql.ID,
				Description: "openid",
			},
			"unionid": &graphql.Field{
				Type:        graphql.ID,
				Description: "unionid",
			},
			"nickname": &graphql.Field{
				Type:        graphql.String,
				Description: "昵称",
			},
			"avatarUrl": &graphql.Field{
				Type:        graphql.String,
				Description: "用户头像",
			},
			"gender": &graphql.Field{
				Type:        graphql.Int,
				Description: "性别 0：未知、1：男、2：女",
			},
			"province": &graphql.Field{
				Type:        graphql.String,
				Description: "省份",
			},
			"city": &graphql.Field{
				Type:        graphql.String,
				Description: "城市",
			},
			"country": &graphql.Field{
				Type:        graphql.String,
				Description: "国家",
			},
			"language": &graphql.Field{
				Type:        graphql.String,
				Description: "语言",
			},
			"ownGroups": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "创建/拥有的群组",
			},
			"manageGroups": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "管理的群组",
			},
			"joinGroups": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "加入的群组",
			},
			"isFollow": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "是否关注了服务号",
				Resolve:     getUserFollowStatus,
			},
		},
	})
)

func getUserFollowStatus(p graphql.ResolveParams) (interface{}, error) {
	userID := getJWTUserID(p)
	isFollow, _ := model.IsFollowOfficeAccount(userID)

	go model.SetUserFollowStatus(userID, isFollow)

	return isFollow, nil
}

func writeUserLog(funcName, errMsg string, err error) {
	writeLog("user.go", funcName, errMsg, err)
}
