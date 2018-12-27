package controller

import (
	"constant"
	"controller/param"
	"fmt"
	"model"
	"net/http"
	"util"

	"github.com/graphql-go/graphql"
	"github.com/labstack/echo"
)

var (
	groupStatusEnumType = graphql.NewEnum(graphql.EnumConfig{
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

	groupType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "group",
		Description: "group",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.ID,
				Description: "id",
			},
			"status": &graphql.Field{
				Type:        groupStatusEnumType,
				Description: "状态",
			},
			"code": &graphql.Field{
				Type:        graphql.ID,
				Description: "圈子code -> 邀请码, unique",
			},
			"nickname": &graphql.Field{
				Type:        graphql.String,
				Description: "昵称",
			},
			"avatarUrl": &graphql.Field{
				Type:        graphql.String,
				Description: "用户头像",
			},
			"ownerID": &graphql.Field{
				Type:        graphql.String,
				Description: "unionid 注：以下三种身份不会重复，如：members中不会有owner",
			},
			"createTime": &graphql.Field{
				Type:        graphql.Int,
				Description: "创建时间毫秒时间戳",
			},
			"managers": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "管理员",
			},
			"members": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "成员",
			},
			"personNum": &graphql.Field{
				Type:        graphql.Int,
				Description: "总人数：1 + 管理员人数 + 成员人数",
			},
		},
	})
)
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
	userID := ""
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

