package controller

func writeTemplateLog(funcName, errMsg string, err error) {
	writeLog("template.go", funcName, errMsg, err)
}
