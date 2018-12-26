package controller

import "model"

func StartHourTimer() {
	model.UpdateRedisAccessToken()
}

func StartDayTimer() {
	model.SendDayNotice()
}

func StartWeekTimer() {
	model.SendWeekNotice()
}

func StartBackUpdate() {
	model.UpdateExpireNotice()
}
