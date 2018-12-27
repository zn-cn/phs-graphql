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
	noticeStatusEnumType = graphql.NewEnum(graphql.EnumConfig{
		Name:        "noticeStatusEnum",
		Description: "通知状态",
		Values: graphql.EnumValueConfigMap{
			"delete": &graphql.EnumValueConfig{
				Value:       constant.NoticeDeleteStatus,
				Description: "删除状态",
			},
			"expire": &graphql.EnumValueConfig{
				Value:       constant.NoticeExpireStatus,
				Description: "过期状态",
			},
			"publish": &graphql.EnumValueConfig{
				Value:       constant.NoticePubStatus,
				Description: "发布状态",
			},
		},
	})

	noticeType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "notice",
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
				Type:        noticeStatusEnumType,
				Description: " 状态",
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

var updateNoticeEnumType = graphql.NewEnum(graphql.EnumConfig{
	Name:        "updateGroupMembersEnumTypeEnum",
	Description: "更新类型",
	Values: graphql.EnumValueConfigMap{
		"Update": &graphql.EnumValueConfig{
			Value:       constant.ReqNoticeUpdate,
			Description: "大更新(空值自动过滤，更新不包含groupID)",
		},
		"UpdateContent": &graphql.EnumValueConfig{
			Value:       constant.ReqNoticeUpdateContentType,
			Description: "更新content",
		},
		"UpdateImgs": &graphql.EnumValueConfig{
			Value:       constant.ReqNoticeUpdateImgsType,
			Description: "更新imgs",
		},
		"UpdateTitle": &graphql.EnumValueConfig{
			Value:       constant.ReqNoticeUpdateTitleType,
			Description: "更新标题",
		},
		"UpdateNote": &graphql.EnumValueConfig{
			Value:       constant.ReqNoticeUpdateNoteType,
			Description: "更新note",
		},
		"UpdateNoticeTime": &graphql.EnumValueConfig{
			Value:       constant.ReqNoticeUpdateNoticeTimeType,
			Description: "更新noticeTime",
		},
		"UpdateGroupID": &graphql.EnumValueConfig{
			Value:       constant.ReqNoticeUpdateGroupIDType,
			Description: "更新groupID",
		},
	},
})

var noticesArgs = graphql.FieldConfigArgument{
	"notices": &graphql.ArgumentConfig{
		Description: "提醒",
		Type:        graphql.NewList(noticeType),
	},
}

func createNotices(p graphql.ResolveParams) (interface{}, error) {
	data := param.NoticesParam{}
	err := util.MapToJSONStruct(p.Args, &data)
	if err != nil {
		writeNoticeLog("CreateNotices", constant.ErrorMsgParamWrong, err)
		return false, err
	}
	if len(data.Notices) == 0 {
		writeNoticeLog("CreateNotices", constant.ErrorMsgParamWrong, constant.ErrorEmpty)
		return false, err
	}

	userID := getJWTUserID(p)
	err = model.CreateNotices(userID, data.Notices)
	if err != nil {
		writeNoticeLog("CreateNotices", constant.ErrorMsgParamWrong, err)
		return false, err
	}
	return true, nil
}

var noticeArgs = graphql.FieldConfigArgument{
	"id": &graphql.ArgumentConfig{
		Type:        graphql.ID,
		Description: "id",
	},
	"type": &graphql.ArgumentConfig{
		Description: "更新类型",
		Type:        graphql.NewNonNull(updateNoticeEnumType),
	},
	"groupID": &graphql.ArgumentConfig{
		Description: "群组id",
		Type:        graphql.String,
	},
	"title": &graphql.ArgumentConfig{
		Description: "作业标题",
		Type:        graphql.String,
	},
	"content": &graphql.ArgumentConfig{
		Description: "作业文字内容",
		Type:        graphql.String,
	},
	"imgs": &graphql.ArgumentConfig{
		Description: "图片",
		Type:        graphql.NewList(imgType),
	},
	"note": &graphql.ArgumentConfig{
		Description: "注释",
		Type:        graphql.String,
	},
	"noticeTime": &graphql.ArgumentConfig{
		Type:        graphql.Int,
		Description: "提醒时间毫秒时间戳",
	},
}

func updateNotice(p graphql.ResolveParams) (interface{}, error) {
	data := model.Notice{}
	err := util.MapToJSONStruct(p.Args, &data)
	if err != nil {
		writeNoticeLog("updateNotice", constant.ErrorMsgParamWrong, err)
		return false, err
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
		writeNoticeLog("updateNotice", constant.ErrorMsgParamWrong, err)
		return false, err
	}

	userID := ""
	err = model.UpdateNotice(data.ID.Hex(), userID, updateData)
	if err != nil {
		writeNoticeLog("updateNotice", "更新通知失败", err)
		return false, err
	}

	// TODO 提醒时间和组变了，更新redis
	return true, nil
}

func deleteNotice(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)
	userID := getJWTUserID(p)
	updateData := map[string]interface{}{
		"status": constant.NoticeDeleteStatus,
	}

	if err := model.UpdateNotice(id, userID, updateData); err != nil {
		writeNoticeLog("DeleteNotice", "删除通知失败", err)
		return false, err
	}
	return true, nil
}

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

func writeNoticeLog(funcName, errMsg string, err error) {
	writeLog("notice.go", funcName, errMsg, err)
}
