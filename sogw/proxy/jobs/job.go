package jobs

import "github.com/recallsong/sogw/sogw/proxy/core"

type Job interface {
	Setup(cfg map[string]interface{}) error
	Run(ctx *core.RuntimeContext, stopCh <-chan struct{}) error
	Close() error
	Update(ctx *core.RuntimeContext) error
}
