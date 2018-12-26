package model

import (
	"config"
	"constant"
	"errors"
	"fmt"
	"model/db"
	"util"

	"github.com/imroc/req"
)

type WeixinTokenRes struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
}

type WeixinSessRes struct {
	Unionid    string `json:"unionid"`
	Openid     string `json:"openid"`
	SessionKey string `json:"session_key"`
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
}

type QrcodeParam struct {
	ExpireSeconds int              `json:"expire_seconds"`
	ActionName    string           `json:"action_name"`
	ActionInfo    QrcodeActionInfo `json:"action_info"`
}

type QrcodeActionInfo struct {
	Scene QrcodeScene `json:"scene"`
}
type QrcodeScene struct {
	SceneStr string `json:"scene_str"`
	SceneId  int32  `json:"scene_id"`
}

type QrcodeRes struct {
	ExpireSeconds int    `json:"expire_seconds,omitempty"`
	Ticket        string `json:"ticket"`
	URL           string `json:"url"`
}

func GetWeixinSession(code string) (WeixinSessRes, error) {
	data := WeixinSessRes{}
	appInfo := config.Conf.Wechat
	param := req.Param{
		"appid":      appInfo.AppID,
		"secret":     appInfo.AppSecret,
		"js_code":    code,
		"grant_type": "authorization_code",
	}
	url := constant.WechatSessionURIPrefix
	err := util.BindGetJSONData(url, param, &data)
	return data, err
}

func DecryptWeixinEncryptedData(sessionKey, encryptedData, iv string) (*util.DecryptUserInfo, error) {
	pc := util.NewWXBizDataCrypt(config.Conf.Wechat.AppID, sessionKey)
	return pc.Decrypt(encryptedData, iv)
}

func getAccessToken() (WeixinTokenRes, error) {
	data := WeixinTokenRes{}
	appInfo := config.Conf.Wechat
	param := req.Param{
		"appid":      appInfo.AppID,
		"secret":     appInfo.AppSecret,
		"grant_type": "client_credential",
	}
	url := constant.WechatTokenURIPrefix
	err := util.BindGetJSONData(url, param, &data)
	return data, err
}

// CreateQrcodeByGroupCode 创建二维码
func CreateQrcodeByGroupCode(code string) (QrcodeRes, error) {
	str := fmt.Sprintf(constant.WechatScanCodeJoinPhsMPGroup, code)
	reqData := QrcodeParam{
		ExpireSeconds: 3600 * 24 * 10,
		ActionName:    "QR_STR_SCENE",
		ActionInfo: QrcodeActionInfo{
			Scene: QrcodeScene{
				SceneStr: str,
			},
		},
	}
	resData := QrcodeRes{}

	r, err := req.Post(constant.URLCreateQrcode, req.BodyJSON(&reqData))
	if err != nil {
		return resData, err
	}
	err = r.ToJSON(&resData)
	return resData, err
}

/****************************************** weixin redis action ****************************************/

func UpdateRedisAccessToken() error {
	data, err := getAccessToken()
	if err != nil || data.Errcode != 0 {
		return errors.New(data.Errmsg)
	}
	return updateRedisAccessToken(data.AccessToken, data.ExpiresIn)
}

func updateRedisAccessToken(accessToken string, expire int64) error {
	cntlr := db.NewRedisDBCntlr()
	defer cntlr.Close()

	key := constant.RedisWeixinAccessToken
	_, err := cntlr.SETEX(key, expire, accessToken)
	return err
}

func getRedisAccessToken() (string, error) {
	cntlr := db.NewRedisDBCntlr()
	defer cntlr.Close()

	key := constant.RedisWeixinAccessToken
	return cntlr.GET(key)
}
