package controller

import (
	"constant"
	"model"

	"github.com/graphql-go/graphql"
)

var userArgs = graphql.FieldConfigArgument{
	"userID": &graphql.ArgumentConfig{
		Description: "用户id，即unionid",
		Type:        graphql.ID,
	},
}

var userStatusEnumType = graphql.NewEnum(graphql.EnumConfig{
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

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "user",
	Description: "user",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.ID,
			Description: "id",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if user, ok := p.Source.(model.User); ok {
					return user.ID.Hex(), nil
				}
				return nil, constant.ErrorEmpty
			},
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
		"isFollow": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "是否关注了服务号",
			Resolve:     getUserFollowStatus,
		},
	},
})

func init() {
	userType.AddFieldConfig("ownGroups", &graphql.Field{
		Type:        graphql.NewList(groupType),
		Description: "创建/拥有的群组",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if user, ok := p.Source.(model.User); ok {
				return model.GetRedisGroupInfos(user.OwnGroupIDs)
			}
			return nil, constant.ErrorEmpty
		},
	})
	userType.AddFieldConfig("manageGroups", &graphql.Field{
		Type:        graphql.NewList(groupType),
		Description: "管理的群组",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if user, ok := p.Source.(model.User); ok {
				return model.GetRedisGroupInfos(user.ManageGroupIDs)
			}
			return nil, constant.ErrorEmpty
		},
	})
	userType.AddFieldConfig("joinGroups", &graphql.Field{
		Type:        graphql.NewList(groupType),
		Description: "加入的群组",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if user, ok := p.Source.(model.User); ok {
				return model.GetRedisGroupInfos(user.JoinGroupIDs)
			}
			return nil, constant.ErrorEmpty
		},
	})
}

func getUserFollowStatus(p graphql.ResolveParams) (interface{}, error) {
	userID := getJWTUserID(p)
	isFollow, _ := model.IsFollowOfficeAccount(userID)

	go model.SetUserFollowStatus(userID, isFollow)

	return isFollow, nil
}

func getUserByUnionid(p graphql.ResolveParams) (interface{}, error) {
	userID, ok := p.Args["userID"].(string)
	if userID == "" || !ok {
		userID = getJWTUserID(p)
	}

	return model.GetUserByUnionid(userID)
}

func writeUserLog(funcName, errMsg string, err error) {
	writeLog("user.go", funcName, errMsg, err)
}
