package param

import "model"

/*---------------------------------- 其他 ---------------------------------------*/
type WeixinLoginData struct {
	CodeParam
	UserInfo      userInfo `json:"userInfo"`
	RawData       string   `json:"rawData"`
	Signature     string   `json:"signature"`
	EncryptedData string   `json:"encryptedData"`
	Iv            string   `json:"iv"`
}

type userInfo struct {
	Nickname  string `json:"nickName"`  // 用户昵称
	Gender    int    `json:"gender"`    // 1代表男性，0代表女性
	Province  string `json:"province"`  // 省份
	City      string `json:"city"`      // 城市
	Country   string `json:"country"`   // 国家
	AvatarURL string `json:"avatarUrl"` // 用户头像
	Language  string `json:"language"`  // 语言
}

type NoticesParam struct {
	Notices []model.Notice `json:"notices"`
}
