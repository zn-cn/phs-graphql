package controller

import (
	"constant"
	"controller/param"
	"model"
	"net/http"
	"util"
	"util/log"

	"github.com/sirupsen/logrus"
)

var (
	logger = log.GetLogger()
)

/**
 * @apiDefine Login Login
 * @apiDescription 登录，登录之后立即向后台发送一次请求更新用户信息
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "code": "code",
 *       "userInfo": {
 *           "nickName": String, // 用户昵称
 *           "gender": Number, // 性别 0：未知、1：男、2：女
 *           "province": String, // 省份
 *           "city": String, // 城市
 *           "country": String, // 国家
 *           "avatarUrl": String, // 用户头像
 *           "language": String, // 用户的语言，简体中文为zh_CN
 *         }
 *       "rawData": String, // 不包括敏感信息的原始数据字符串，用于计算签名
 *       "signature": String, // 使用 sha1( rawData + sessionkey ) 得到字符串，用于校验用户信息
 *       "encryptedData": String, // 包括敏感数据在内的完整用户信息的加密数据
 *       "iv": String, // 加密算法的初始向量
 *     }
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *        "jwt_token": "jwt_token",  // 有效时间为七天，发过来的时候需要在前面加上"Bearer "
 *     }
 *
 * @apiError {Number} status 状态码
 * @apiError {String} err_msg 错误信息
 *
 * @apiErrorExample Error-Response:
 *     HTTP/1.1 502 Bad Gateway
 *     {
 *       "err_msg": "Bad Gateway"
 *     }
 */
/**
 * @api {post} /api/v1/login Login
 * @apiVersion 1.0.0
 * @apiName Login
 * @apiGroup Index
 * @apiUse Login
 */

func Login(w http.ResponseWriter, r *http.Request) {
	data := param.WeixinLoginData{}
	err := loadJSONData(r, &data)
	if err != nil {
		writeRestLog("Login", constant.ErrorMsgParamWrong, err)
		resJSONError(w, http.StatusBadRequest, constant.ErrorMsgParamWrong)
		return
	}

	weixinSessRes, err := model.GetWeixinSession(data.Code)
	if err != nil {
		writeRestLog("Login", constant.ErrorMsgParamWrong, err)
		resJSONError(w, http.StatusBadRequest, constant.ErrorMsgParamWrong)
		return
	}

	var userInfo *util.DecryptUserInfo
	if weixinSessRes.Unionid == "" {
		userInfo, err = model.DecryptWeixinEncryptedData(weixinSessRes.SessionKey, data.EncryptedData, data.Iv)
		if err != nil {
			writeRestLog("Login", constant.ErrorMsgParamWrong, err)
			resJSONError(w, http.StatusBadRequest, constant.ErrorMsgParamWrong)
			return
		}
	} else {
		userInfo = &util.DecryptUserInfo{
			UnionID:   weixinSessRes.Unionid,
			OpenID:    weixinSessRes.Openid,
			NickName:  data.UserInfo.Nickname,
			Gender:    data.UserInfo.Gender,
			Province:  data.UserInfo.Province,
			City:      data.UserInfo.City,
			Country:   data.UserInfo.Country,
			AvatarURL: data.UserInfo.AvatarURL,
			Language:  data.UserInfo.Language,
		}
	}

	err = model.CreateUser(userInfo)
	if err != nil {
		writeRestLog("Login", constant.ErrorMsgUserCreate, err)
		resJSONError(w, http.StatusBadGateway, constant.ErrorMsgUserCreate)
		return
	}

	jwtAuth := map[string]interface{}{
		"user_id": userInfo.UnionID,
	}

	resData := map[string]interface{}{
		"jwt_token": getJWTToken(jwtAuth),
	}
	resJSONData(w, resData)
}

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
		writeRestLog("JoinGroupFromOfficialAccounts", constant.ErrorMsgParamWrong, err)
		resJSONError(w, http.StatusBadRequest, constant.ErrorMsgParamWrong)
		return
	}

	if data.ID == "" || data.Code == "" {
		writeRestLog("JoinGroupFromOfficialAccounts", constant.ErrorMsgParamWrong, err)
		resJSONError(w, http.StatusBadRequest, constant.ErrorMsgParamWrong)
		return
	}

	model.CreateUserByUnionid(data.ID)
	err = model.JoinGroup(data.Code, data.ID)
	if err != nil {
		writeRestLog("JoinGroupFromOfficialAccounts", "加入群组失败", err)
		resJSONError(w, http.StatusBadGateway, "加入群组失败")
		return
	}

	// 发送模板消息
	go model.SendGroupJoinTemplate(data.ID, data.Code)

	resJSONData(w, nil)
}

func writeRestLog(funcName, errMsg string, err error) {
	writeLog("rest.go", funcName, errMsg, err)
}

func writeLog(fileName, funcName, errMsg string, err error) {
	logger.WithFields(logrus.Fields{
		"package":  "controller",
		"file":     fileName,
		"function": funcName,
		"err":      err,
	}).Warn(errMsg)
}
