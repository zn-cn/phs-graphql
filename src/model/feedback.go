package model

import "gopkg.in/mgo.v2/bson"

type Feedback struct {
	ID     bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
	Type   int           `bson:"type" json:"type"`     // 反馈类别
	Status int           `bson:"status" json:"status"` // 状态：0 未查看，1 已经查看

	UserID     string `bson:"userID" json:"userID"`         // _id
	CreateTime int64  `bson:"createTime" json:"createTime"` // 创建时间
	ContactWay string `bson:"contactWay" json:"contactWay"` // 联系方式
	Content    string `bson:"content"  json:"content"`      // 内容
	Imgs       []Img  `bson:"imgs" json:"imgs"`             // 图片
}
