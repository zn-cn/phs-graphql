package controller

import (
	"constant"
	"controller/param"
	"model"
	"net/http"
)

/**
 * @apiDefine JoinGroupFromOfficialAccounts JoinGroupFromOfficialAccounts
 * @apiDescription 加入群组(不对前端开放)
 *
 * @apiParam {String} code 圈子code
 * @apiParam {String} id unionid
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "code": "圈子code"
 *       "id": "unionid"
 *     }
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *     }
 * @apiError {Number} status 状态码
 * @apiError {String} err_msg 错误信息
 *
 * @apiErrorExample Error-Response:
 *     HTTP/1.1 401 Unauthorized
 *     {
 *       "err_msg": "Unauthorized"
 *     }
 */
/**
 * @api {post} /api/unopen/group/action/join JoinGroupFromOfficialAccounts
 * @apiVersion 1.0.0
 * @apiName JoinGroupFromOfficialAccounts
 * @apiGroup UnOpen
 * @apiUse JoinGroupFromOfficialAccounts
 */
func JoinGroupFromOfficialAccounts(w http.ResponseWriter, r *http.Request) {
	data := param.CodeID{}
	err := loadJSONData(r, &data)
	if err != nil {
		writeUnOpenLog("JoinGroupFromOfficialAccounts", constant.ErrorMsgParamWrong, err)
		resJSONError(w, http.StatusBadRequest, constant.ErrorMsgParamWrong)
		return
	}

	if data.ID == "" || data.Code == "" {
		writeUnOpenLog("JoinGroupFromOfficialAccounts", constant.ErrorMsgParamWrong, err)
		resJSONError(w, http.StatusBadRequest, constant.ErrorMsgParamWrong)
		return
	}

	model.CreateUserByUnionid(data.ID)
	err = model.JoinGroup(data.Code, data.ID)
	if err != nil {
		writeUnOpenLog("JoinGroupFromOfficialAccounts", "加入群组失败", err)
		resJSONError(w, http.StatusBadGateway, "加入群组失败")
		return
	}

	// 发送模板消息
	go model.SendGroupJoinTemplate(data.ID, data.Code)

	resJSONData(w, nil)
}

func writeUnOpenLog(funcName, errMsg string, err error) {
	writeLog("unopen.go", funcName, errMsg, err)
}
