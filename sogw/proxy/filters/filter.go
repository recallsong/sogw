package filters

import "github.com/recallsong/sogw/sogw/proxy/core"

// HookPair hook对
type HookPair interface {
	Start(c *core.RequestContext) error
	End(c *core.RequestContext) error // 除非Start返回ErrExit或者ErrAbort，否则执行Start，必会执行End
}

// HookFunc hook函数
type HookFunc func(c *core.RequestContext) error

// When 拦截的时间点
type When int8

func (w When) String() string {
	return whenNames[w]
}

const (
	BeforeAll      = When(iota) // 所有处理开始前
	BeforeForward               // 处理api请求前
	BeforeDispatch              // 准备发送服务器前，或执行lambda前
	AfterDispatch               // 发送服务器请求后，或之行lambda后
	AfterForward                // 处理api响应后
	AfterAll                    // 所有处理完毕后，或者任何发生出错时
)

var whenNames = map[When]string{
	BeforeAll:      "BeforeAll",
	BeforeForward:  "BeforeForward",
	BeforeDispatch: "BeforeDispatch",
	AfterDispatch:  "AfterDispatch",
	AfterForward:   "AfterForward",
	AfterAll:       "AfterAll",
}

type FilterFactory struct {
	When            When
	HookPairFactory func(cfg map[string]interface{}) (HookPair, error)
	HookFuncFactory func(cfg map[string]interface{}) (HookFunc, error)
}

var supportFilters map[string]FilterFactory

func RegisterFilter(name string, ff FilterFactory) {
	supportFilters[name] = ff
	return
}
