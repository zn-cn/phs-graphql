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
	ReqNoticeGetByGroupCodeType
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

const (
	ImgTypeHomework = iota + 1
	ImgTypeHead
	ImgTypeFeedback
)

var (
	ImgPrefix = map[int]string{
		ImgTypeHomework: ImgPrefixHomework,
		ImgTypeHead:     ImgPrefixHead,
		ImgTypeFeedback: ImgPrefixFeedback,
	}
	ImgPrefixMicro = map[int]string{
		ImgTypeHomework: ImgPrefixMicroHomework,
		ImgTypeHead:     ImgPrefixMicroHead,
		ImgTypeFeedback: ImgPrefixMicroFeedback,
	}
)
