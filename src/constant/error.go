package constant

import "errors"

const (
	ErrorMsgParamWrong = "param wrong"
	ErrorMsgUserCreate = "创建用户错误"
)

var (
	ErrorOutOfRange    = errors.New("out of range")
	ErrorIDFormatWrong = errors.New("id format is wrong")
	ErrorNotFound      = errors.New("not found")
	ErrorHasExist      = errors.New("has exist")
	ErrorNotExist      = errors.New("not exist")
	ErrorParamWrong    = errors.New("param is wrong")
	ErrorUnAuth        = errors.New("un auth")
	ErrorEmpty         = errors.New("empty error")
)
