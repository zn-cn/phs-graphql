package model

type Img struct {
	URL      string `bson:"url" json:"url"`             // 图片URL
	MicroURL string `bson:"micro_url" json:"micro_url"` // 缩略图URL
}
