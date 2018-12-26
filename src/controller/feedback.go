package controller

import (
	"config"
	"constant"
	"fmt"
	"model"
	"net/http"
	"util"

	"github.com/labstack/echo"
)

/**
 * @apiDefine CreateFeedback CreateFeedback
 * @apiDescription 添加反馈
 *
 * @apiParam  {String} contact_way 联系方式
 * @apiParam  {String} content 内容
 * @apiParam  {object[]} imgs 图片
 * @apiParam  {String} imgs.url 评论图片URL
 * @apiParam  {String} imgs.micro_url 评论图片缩略图URL
 *
 * @apiParamExample  {json} Request-Example:
 *     {
 *       "contact_way": "contact_way",
 *       "content": "content",
 *       "imgs": [{
 *           "url": "https://phs.<username>.net/PhsComment-20530f80-af23-11e7-8e0d-0715760288c3.jpg",
 *           "micro_url": "https://phs.<username>.net/PhsComment-Small-20530f81-af23-11e7-8e0d-0715760288c3.jpg",
 *         }]
 *     }
 *
 * @apiSuccess {Number} status=200 状态码
 * @apiSuccess {Object} data 正确返回数据
 *
 * @apiSuccessExample Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": 200,
 *       "data": ""
 *     }
 *
 * @apiError {Number} status 状态码
 * @apiError {String} err_msg 错误信息
 *
 * @apiErrorExample Error-Response:
 *     HTTP/1.1 401 Unauthorized
 *     {
 *       "status": 401,
 *       "err_msg": "Unauthorized"
 *     }
 */
/**
 * @api {post} /api/v1/feedback CreateFeedback
 * @apiVersion 1.0.0
 * @apiName CreateFeedback
 * @apiGroup Feedback
 * @apiUse CreateFeedback
 */
func CreateFeedback(c echo.Context) error {
	feedback := model.Feedback{}
	err := c.Bind(&feedback)
	if err != nil {
		writeFeedbackLog("CreateFeedback", constant.ErrorMsgParamWrong, err)
		return retError(c, http.StatusBadRequest, http.StatusBadRequest, constant.ErrorMsgParamWrong)
	}

	// 提醒管理员有人反馈了
	go func() {
		imgHTML := "<img src=\"%s\"  alt=\"反馈图片\" />"

		content := fmt.Sprintf(constant.EmailFeedbackNotice, feedback.ContactWay, feedback.Content)
		for _, img := range feedback.Imgs {
			content += "<br />" + fmt.Sprintf(imgHTML, img.URL)
		}
		util.SendEmail("小灵通", "小灵通反馈", content, config.Conf.EmailInfo.To)
	}()

	return retData(c, "")
}

func writeFeedbackLog(funcName, errMsg string, err error) {
	writeLog("feedback.go", funcName, errMsg, err)
}
