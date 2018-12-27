package model

import (
	"config"
	"constant"
	"fmt"
	"model/db"
	"time"

	"github.com/imroc/req"
	"gopkg.in/mgo.v2/bson"
)

type Template struct {
	ID     bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Type   int           `bson:"type" json:"type"`     // 类型：
	Status int           `bson:"status" json:"status"` // 状态:

	CreatorID string   `bson:"creatorID" json:"creatorID"` // 创建者 unionid
	Notices   []Notice `bson:"notices" json:"notices"`     // 模板内容

	CreateTime int64 `bson:"createTime" json:"create_time"` // 创建时间毫秒时间戳
}

// WechatTemplate 微信模板
type WechatTemplate struct {
	ToUser      string                 `json:"touser"`                // 必须, 接受者OpenID
	TemplateID  string                 `json:"template_id"`           // 必须, 模版ID
	URL         string                 `json:"url,omitempty"`         // 可选, 用户点击后跳转的URL, 该URL必须处于开发者在公众平台网站中设置的域中
	MiniProgram *MiniProgram           `json:"miniprogram,omitempty"` // 可选, 跳小程序所需数据，不需跳小程序可不用传该数据
	Data        map[string]interface{} `json:"data"`                  // 必须, 模板数据
}

type MiniProgram struct {
	AppID    string `json:"appid"` // 必选; 所需跳转到的小程序appid（该小程序appid必须与发模板消息的公众号是绑定关联关系）
	PagePath string `json:"path"`  // 必选; 注意：官方文档错了！
}

type TemplateResult struct {
	Error
	MsgID int64 `json:"msgid"`
}

type Error struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// 模板草稿通过 redis 存储, 取出即删redis草稿

func GetTemplate(id string) (Template, error) {
	if !bson.IsObjectIdHex(id) {
		return Template{}, constant.ErrorIDFormatWrong
	}
	query := bson.M{
		"_id": bson.ObjectIdHex(id),
	}
	return findTemplate(query, DefaultSelector)
}

func SendGroupJoinTemplate(unionid, groupCode string) error {
	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	groupTable := cntrl.GetTable(constant.TableGroup)
	userTable := cntrl.GetTable(constant.TableUser)

	query := bson.M{
		"unionid": unionid,
	}
	selector := bson.M{
		"nickname": 1,
	}

	user := User{}
	err := userTable.Find(query).Select(selector).One(&user)
	if err != nil {
		return err
	}
	query = bson.M{
		"code": groupCode,
		"status": bson.M{
			"$gte": constant.GroupCommonStatus,
		},
	}

	selector = bson.M{
		"nickname": 1,
	}

	group := Group{}
	err = groupTable.Find(query).Select(selector).One(&group)
	if err != nil {
		return err
	}
	template := getGroupTemplate(unionid, user.Nickname, group.Nickname)
	_, err = sendOfficeAccountTemplate([]WechatTemplate{template})
	return err
}

func getNoticeTemplate(unionid, title, content, timeStr string) WechatTemplate {
	data := map[string]interface{}{
		"first": map[string]string{
			"value": constant.TemplateNoticeFirst,
			"color": "#173177",
		},
		"keyword1": map[string]string{
			"value": title,
			"color": "#173177",
		},
		"keyword2": map[string]string{
			"value": content,
			"color": "#173177",
		},
		"keyword3": map[string]string{
			"value": timeStr,
			"color": "#173177",
		},
		"remark": map[string]string{
			"value": constant.TemplateNoticeRemark,
			"color": "#173177",
		},
	}
	return getTemplate(unionid, constant.TemplateIDSendNotice, data)
}

func getGroupTemplate(unionid, userNickname, groupNickname string) WechatTemplate {
	year, month, day := time.Now().Date()
	data := map[string]interface{}{
		"first": map[string]string{
			"value": fmt.Sprintf(constant.TemplateGroupJoinFirst, userNickname),
			"color": "#173177",
		},
		"keyword1": map[string]string{
			"value": fmt.Sprintf(constant.TemplateTime, year, month, day),
			"color": "#173177",
		},
		"keyword2": map[string]string{
			"value": fmt.Sprintf(constant.TemplateGroupJoinKeyword2, groupNickname),
			"color": "#173177",
		},
		"remark": map[string]string{
			"value": constant.TemplateGroupJoinRemark,
			"color": "#173177",
		},
	}
	return getTemplate(unionid, constant.TemplateIDGroupJoin, data)
}

func getTemplate(unionid, templateID string, data map[string]interface{}) WechatTemplate {
	return WechatTemplate{
		ToUser:     unionid,
		TemplateID: templateID,
		MiniProgram: &MiniProgram{
			AppID:    config.Conf.Wechat.AppID,
			PagePath: constant.MPPagePath,
		},
		Data: data,
	}
}

func sendOfficeAccountTemplate(templates []WechatTemplate) ([]TemplateResult, error) {
	data := map[string]interface{}{
		"templates": templates,
	}
	resp, err := req.Post(constant.URLBingYanSendTemplate, req.BodyJSON(&data))
	if err != nil {
		return nil, err
	}
	resData := struct {
		Status int              `json:"status"`
		Data   []TemplateResult `json:"data"`
	}{}
	err = resp.ToJSON(&resData)
	return resData.Data, err
}

/****************************************** template basic action ****************************************/

func findTemplate(query, selectField interface{}) (Template, error) {
	data := Template{}
	cntrl := db.NewCopyMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(constant.TableTemplate)
	err := table.Find(query).Select(selectField).One(&data)
	return data, err
}

func updateTemplate(query, update interface{}) error {
	return updateDoc(constant.TableTemplate, query, update)
}

func insertTemplates(docs ...interface{}) error {
	return insertDocs(constant.TableTemplate, docs...)
}
