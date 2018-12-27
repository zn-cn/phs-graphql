package param

/*------------------------------------ 二层 -------------------------------------*/

type TypeGroupIDUserIDs struct {
	TypeParam
	GroupIDParam
	UserIDsParam
}

type TypePageCode struct {
	TypeParam
	PageParam
	CodeParam
}

type CodeID struct {
	CodeParam
	IDParam
}
