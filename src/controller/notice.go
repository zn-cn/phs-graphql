package controller

import (
	"constant"
	"controller/param"
	"model"
	"net/http"
	"util"

	"github.com/graphql-go/graphql"

	"github.com/labstack/echo"
)

var (
	noticeType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Notice",
		Description: "通知",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.ID,
				Description: "id",
			},
			"type": &graphql.Field{
				Type:        graphql.Int,
				Description: "类型：1表示前一天发送通知，2表示前两天发送通知",
			},
			"status": &graphql.Field{
				Type:        graphql.Int,
				Description: "状态: -10 表示解散状态, 5 表示正常状态",
			},
			"creatorID": &graphql.Field{
				Type:        graphql.ID,
				Description: "创建者 unionid",
			},
			"groupID": &graphql.Field{
				Type:        graphql.ID,
				Description: "群组 _id",
			},
			"title": &graphql.Field{
				Type:        graphql.String,
				Description: "标题",
			},
			"content": &graphql.Field{
				Type:        graphql.String,
				Description: "内容",
			},
			"imgs": &graphql.Field{
				Type:        graphql.NewList(imgType),
				Description: "图片",
			},
			"note": &graphql.Field{
				Type:        graphql.String,
				Description: "备注",
			},
			"createTime": &graphql.Field{
				Type:        graphql.Int,
				Description: "创建时间毫秒时间戳",
			},
			"noticeTime": &graphql.Field{
				Type:        graphql.Int,
				Description: "提醒时间毫秒时间戳",
			},
			"watchUsers": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "查看用户",
			},
			"watchNum": &graphql.Field{
				Type:        graphql.Int,
				Description: "查看人数",
			},
			"likeUsers": &graphql.Field{
				Type:        graphql.NewList(graphql.String),
				Description: "点赞用户",
			},
			"likeNum": &graphql.Field{
				Type:        graphql.Int,
				Description: "点赞人数",
			},
		},
	})
)

/**
 * @apiDefine GetNotices GetNotices
 * @apiDescription 获取通知列表
 *
 * @apiParam {Number} type 类型：1->获取全部, 2->按照圈子获取提醒
 * @apiParam {Number} page 页数, 从1开始
 * @apiParam {Number} per_page 一页数量，限制范围: 1~20
 * @apiParam {String} id 圈子id
 *
 * @apiParamExample  {query} Request-Example:
 *     {
 *       "type": 1,
 *       "page": 1,
 *       "per_page": 10,
 *     }
 *
 * @apiParamExample  {query} Request-Example:
 *     {
 *       "type": 2,
 *       "id": String, // 圈子id
 *       "page": 1,
 *       "per_page": 10,
 *     }
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": 200,
 *       "data": [{
 *           "id": String,
 *           "type": Number, // 类型：
 *           "status": Number, // 状态：

 *           "creator_id": String,
 *           "group_id": String,
 *           "group_avatar_url": String, // 圈子头像
 *           "group_nickname": String, // 圈子头像
 *           "title": "微积分作业",
 *           "content": "作业文字内容",
 *           "imgs": [{
 *               "url": String,
 *               "micro_url": String,
 *             }],
 *           "note": "注释",
 *           "create_time": Number, // 创建时间的毫秒时间戳
 *           "notice_time": Number, // 提醒时间的毫秒时间戳
 *           "watch_num": Number, // 查看人数
 *           "like_num": Number, // 点赞人数
 *         }]
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
 * @api {get} /api/v1/notice/list GetNotices
 * @apiVersion 1.0.0
 * @apiName GetNotices
 * @apiGroup Notice
 * @apiUse GetNotices
 */
