package filters

import "errors"

var (
	ErrAbort = errors.New("abort filters") // 不执行当前的以及后注册的hook，但会之行前注册的hook，并结束当前请求
	ErrExit  = errors.New("exit filters")  // 不执行所有Hook，立即结束当前请求
)
