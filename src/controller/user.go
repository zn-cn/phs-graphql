package controller

import (
	"model"

	"github.com/labstack/echo"
)

/**
 * @apiDefine GetUserFollowStatus GetUserFollowStatus
 * @apiDescription 获取是否关注了服务号
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": 200,
 *       "data": {
 *           "is_follow": Boolean, // 是否关注了服务号
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
 * @api {get} /api/v1/user/status/follow GetUserFollowStatus
 * @apiVersion 1.0.0
 * @apiName GetUserFollowStatus
 * @apiGroup User
 * @apiUse GetUserFollowStatus
 */
func GetUserFollowStatus(c echo.Context) error {
	userID := getJWTUserID(c)
	isFollow, _ := model.IsFollowOfficeAccount(userID)

	go model.SetUserFollowStatus(userID, isFollow)
	resData := map[string]bool{
		"is_follow": isFollow,
	}
	return retData(c, resData)
}

func writeUserLog(funcName, errMsg string, err error) {
	writeLog("user.go", funcName, errMsg, err)
}