func GetNotices(c echo.Context) error {
	data := param.TypePageID{}
	err := c.Bind(&data)
	validateErr := c.Validate(&data)
	if err != nil || validateErr != nil {
		writeNoticeLog("GetNotices", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}
	userID := ""
	var groups []string
	if data.Type == constant.ReqNoticeGetAllType {
		ownGroups, manageGroups, joinGroups, err := model.FindGroupsByUserID(userID)
		if err != nil {
			writeGroupLog("GetNotices", "查询用户群组", err)
			return retError(c, http.StatusBadRequest, http.StatusBadRequest, "查询用户群组")
		}
		groups = append(append(ownGroups, manageGroups...), joinGroups...)
	} else {
		groups = []string{data.ID}
	}

	notices, _ := model.GetNotices(groups, data.Page, data.PerPage)
	resData := make([]interface{}, len(notices))
	groupIDs := make([]string, len(notices))
	for i, notice := range notices {
		groupIDs[i] = notice.GroupID
	}

	groupInfos, _ := model.GetRedisGroupInfos(groupIDs)
	for i, notice := range notices {
		tempData := util.JSONStructToMap(notice)
		tempData["group_avatar_url"] = groupInfos[i]["avatar_url"]
		tempData["group_nickname"] = groupInfos[i]["nickname"]
		resData[i] = tempData
	}

	return retData(c, resData)
}

/**
 * @apiDefine CreateNotices CreateNotices
 * @apiDescription 创建提醒
 *
 * @apiParam {Array} notices 提醒数组
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "notices": [{
 *           "group_id": String,
 *           "title": "微积分作业", // 前端要控制必须有标题
 *           "content": "作业文字内容",
 *           "imgs": [{
 *               "url": String,
 *               "micro_url": String,
 *             }],
 *           "note": "注释",
 *           "notice_time": Number, // 提醒时间戳（毫秒）
 *         }],
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
 * @api {post} /api/v1/notice/list CreateNotices
 * @apiVersion 1.0.0
 * @apiName CreateNotices
 * @apiGroup Notice
 * @apiUse CreateNotices
 */
func CreateNotices(c echo.Context) error {
	data := param.NoticesParam{}
	err := c.Bind(&data)
	if err != nil {
		writeNoticeLog("CreateNotices", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}
	if len(data.Notices) == 0 {
		writeNoticeLog("CreateNotices", constant.ErrorMsgParamWrong, constant.ErrorEmpty)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}

	userID := ""
	err = model.CreateNotices(userID, data.Notices)
	if err != nil {
		writeNoticeLog("CreateNotices", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}
	return retData(c, "")
}

/**
 * @apiDefine UpdateNotice UpdateNotice
 * @apiDescription 更新提醒
 *
 * @apiParam {Number} type 类型：0 -> 大更新(空值自动过滤，更新不包含group_id), 1->更新content, 2->更新imgs, 3->更新title, 4->更新note, 5->更新notice_time，6->更新group_id
 * @apiParam {String} id 作业id
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "type": 0，
 *       "id": "作业id",
 *       "content": "作业文字内容",
 *       "title": "微积分作业",
 *       "note": "注释",
 *       "imgs": [{
 *           "url": String,
 *           "micro_url": String,
 *         }],
 *     }
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "type": 1，
 *       "id": "作业id",
 *       "content": "作业文字内容",
 *     }
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "type": 2，
 *       "id": "作业id",
 *       "imgs": [{
 *           "url": String,
 *           "micro_url": String,
 *         }],
 *     }
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "type": 3，
 *       "id": "作业id",
 *       "title": "微积分作业",
 *     }
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "type": 4，
 *       "id": "作业id",
 *       "note": "注释",
 *     }
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "type": 5，
 *       "id": "作业id",
 *       "notice_time": Number, // 提醒时间戳（毫秒）
 *     }
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "type": 6，
 *       "id": "作业id",
 *       "group_id": String,
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
 * @api {put} /api/v1/notice UpdateNotice
 * @apiVersion 1.0.0
 * @apiName UpdateNotice
 * @apiGroup Notice
 * @apiUse UpdateNotice
 */
func UpdateNotice(c echo.Context) error {
	data := model.Notice{}
	err := c.Bind(&data)
	if err != nil {
		writeNoticeLog("UpdateNotice", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}

	updateData := map[string]interface{}{}
	switch data.Type {
	case constant.ReqNoticeUpdate:
		if data.Content != "" {
			updateData["content"] = data.Content
		}
		if data.Title != "" {
			updateData["title"] = data.Title
		}
		if data.Note != "" {
			updateData["note"] = data.Note
		}
		if data.NoticeTime > util.GetNowTimestamp() {
			updateData["notice_time"] = data.NoticeTime
		}
		if len(data.Imgs) > 0 {
			updateData["imgs"] = data.Imgs
		}
	case constant.ReqNoticeUpdateContentType:
		if data.Content == "" {
			err = constant.ErrorParamWrong
		}
		updateData["content"] = data.Content
	case constant.ReqNoticeUpdateImgsType:
		updateData["imgs"] = data.Imgs
	case constant.ReqNoticeUpdateTitleType:
		if data.Title == "" {
			err = constant.ErrorParamWrong
		}
		updateData["title"] = data.Title
	case constant.ReqNoticeUpdateNoteType:
		if data.Note == "" {
			err = constant.ErrorParamWrong
		}
		updateData["note"] = data.Note
	case constant.ReqNoticeUpdateNoticeTimeType:
		if data.NoticeTime <= util.GetNowTimestamp() {
			err = constant.ErrorParamWrong
		}
		updateData["notice_time"] = data.NoticeTime
	case constant.ReqNoticeUpdateGroupIDType:
		if data.GroupID == "" {
			err = constant.ErrorParamWrong
		}
		updateData["group_id"] = data.GroupID
	default:
		err = constant.ErrorParamWrong
	}

	if err != nil {
		writeNoticeLog("UpdateNotice", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}

	userID := ""
	err = model.UpdateNotice(data.ID.Hex(), userID, updateData)
	if err != nil {
		writeNoticeLog("UpdateNotice", "更新通知失败", err)
		return retError(c, http.StatusBadGateway, http.StatusBadGateway, "更新通知失败")
	}

	// TODO 提醒时间和组变了，更新redis
	return retData(c, "")
}

/**
 * @apiDefine DeleteNotice DeleteNotice
 * @apiDescription 删除提醒
 *
 * @apiParamExample  {query} Request-Example:
 *     {
 *       "id": String, // 作业id
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
 * @api {delete} /api/v1/notice DeleteNotice
 * @apiVersion 1.0.0
 * @apiName DeleteNotice
 * @apiGroup Notice
 * @apiUse DeleteNotice
 */
func DeleteNotice(c echo.Context) error {
	data := param.IDParam{}
	err := c.Bind(&data)
	if err != nil {
		writeNoticeLog("DeleteNotice", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}

	userID := ""
	updateData := map[string]interface{}{
		"status": constant.NoticeDeleteStatus,
	}
	err = model.UpdateNotice(data.ID, userID, updateData)
	if err != nil {
		writeNoticeLog("DeleteNotice", "删除通知失败", err)
		return retError(c, http.StatusBadGateway, http.StatusBadGateway, "删除通知失败")
	}
	return retData(c, "")
}

func writeNoticeLog(funcName, errMsg string, err error) {
	writeLog("notice.go", funcName, errMsg, err)
}
