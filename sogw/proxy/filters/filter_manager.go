package filters

import (
	"fmt"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/sogw/sogw/proxy/core"
	log "github.com/sirupsen/logrus"
)

type FilterManager struct {
	_ lang.NoCopy

	whenIndex map[When]struct{ before, after int }

	bStepHooks  []HookFunc
	beforeHooks [][]HookFunc
	bh2apIndex  [][]struct{ aStep, aPair int }

	aStepHooks []HookFunc
	afterHooks [][]HookFunc
	aPairIndex [][]int
	ah2apIndex [][]int
}

func NewFilterManager() *FilterManager {
	return &FilterManager{
		whenIndex: make(map[When]struct{ before, after int }),
	}
}

func (fs *FilterManager) Init(cfg map[string]interface{}) error {
	for name, c := range cfg {
		if len(name) > 0 {
			continue
		}
		var cfg map[string]interface{}
		if val, ok := c.(map[string]interface{}); ok {
			cfg = val
		}
		if ff, ok := supportFilters[name]; ok {
			if ff.HookFuncFactory != nil {
				f, err := ff.HookFuncFactory(cfg)
				if err != nil {
					log.Error("[filters] invalid filter hook config, ", err)
					return err
				}
				fs.AddHook(ff.When, f)
			}
			if ff.HookPairFactory != nil {
				f, err := ff.HookPairFactory(cfg)
				if err != nil {
					log.Error("[filters] invalid filter pair config, ", err)
					return err
				}
				fs.AddPair(ff.When, f)
			}
		} else {
			err := fmt.Errorf("filter %s not support", name)
			log.Error("[filters] ", err)
			return err
		}
	}
	return nil
}

func (fs *FilterManager) PushStepBefore(when When, hook HookFunc) {
	if hook == nil {
		panic(fmt.Sprintf("hook must not be nil"))
	}
	blen := len(fs.bStepHooks)
	fs.whenIndex[when] = struct{ before, after int }{blen, -1}
	fs.bStepHooks = append(fs.bStepHooks, hook)
	fs.beforeHooks = append(fs.beforeHooks, nil)
	fs.bh2apIndex = append(fs.bh2apIndex, nil)
	return
}

func (fs *FilterManager) PushStepAfter(when When, hook HookFunc) {
	alen := len(fs.aStepHooks)
	fs.whenIndex[when] = struct{ before, after int }{-1, alen}
	fs.aStepHooks = append(fs.aStepHooks, hook)
	fs.afterHooks = append(fs.afterHooks, nil)
	fs.aPairIndex = append(fs.aPairIndex, nil)
	fs.ah2apIndex = append(fs.ah2apIndex, nil)
	return
}

func (fs *FilterManager) PushStepPair(before When, bhook HookFunc, after When, ahook HookFunc) {
	blen, alen := len(fs.bStepHooks), len(fs.aStepHooks)
	fs.whenIndex[before] = struct{ before, after int }{blen, alen}
	fs.whenIndex[after] = struct{ before, after int }{-1, alen}
	fs.bStepHooks = append(fs.bStepHooks, bhook)
	fs.beforeHooks = append(fs.beforeHooks, nil)
	fs.bh2apIndex = append(fs.bh2apIndex, nil)
	fs.aStepHooks = append(fs.aStepHooks, ahook)
	fs.afterHooks = append(fs.afterHooks, nil)
	fs.aPairIndex = append(fs.aPairIndex, nil)
	fs.ah2apIndex = append(fs.ah2apIndex, nil)
	return
}

func (fs *FilterManager) AddPair(w When, pair HookPair) {
	if pair == nil {
		panic(fmt.Sprintf("hook pair must not be nil"))
	}
	idx, ok := fs.whenIndex[w]
	if !ok {
		panic(fmt.Sprintf("invalid When (%v)", w))
	}
	if idx.before >= 0 && idx.after >= 0 {
		fs.beforeHooks[idx.before] = append(fs.beforeHooks[idx.before], pair.Start)
		ap := len(fs.aPairIndex[idx.after])
		fs.bh2apIndex[idx.before] = append(fs.bh2apIndex[idx.before], struct{ aStep, aPair int }{idx.after, ap})
		fs.aPairIndex[idx.after] = append(fs.aPairIndex[idx.after], len(fs.afterHooks[idx.after]))
		fs.afterHooks[idx.after] = append(fs.afterHooks[idx.after], pair.End)
		fs.ah2apIndex[idx.after] = append(fs.ah2apIndex[idx.after], ap)
	} else {
		panic(fmt.Sprintf("this When (%v) can't do AddPair", w))
	}
}

