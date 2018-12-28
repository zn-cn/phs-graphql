package controller

import (
	"constant"
	"controller/param"
	"fmt"
	"model"
	"util"

	"github.com/graphql-go/graphql"
)

var codeArgs = graphql.FieldConfigArgument{
	"code": &graphql.ArgumentConfig{
		Description: "圈子code",
		Type:        graphql.NewNonNull(graphql.String),
	},
}

var groupStatusEnumType = graphql.NewEnum(graphql.EnumConfig{
	Name:        "groupStatusEnum",
	Description: "圈子状态",
	Values: graphql.EnumValueConfigMap{
		"delete": &graphql.EnumValueConfig{
			Value:       constant.GroupDelStatus,
			Description: "解散状态",
		},
		"common": &graphql.EnumValueConfig{
			Value:       constant.GroupCommonStatus,
			Description: "正常状态",
		},
	},
})

var groupUserStatusEnumType = graphql.NewEnum(graphql.EnumConfig{
	Name:        "groupStatusEnum",
	Description: "圈子状态",
	Values: graphql.EnumValueConfigMap{
		"owner": &graphql.EnumValueConfig{
			Value:       constant.GroupUserStatusOwner,
			Description: "创建者状态",
		},
		"manager": &graphql.EnumValueConfig{
			Value:       constant.GroupUserStatusManager,
			Description: "管理员状态",
		},
		"member": &graphql.EnumValueConfig{
			Value:       constant.GroupUserStatusMember,
			Description: "成员状态",
		},
	},
})

var ticketType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ticket",
	Description: "二维码",
	Fields: graphql.Fields{
		"ticketUrl": &graphql.Field{
			Type:        graphql.String,
			Description: "二维码图片链接",
		},
		"url": &graphql.Field{
			Type:        graphql.String,
			Description: "二维码图片解析后的地址，开发者可根据该地址自行生成需要的二维码图片",
		},
	},
})

var groupType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "group",
	Description: "group",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.ID,
			Description: "id",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if group, ok := p.Source.(model.Group); ok == true {
					return group.ID.Hex(), nil
				}
				return nil, constant.ErrorEmpty
			},
		},
		"status": &graphql.Field{
			Type:        groupStatusEnumType,
			Description: "状态",
		},
		"userStatus": &graphql.Field{
			Type:        groupUserStatusEnumType,
			Description: "用户的状态",
			Resolve:     getGroupUserStatus,
		},
		"code": &graphql.Field{
			Type:        graphql.ID,
			Description: "圈子code -> 邀请码, unique",
		},
		"ticket": &graphql.Field{
			Type:        ticketType,
			Description: "二维码",
			Resolve:     getGroupQrcode,
		},
		"nickname": &graphql.Field{
			Type:        graphql.String,
			Description: "昵称",
		},
		"avatarUrl": &graphql.Field{
			Type:        graphql.String,
			Description: "用户头像",
		},
		"createTime": &graphql.Field{
			Type:        graphql.Int,
			Description: "创建时间毫秒时间戳",
		},
		"personNum": &graphql.Field{
			Type:        graphql.Int,
			Description: "总人数：1 + 管理员人数 + 成员人数",
		},
	},
})

func init() {
	groupType.AddFieldConfig("owner", &graphql.Field{
		Type:        userType,
		Description: "创建者信息",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if group, ok := p.Source.(model.Group); ok == true {
				return model.GetRedisUserInfo(group.OwnerID)
			}
			writeGroupLog("managers", "获取群组创建者信息失败", nil)
			return nil, constant.ErrorEmpty
		},
	})
	groupType.AddFieldConfig("managers", &graphql.Field{
		Type:        graphql.NewList(userType),
		Description: "管理员",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if group, ok := p.Source.(model.Group); ok == true {
				return model.GetRedisUserInfos(group.ManagerIDs)
			}
			writeGroupLog("managers", "获取群组管理员信息失败", nil)
			return nil, constant.ErrorEmpty
		},
	})
	groupType.AddFieldConfig("members", &graphql.Field{
		Type:        graphql.NewList(userType),
		Description: "成员",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if group, ok := p.Source.(model.Group); ok == true {
				return model.GetRedisUserInfos(group.MemberIDs)
			}
			writeGroupLog("managers", "获取群组成员信息失败", nil)
			return nil, constant.ErrorEmpty
		},
	})
}

func getGroupQrcode(p graphql.ResolveParams) (interface{}, error) {
	code, _ := p.Args["code"].(string)
	if code == "" {
		return nil, constant.ErrorParamWrong
	}
	res, err := model.CreateQrcodeByGroupCode(code)
	if err != nil {
		writeGroupLog("getGroupQrcode", "获取二维码失败", err)
		return nil, constant.ErrorBadGateway
	}
	resData := map[string]interface{}{
		"url":       res.URL,
		"ticketUrl": fmt.Sprintf(constant.URLQrcodeTicket, res.Ticket),
	}
	return resData, nil
}

func getGroupUserStatus(p graphql.ResolveParams) (interface{}, error) {
	if group, ok := p.Source.(model.Group); ok == true {
		status := 0
		userID := getJWTUserID(p)
		if userID == group.OwnerID {
			status = 1
		} else {
			for _, id := range group.ManagerIDs {
				if id == userID {
					status = 2
					break
				}
			}
			if status == 0 {
				status = 3
			}
		}
		return status, nil
	}
	return nil, constant.ErrorEmpty
}

