package model

type Img struct {
	URL      string `bson:"url" json:"url"`           // 图片URL
	MicroURL string `bson:"microUrl" json:"microUrl"` // 缩略图URL
}