func (fs *FilterManager) AddHook(w When, hook HookFunc) {
	if hook == nil {
		panic(fmt.Sprintf("hook must not be nil"))
	}
	idx, ok := fs.whenIndex[w]
	if !ok {
		panic(fmt.Sprintf("invalid When (%v)", w))
	}
	if idx.before >= 0 {
		fs.beforeHooks[idx.before] = append(fs.beforeHooks[idx.before], hook)
		if idx.after >= 0 {
			ap := len(fs.aPairIndex[idx.after])
			fs.bh2apIndex[idx.before] = append(fs.bh2apIndex[idx.before], struct{ aStep, aPair int }{idx.after, ap - 1})
		} else {
			last := len(fs.bh2apIndex[idx.before]) - 1
			if last >= 0 {
				fs.bh2apIndex[idx.before] = append(fs.bh2apIndex[idx.before], fs.bh2apIndex[idx.before][last])
			} else {
				fs.bh2apIndex[idx.before] = append(fs.bh2apIndex[idx.before], struct{ aStep, aPair int }{-1, -1})
			}
		}
	} else {
		fs.afterHooks[idx.after] = append(fs.afterHooks[idx.after], hook)
		ap := len(fs.aPairIndex[idx.after])
		fs.ah2apIndex[idx.after] = append(fs.ah2apIndex[idx.after], ap-1)
	}
}

func (fs *FilterManager) Do(ctx *core.RequestContext) (err error) {
	var si int
	var s HookFunc
	for si, s = range fs.bStepHooks {
		for hi, h := range fs.beforeHooks[si] {
			if err = h(ctx); err != nil {
				ctx.Err = err
				fs.logError(err)
				fs.abortBeforeWithError(ctx, si, hi)
				return
			}
		}
		if s != nil {
			err = s(ctx)
			if err != nil {
				ctx.Err = err
				fs.abortBeforeWithError(ctx, si, len(fs.beforeHooks[si])-1)
				return
			}
		}
	}
	for si = len(fs.aStepHooks) - 1; si >= 0; si-- {
		s := fs.aStepHooks[si]
		if s != nil {
			err = s(ctx)
			if err != nil {
				ctx.Err = err
				fs.abortAfterWithError(ctx, si, len(fs.afterHooks[si])-1)
				return
			}
		}
		hs := fs.afterHooks[si]
		for hi := len(hs) - 1; hi >= 0; hi-- {
			err = hs[hi](ctx)
			if err != nil {
				ctx.Err = err
				fs.logError(err)
				fs.abortAfterWithError(ctx, si, hi-1)
				return
			}
		}
	}
	return
}

func (fs *FilterManager) abortBeforeWithError(ctx *core.RequestContext, step, hook int) {
	var err error
	var idxs []int
	if hook >= 0 {
		idx := fs.bh2apIndex[step][hook]
		if idx.aStep < 0 {
			return
		}
		step = idx.aStep
		idxs = fs.aPairIndex[step]
		hooks := fs.afterHooks[step]
		for i := idx.aPair; i >= 0; i-- {
			err = hooks[idxs[i]](ctx)
			if err != nil {
				fs.logError(err)
			}
		}
		step--
	} else {
		for step--; step >= 0; step-- {
			idx := fs.bh2apIndex[step]
			l := len(idx)
			if l > 0 {
				step = idx[l-1].aStep
				break
			}
		}
	}
	for ; step >= 0; step-- {
		idxs = fs.aPairIndex[step]
		hooks := fs.afterHooks[step]
		for i := len(idxs) - 1; i >= 0; i-- {
			err = hooks[idxs[i]](ctx)
			if err != nil {
				fs.logError(err)
			}
		}
	}
}

func (fs *FilterManager) abortAfterWithError(ctx *core.RequestContext, step, hook int) {
	var err error
	var idxs []int
	if hook >= 0 {
		idx := fs.ah2apIndex[step][hook]
		if idx < 0 {
			return
		}
		idxs = fs.aPairIndex[step]
		hooks := fs.afterHooks[step]
		for i := idx; i >= 0; i-- {
			err = hooks[idxs[i]](ctx)
			if err != nil {
				fs.logError(err)
			}
		}
		step--
	} else {
		for step--; step >= 0; step-- {
			idx := fs.ah2apIndex[step]
			l := len(idx)
			if l > 0 {
				step = idx[l-1]
				break
			}
		}
	}
	for ; step >= 0; step-- {
		idxs = fs.aPairIndex[step]
		hooks := fs.afterHooks[step]
		for i := len(idxs) - 1; i >= 0; i-- {
			err = hooks[idxs[i]](ctx)
			if err != nil {
				fs.logError(err)
			}
		}
	}
}

func (fs *FilterManager) logError(err error) {
	if err != ErrExit && err != ErrAbort {
		log.Error("[filters] ", err)
	} else if cobrax.Flags.Debug {
		text := ""
		if err == ErrExit {
			text = "exit"
		} else {
			text = "abort"
		}
		log.Debug("[filter] ", text)
	}
}
