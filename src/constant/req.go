package constant

const (
	ReqGroupUpdateOwnerType = iota + 1
	ReqGroupDelOwnerType
	ReqGroupSetManagerType
	ReqGroupUnSetManagerType
	ReqGroupDelManagerType
	ReqGroupDelMemberType
)

const (
	ReqNoticeGetAllType = iota + 1
	ReqNoticeGetByGroupIDType
)

const (
	ReqNoticeUpdate = iota
	ReqNoticeUpdateContentType
	ReqNoticeUpdateImgsType
	ReqNoticeUpdateTitleType
	ReqNoticeUpdateNoteType
	ReqNoticeUpdateNoticeTimeType
	ReqNoticeUpdateGroupIDType
)
