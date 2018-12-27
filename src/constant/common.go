package constant

import "time"

const (
	TemplateIDSendNotice      = "srUKofBihJvr6X5zJD8ajPmxOXC80YUmn2_Oco9YTp8"
	TemplateNoticeFirst       = "别忘了，作业写了吗！"
	TemplateNoticeRemark      = "更多详情请点击查看小程序"
	TemplateIDGroupJoin       = "Bg6NcFMJ_RC6LfaX9a21cnZk_GTfk90psYw94q4ZHiw"
	TemplateGroupJoinFirst    = "亲爱的%s，你已成功加入班级！"
	TemplateTime              = "%d年%d月%d日"
	TemplateGroupJoinKeyword2 = "%s"
	TemplateGroupJoinRemark   = "你将及时收到下周作业预告和日常作业提醒\n更多详情点击查看小程序"

	MPPagePath = "pages/home/Home" // 微信公众号跳转小程序的页面, 支持带参数

	/****************************************** timer ****************************************/

	TimerBackUpdate     = "0 0 4 * * *"  // 后台更新
	TimerSendDayNotice  = "0 0 9 * * *"  // 每天九点提醒
	TimerSendWeekNotice = "0 0 18 * * 5" // 每周周五18点提醒
	TimerEveryHour      = "@hourly"      // 每小时触发

	/****************************************** user ****************************************/

	/****************************************** feedback ****************************************/

	/****************************************** wechat ****************************************/

	WechatScanCodeJoinPhsMPGroup = "join/phs-mp/group/%s" // 加入班级事件 join/phs-mp/group/<group-code>

	/****************************************** img ****************************************/

	ImgOps       = "imageView2/2/w/160/h/160" // 图片做缩略处理：w: 160, h: 160
	ImgURIPrefix = "http://<image hostname>/"
	ImgMicroSize = 160

	ImgDefaultGraoupHead = "http://<image hostname>/mp/head/logo1_%E6%96%B9.png"

	ImgSuffix = ".jpg"

	ImgPrefixHomework      = "mp/homework/"
	ImgPrefixMicroHomework = "mp/homework/micro/"
	ImgPrefixHead          = "mp/head/"
	ImgPrefixMicroHead     = "mp/head/micro/"
	ImgPrefixFeedback      = "mp/feedback/"
	ImgPrefixMicroFeedback = "mp/feedback/micro/"

	/****************************************** token ****************************************/

	TokenQiniuExpire = 7200
	JWTContextKey    = "user"
	JWTAuthScheme    = "Bearer"
	JWTExpire        = time.Hour * 24 * 7

	/****************************************** other ****************************************/

	// api prefix
	APIPrefix = "/api/v1"

	EmailFeedbackNotice = "联系方式：%s <br /> 反馈内容：%s <br />"
)