func getGroupByCode(p graphql.ResolveParams) (interface{}, error) {
	if code, ok := p.Args["code"].(string); ok {
		return model.GetGroupByCode(code)
	}
	return nil, constant.ErrorParamWrong
}

var createGroupArgs = graphql.FieldConfigArgument{
	"nickname": &graphql.ArgumentConfig{
		Description: "昵称",
		Type:        graphql.NewNonNull(graphql.String),
	},
	"avatarUrl": &graphql.ArgumentConfig{
		Description:  "圈子头像",
		Type:         graphql.String,
		DefaultValue: constant.ImgDefaultGraoupHead,
	},
}

func createGroup(p graphql.ResolveParams) (interface{}, error) {
	nickname := p.Args["nickname"].(string)
	avatarURL := p.Args["avatarURL"].(string)
	userID := getJWTUserID(p)
	if isFollow, _ := model.IsFollowOfficeAccount(userID); !isFollow {
		model.SetUserFollowStatus(userID, isFollow)
		return nil, constant.ErrorUnFollow
	}
	code, err := model.CreateGroup(userID, nickname, avatarURL)
	if err != nil {
		writeGroupLog("CreateGroup", "创建群组失败", err)
		return nil, err
	}

	resData := map[string]string{
		"code": code,
	}
	return resData, nil
}

func joinGroup(p graphql.ResolveParams) (interface{}, error) {
	code := p.Args["code"].(string)
	userID := getJWTUserID(p)

	if isFollow, _ := model.IsFollowOfficeAccount(userID); !isFollow {
		model.SetUserFollowStatus(userID, isFollow)
		return false, constant.ErrorUnFollow
	}
	err := model.JoinGroup(code, userID)
	if err != nil {
		writeGroupLog("joinGroup", "加入群组失败", err)
		return false, err
	}

	return true, nil
}

func leaveGroup(p graphql.ResolveParams) (interface{}, error) {
	code := p.Args["code"].(string)
	userID := getJWTUserID(p)

	err := model.LeaveGroup(code, userID)
	if err != nil {
		writeGroupLog("leaveGroup", "离开群组失败", err)
		return false, err
	}

	return true, nil
}

var updateGroupMembersEnumType = graphql.NewEnum(graphql.EnumConfig{
	Name:        "updateGroupMembersEnum",
	Description: "更新类型",
	Values: graphql.EnumValueConfigMap{
		"UpdateOwner": &graphql.EnumValueConfig{
			Value:       constant.ReqGroupUpdateOwnerType,
			Description: "更新拥有者(转让群组，且只能转给管理员)",
		},
		"DelOwner": &graphql.EnumValueConfig{
			Value:       constant.ReqGroupDelOwnerType,
			Description: "删除拥有者，即解散群组",
		},
		"SetManager": &graphql.EnumValueConfig{
			Value:       constant.ReqGroupSetManagerType,
			Description: "设置管理者",
		},
		"UnSetManager": &graphql.EnumValueConfig{
			Value:       constant.ReqGroupUnSetManagerType,
			Description: "取消管理员权限",
		},
		"DelManager": &graphql.EnumValueConfig{
			Value:       constant.ReqGroupDelManagerType,
			Description: "删除管理员",
		},
		"DelMember": &graphql.EnumValueConfig{
			Value:       constant.ReqGroupDelMemberType,
			Description: "删除成员",
		},
	},
})

var updateGroupMembersArgs = graphql.FieldConfigArgument{
	"type": &graphql.ArgumentConfig{
		Description: "更新类型",
		Type:        graphql.NewNonNull(updateGroupMembersEnumType),
	},
	"groupID": &graphql.ArgumentConfig{
		Description: "群组id",
		Type:        graphql.String,
	},
	"userIDs": &graphql.ArgumentConfig{
		Description: " 用户unionid数组, type=1时 数组长度为1, type=2时 数组长度为0",
		Type:        graphql.NewList(graphql.String),
	},
}

func updateGroupMembers(p graphql.ResolveParams) (interface{}, error) {
	data := param.TypeGroupIDUserIDs{}
	err := util.MapToJSONStruct(p.Args, &data)
	if err != nil {
		writeGroupLog("updateGroupMembers", constant.ErrorMsgParamWrong, err)
		return false, err
	}

	update := map[int]func(string, string, []string) error{
		constant.ReqGroupUpdateOwnerType:  model.UpdateGroupOwner,
		constant.ReqGroupDelOwnerType:     model.DelGroupOwner,
		constant.ReqGroupSetManagerType:   model.SetGroupManager,
		constant.ReqGroupUnSetManagerType: model.UnSetGroupManager,
		constant.ReqGroupDelManagerType:   model.DelGroupManager,
		constant.ReqGroupDelMemberType:    model.DelGroupMember,
	}
	if f, ok := update[data.Type]; ok {
		userID := getJWTUserID(p)
		err = f(data.GroupID, userID, data.UserIDs)

	} else {
		err = constant.ErrorParamWrong
	}

	if err != nil {
		writeGroupLog("updateGroupMembers", "更新失败", err)
		return false, err
	}
	return true, nil
}

func writeGroupLog(funcName, errMsg string, err error) {
	writeLog("group.go", funcName, errMsg, err)
}
