package controller

// 定时器
import "model"

func StartHourTimer() {
	if isProd {
		model.UpdateRedisAccessToken()
	}
	model.UpdateExpireNotice()
}

func StartDayTimer() {
	model.SendDayNotice()
}

func StartWeekTimer() {
	model.SendWeekNotice()
}
