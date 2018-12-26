package controller

import (
	"config"
	"constant"
	"controller/param"
	"model"
	"net/http"
	"strings"
	"util"
	"util/log"
	"util/token"

	"github.com/graphql-go/graphql"
	"github.com/qiniu/api.v7/storage"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

var (
	logger         = log.GetLogger()
	qiniuTokenType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "QiniuToken",
		Description: "QiniuToken",
		Fields: graphql.Fields{
			"uploadToken": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "uploadToken",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
			"key": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "key",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
			"img": &graphql.Field{
				Type:        graphql.NewNonNull(imgType),
				Description: "img",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
		},
	})

	qiniuTokenArgs = graphql.FieldConfigArgument{
		"type": &graphql.ArgumentConfig{
			Description: "类型：1 -> 作业图片, 2 -> 圈子头像, 3 -> 反馈图片",
			Type:        graphql.NewNonNull(graphql.Int),
		},
		"suffix": &graphql.ArgumentConfig{
			Description: "后缀，如：.jpg",
			Type:        graphql.String,
		},
	}

	imgType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Img",
		Description: "Img",
		Fields: graphql.Fields{
			"url": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "url",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
			"microUrl": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "microUrl",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
		},
	})
)

func getQiniuToken(p graphql.ResolveParams) (interface{}, error) {
	tokenType := p.Args["type"].(int)
	suffix := p.Args["suffix"].(string)
	if suffix == "" {
		suffix = constant.ImgSuffix
	}

	imgID := uuid.NewV4().String()

	imgPrefix, ok := constant.ImgPrefix[tokenType]
	if !ok {
		writeIndexLog("GetQiniuImgUpToken", constant.ErrorMsgParamWrong, nil)
		return nil, constant.ErrorParamWrong
	}
	microImgPrefix := constant.ImgPrefixMicro[tokenType]
	keyToOverwrite := imgPrefix + imgID + suffix
	saveAsKey := microImgPrefix + imgID + suffix

	fop := constant.ImgOps + "|saveas/" + storage.EncodedEntry(config.Conf.Qiniu.Bucket, saveAsKey)
	persistentOps := strings.Join([]string{fop}, ";")
	upToken := token.GetCustomUpToken(keyToOverwrite, persistentOps, constant.TokenQiniuExpire)

	img := model.Img{
		URL:      constant.ImgURIPrefix + keyToOverwrite,
		MicroURL: constant.ImgURIPrefix + saveAsKey,
	}

	resData := map[string]interface{}{
		"uploadToken": upToken,
		"key":         keyToOverwrite,
		"img":         img,
	}
	return resData, nil
}

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
		writeIndexLog("Login", constant.ErrorMsgParamWrong, err)
		resJSONError(w, http.StatusBadRequest, constant.ErrorMsgParamWrong)
		return
	}

	weixinSessRes, err := model.GetWeixinSession(data.Code)
	if err != nil {
		writeIndexLog("Login", constant.ErrorMsgParamWrong, err)
		resJSONError(w, http.StatusBadRequest, constant.ErrorMsgParamWrong)
		return
	}

	var userInfo *util.DecryptUserInfo
	if weixinSessRes.Unionid == "" {
		userInfo, err = model.DecryptWeixinEncryptedData(weixinSessRes.SessionKey, data.EncryptedData, data.Iv)
		if err != nil {
			writeIndexLog("Login", constant.ErrorMsgParamWrong, err)
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
		writeIndexLog("Login", constant.ErrorMsgUserCreate, err)
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

func writeIndexLog(funcName, errMsg string, err error) {
	writeLog("index.go", funcName, errMsg, err)
}

func writeLog(fileName, funcName, errMsg string, err error) {
	logger.WithFields(logrus.Fields{
		"package":  "controller",
		"file":     fileName,
		"function": funcName,
		"err":      err,
	}).Warn(errMsg)
}
