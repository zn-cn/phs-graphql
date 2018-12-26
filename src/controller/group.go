package controller

/*
   圈子群体
*/
import (
	"constant"
	"controller/param"
	"fmt"
	"model"
	"net/http"

	"github.com/labstack/echo"
)

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
	userID := getJWTUserID(c)
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
 * @apiDefine CreateGroup CreateGroup
 * @apiDescription 创建群组
 *
 * @apiParam {String} avatar_url 圈子头像链接
 * @apiParam {String} nickname 群昵称
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "avatar_url": "圈子头像"，
 *       "nickname": "软件1601",
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
 *           "code": String, // 唯一群code，同时也是邀请码
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
 * @api {post} /api/v1/group CreateGroup
 * @apiVersion 1.0.0
 * @apiName CreateGroup
 * @apiGroup Group
 * @apiUse CreateGroup
 */
func CreateGroup(c echo.Context) error {
	data := param.AvatarNicknameParam{}
	err := c.Bind(&data)
	if err != nil {
		writeGroupLog("CreateGroup", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}

	if data.AvatarURL == "" {
		// 使用默认头像
		data.AvatarURL = constant.ImgDefaultGraoupHead
	}
	userID := getJWTUserID(c)
	if isFollow, _ := model.IsFollowOfficeAccount(userID); !isFollow {
		model.SetUserFollowStatus(userID, isFollow)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, "你还没有关注公众号")
	}
	code, err := model.CreateGroup(userID, data.Nickname, data.AvatarURL)
	if err != nil {
		writeGroupLog("CreateGroup", "创建群组失败", err)
		return retError(c, http.StatusBadGateway, http.StatusBadGateway, err.Error())
	}

	resData := map[string]string{
		"code": code,
	}
	return retData(c, resData)
}

/**
 * @apiDefine JoinGroup JoinGroup
 * @apiDescription 加入群组
 *
 * @apiParam {String} code 圈子code
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "code": "圈子code"
 *     }
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": 200,
 *       "data": ""
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
 * @api {post} /api/v1/group/action/join JoinGroup
 * @apiVersion 1.0.0
 * @apiName JoinGroup
 * @apiGroup Group
 * @apiUse JoinGroup
 */
func JoinGroup(c echo.Context) error {
	data := param.CodeParam{}
	err := c.Bind(&data)
	if err != nil {
		writeGroupLog("JoinGroup", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}

	userID := getJWTUserID(c)
	if isFollow, _ := model.IsFollowOfficeAccount(userID); !isFollow {
		model.SetUserFollowStatus(userID, isFollow)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, "你还没有关注公众号")
	}
	err = model.JoinGroup(data.Code, userID)
	if err != nil {
		writeGroupLog("JoinGroup", "加入群组失败", err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, "加入群组失败")
	}
	return retData(c, "")
}

/**
 * @apiDefine LeaveGroup LeaveGroup
 * @apiDescription 离开群组
 *
 * @apiParam {String} code 圈子code
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "code": "圈子code"
 *     }
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": 200,
 *       "data": ""
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
 * @api {post} /api/v1/group/action/leave LeaveGroup
 * @apiVersion 1.0.0
 * @apiName LeaveGroup
 * @apiGroup Group
 * @apiUse LeaveGroup
 */
func LeaveGroup(c echo.Context) error {
	data := param.CodeParam{}
	err := c.Bind(&data)
	if err != nil {
		writeGroupLog("LeaveGroup", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}

	userID := getJWTUserID(c)
	err = model.LeaveGroup(data.Code, userID)
	if err != nil {
		writeGroupLog("LeaveGroup", "离开群组失败", err)
		return retError(c, http.StatusBadGateway, http.StatusBadGateway, "离开群组失败")
	}
	return retData(c, "")
}

/**
 * @apiDefine UpdateGroupMembers UpdateGroupMembers
 * @apiDescription 更新群组成员, 按照权限限制: 创建者 > 管理员 > 成员, 如：管理员可以删除成员
 *
 * @apiParam {Number} type 类型：1->更新拥有者(转让群组，且只能转给管理员), 2->删除拥有者，即解散群组, 3->设置管理者, 4->取消管理员权限, 5->删除管理员，6->删除成员
 * @apiParam {String} group_id 群组id
 * @apiParam {Array} user_ids 用户id数组
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "type": 1，
 *       "group_id": "group_id"，// 群组id
 *       "user_ids": ["user_id"]，// 用户unionid数组, type=1时 数组长度为1, type=2时 数组长度为0
 *     }
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": 200,
 *       "data": ""
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
 * @api {put} /api/v1/group/member/list UpdateGroupMembers
 * @apiVersion 1.0.0
 * @apiName UpdateGroupMembers
 * @apiGroup Group
 * @apiUse UpdateGroupMembers
 */
func UpdateGroupMembers(c echo.Context) error {
	data := param.TypeGroupIDUserIDs{}
	err := c.Bind(&data)
	if err != nil {
		writeGroupLog("UpdateGroupMembers", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
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
		userID := getJWTUserID(c)
		err = f(data.GroupID, userID, data.UserIDs)
	} else {
		err = constant.ErrorParamWrong
	}

	if err != nil {
		writeGroupLog("UpdateGroupMembers", "更新失败", err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, "更新失败")
	}
	return retData(c, "")
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