var codeArgs = graphql.FieldConfigArgument{
	"code": &graphql.ArgumentConfig{
		Description: "圈子code",
		Type:        graphql.NewNonNull(graphql.String),
	},
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
	Name:        "updateGroupMembersEnumTypeEnum",
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
		userID := ""
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

/**
 * @apiDefine GetGroups GetGroups
 * @apiDescription 获取群组，我的页面的群组
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": 200,
 *       "data": {
 *           "is_follow": Boolean, // 是否关注关注公众号
 *           "own_groups": [{
 *               "id": String,
 *               "avatar_url": String, // 圈子头像
 *               "nickname": "软件1601",
 *               "code": "唯一群code，同时也是邀请码",
 *             }]
 *           "manage_groups": [{
 *               "id": String,
 *               "avatar_url": String, // 圈子头像
 *               "nickname": "软件1601",
 *               "code": "唯一群code，同时也是邀请码",
 *             }]
 *           "join_groups": [{
 *               "id": String,
 *               "avatar_url": String, // 圈子头像
 *               "nickname": "软件1601",
 *               "code": "唯一群code，同时也是邀请码",
 *             }]
 *         }
 *     }
 * @apiError {Number} status 状态码
 * @apiError {String} err_msg 错误信息
 *
 * @apiErrorExample Error-Response:
 *     HTTP/1.1 401 Unauthorized
 *     {
 *       "status": 401,
 *       "err_msg": "Unauthorized"
 *     }
 */
/**
 * @api {get} /api/v1/group/list GetGroups
 * @apiVersion 1.0.0
 * @apiName GetGroups
 * @apiGroup Group
 * @apiUse GetGroups
 */
func GetGroups(c echo.Context) error {
	userID := ""
	ownGroups, manageGroups, joinGroups, err := findGroupInfosByUserID(userID)
	if err != nil {
		writeGroupLog("GetGroups", "查询群组信息错误", err)
		return retError(c, http.StatusBadGateway, http.StatusBadGateway, "查询群组信息错误")
	}
	status, _ := model.GetUserStatus(userID)
	resData := map[string]interface{}{
		"is_follow":     status == constant.UserFollowStatus,
		"own_groups":    ownGroups,
		"manage_groups": manageGroups,
		"join_groups":   joinGroups,
	}
	return retData(c, resData)
}

/**
 * @apiDefine GetGroupQrcode GetGroupQrcode
 * @apiDescription 生成群组邀请码，扫码之后直接会定位到服务号
 *
 * @apiParam {String} code 圈子code
 *
 * @apiParamExample  {query} Request-Example:
 *     {
 *       "code": "圈子code"，
 *     }
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": 200,
 *       "data": {
 *           "ticket_url": "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=TICKET", // 获取二维码ticket后，开发者可用ticket换取二维码图片
 *           "url": String, // 二维码图片解析后的地址，开发者可根据该地址自行生成需要的二维码图片
 *         }
 *     }
 * @apiError {Number} status 状态码
 * @apiError {String} err_msg 错误信息
 *
 * @apiErrorExample Error-Response:
 *     HTTP/1.1 401 Unauthorized
 *     {
 *       "status": 401,
 *       "err_msg": "Unauthorized"
 *     }
 */
/**
 * @api {get} /api/v1/group/qrcode GetGroupQrcode
 * @apiVersion 1.0.0
 * @apiName GetGroupQrcode
 * @apiGroup Group
 * @apiUse GetGroupQrcode
 */
func GetGroupQrcode(c echo.Context) error {
	code := c.QueryParam("code")
	res, err := model.CreateQrcodeByGroupCode(code)
	if err != nil {
		writeGroupLog("GetGroupQrcode", "获取二维码失败", err)
		return retError(c, http.StatusBadGateway, http.StatusBadGateway, "获取二维码失败")
	}
	resData := map[string]interface{}{
		"url":        res.URL,
		"ticket_url": fmt.Sprintf(constant.URLQrcodeTicket, res.Ticket),
	}
	return retData(c, resData)
}

func findGroupInfosByUserID(userID string) ([]map[string]string, []map[string]string, []map[string]string, error) {
	ownGroups, manageGroups, joinGroups, err := model.FindGroupsByUserID(userID)
	if err != nil {
		return nil, nil, nil, err
	}
	groups := append(append(append([]string{}, ownGroups...), manageGroups...), joinGroups...)
	groupInfos, err := model.GetRedisGroupInfos(groups)
	if err != nil {
		return nil, nil, nil, err
	}
	ownGroupsLen := len(ownGroups)
	manageGroupsLen := len(manageGroups)
	return groupInfos[:ownGroupsLen], groupInfos[ownGroupsLen : ownGroupsLen+manageGroupsLen], groupInfos[ownGroupsLen+manageGroupsLen:], nil
}

/**
 * @apiDefine GetGroupMembers GetGroupMembers
 * @apiDescription 获取群组成员列表
 *
 * @apiParam {String} id 群组id
 *
 * @apiParamExample  {query} Request-Example:
 *     {
 *       "id": "group_id"，// 群组id
 *     }
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": 200,
 *       "data": {
 *           "id": String,
 *           "avatar_url": String,
 *           "nickname": String,
 *           "code": "圈子code -> 邀请码, unique",
 *           "owner": {
 *               "user_id": String,
 *               "nickname": String, // 用户昵称
 *               "gender": Number, // 性别 0：未知、1：男、2：女
 *               "province": String, // 省份
 *               "city": String, // 城市
 *               "country": String, // 国家
 *               "avatar_url": String, // 用户头像
 *               "language": String, // 用户的语言，简体中文为zh_CN
 *             }
 *           "managers": [{
 *               "user_id": String,
 *               "nickname": String, // 用户昵称
 *               "gender": Number, // 性别 0：未知、1：男、2：女
 *               "province": String, // 省份
 *               "city": String, // 城市
 *               "country": String, // 国家
 *               "avatar_url": String, // 用户头像
 *               "language": String, // 用户的语言，简体中文为zh_CN
 *             }]
 *           "members": [{
 *               "user_id": String,
 *               "nickname": String, // 用户昵称
 *               "gender": Number, // 性别 0：未知、1：男、2：女
 *               "province": String, // 省份
 *               "city": String, // 城市
 *               "country": String, // 国家
 *               "avatar_url": String, // 用户头像
 *               "language": String, // 用户的语言，简体中文为zh_CN
 *             }],
 *           "person_num": Number, // 总人数
 *         }
 *     }
 * @apiError {Number} status 状态码
 * @apiError {String} err_msg 错误信息
 *
 * @apiErrorExample Error-Response:
 *     HTTP/1.1 401 Unauthorized
 *     {
 *       "status": 401,
 *       "err_msg": "Unauthorized"
 *     }
 */
/**
 * @api {get} /api/v1/group/member/list GetGroupMembers
 * @apiVersion 1.0.0
 * @apiName GetGroupMembers
 * @apiGroup Group
 * @apiUse GetGroupMembers
 */
func GetGroupMembers(c echo.Context) error {
	data := param.IDParam{}
	err := c.Bind(&data)
	if err != nil {
		writeGroupLog("GetGroupMembers", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}

	group, err := model.GetGroup(data.ID)
	if err != nil {
		writeGroupLog("GetGroupMembers", "获取群组成员列表失败", err)
		return retError(c, http.StatusBadGateway, http.StatusBadGateway, "获取群组成员列表失败")
	}

	unionids := append(append([]string{group.OwnerID}, group.Managers...), group.Members...)
	userInfos, err := model.GetRedisUserInfos(unionids)
	if err != nil {
		writeGroupLog("GetGroupMembers", "获取群组成员信息失败", err)
		return retError(c, http.StatusBadGateway, http.StatusBadGateway, "获取群组成员信息失败")
	}

	managersLen := len(group.Managers)
	resData := map[string]interface{}{
		"id":         group.ID.Hex(),
		"nickname":   group.Nickname,
		"avatar_url": group.AvatarURL,
		"code":       group.Code,
		"owner":      userInfos[0],
		"managers":   userInfos[1 : 1+managersLen],
		"members":    userInfos[1+managersLen:],
	}
	return retData(c, resData)
}

func writeGroupLog(funcName, errMsg string, err error) {
	writeLog("group.go", funcName, errMsg, err)
}
