package param

/*------------------------------------ 一层 -------------------------------------*/

type TypeParam struct {
	Type int `json:"type" query:"type"`
}

type StatusParam struct {
	Status int `json:"status" query:"status"`
}

type ImgParam struct {
	ImgURL string `json:"img_url" query:"img_url"`
}

type PageParam struct {
	Page    int `json:"page" query:"page" validate:"min=1"`
	PerPage int `json:"per_page" query:"per_page" validate:"min=1,max=20"`
}

type IDParam struct {
	ID string `json:"id" query:"id"`
}

type GroupIDParam struct {
	GroupID string `json:"group_id" query:"group_id" validate:"required"`
}

type CodeParam struct {
	Code string `json:"code" query:"code" validate:"required"`
}

type UserIDsParam struct {
	UserIDs []string `json:"user_ids" query:"user_ids"`
}

type AvatarNicknameParam struct {
	Nickname  string `json:"nickname" query:"nickname"`
	AvatarURL string `json:"avatar_url" query:"avatar_url"`
}
