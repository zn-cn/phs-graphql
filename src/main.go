package main

import (
	"config"
	"constant"
	"controller"
	"net/http"

	"github.com/robfig/cron"
)

func main() {
	go startTimer()
	startWeb()
}

func startWeb() {
	// REST 部分：用于认证和未开放的API
	http.HandleFunc("/api/v1/login", controller.Login)
	http.HandleFunc("/api/unopen/group/action/join", controller.JoinGroupFromOfficialAccounts)

	// Graphql 部分：后台主体部分
	http.HandleFunc("/api/graphql", controller.Graphql)

	http.ListenAndServe(config.Conf.AppInfo.Addr, nil)
}

func startTimer() {
	c := cron.New()

	if config.Conf.AppInfo.Env == "prod" {
		c.AddFunc(constant.TimerEveryHour, controller.StartHourTimer)
	}
	c.AddFunc(constant.TimerBackUpdate, controller.StartBackUpdate)
	c.AddFunc(constant.TimerSendDayNotice, controller.StartDayTimer)
	c.AddFunc(constant.TimerSendWeekNotice, controller.StartWeekTimer)

	c.Start()
}
