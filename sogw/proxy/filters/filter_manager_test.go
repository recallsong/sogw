package filters

import (
	"strings"
	"testing"

	"github.com/recallsong/sogw/sogw/proxy/core"
	"github.com/stretchr/testify/assert"
)

type testPair struct {
	Name   string
	result *[]string
	Abort  bool
}

func (t *testPair) Start(c *core.RequestContext) error {
	*t.result = append(*t.result, "pair.Start "+t.Name)
	if t.Abort {
		return ErrAbort
	}
	return nil
}
func (t *testPair) End(c *core.RequestContext) error {
	*t.result = append(*t.result, "pair.End "+t.Name)
	return nil
}

func testHook(result *[]string, name string) func(*core.RequestContext) error {
	return func(c *core.RequestContext) error {
		*result = append(*result, name)
		return nil
	}
}

func TestFilters(t *testing.T) {
	fs := NewFilterManager()
	result := []string{}
	fs.PushStepPair(BeforeAll, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeAll")
		return nil
	}, AfterAll, func(c *core.RequestContext) error {
		result = append(result, "DoAfterAll")
		return nil
	})
	fs.PushStepPair(BeforeForward, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeForward")
		return nil
	}, AfterForward, func(c *core.RequestContext) error {
		result = append(result, "DoAfterForward")
		return nil
	})
	fs.PushStepPair(BeforeDispatch, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeDispatch")
		return nil
	}, AfterDispatch, func(c *core.RequestContext) error {
		result = append(result, "DoAfterDispatch")
		return nil
	})
	fs.AddHook(BeforeAll, testHook(&result, "BeforeAll"))
	fs.AddHook(AfterAll, testHook(&result, "AfterAll 1"))
	fs.AddPair(BeforeAll, &testPair{"All", &result, false})
	fs.AddHook(AfterAll, testHook(&result, "AfterAll 2"))

	fs.AddHook(BeforeForward, testHook(&result, "BeforeForward 1"))
	fs.AddPair(BeforeForward, &testPair{"Forward 1", &result, false})
	fs.AddHook(BeforeForward, testHook(&result, "BeforeForward 2"))
	fs.AddHook(AfterForward, testHook(&result, "AfterForward 1"))
	fs.AddPair(BeforeForward, &testPair{"Forward 2", &result, false})
	fs.AddHook(AfterForward, testHook(&result, "AfterForward 2"))

	fs.AddPair(BeforeDispatch, &testPair{"Dispatch", &result, false})
	fs.AddHook(BeforeDispatch, testHook(&result, "BeforeDispatch"))
	fs.AddHook(AfterDispatch, testHook(&result, "AfterDispatch"))
	fs.Do(nil)
	expect := []string{
		"BeforeAll", "pair.Start All",
		"DoBeforeAll",
		"BeforeForward 1", "pair.Start Forward 1", "BeforeForward 2", "pair.Start Forward 2",
		"DoBeforeForward",
		"pair.Start Dispatch", "BeforeDispatch",
		"DoBeforeDispatch",
		"DoAfterDispatch",
		"AfterDispatch",
		"pair.End Dispatch",
		"DoAfterForward",
		"AfterForward 2", "pair.End Forward 2", "AfterForward 1", "pair.End Forward 1",
		"DoAfterAll",
		"AfterAll 2", "pair.End All", "AfterAll 1",
	}
	assert.Equal(t, strings.Join(expect, ","), strings.Join(result, ","))
}

func TestFilters_abort1(t *testing.T) {
	fs := NewFilterManager()
	result := []string{}
	fs.PushStepPair(BeforeAll, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeAll")
		return nil
	}, AfterAll, func(c *core.RequestContext) error {
		result = append(result, "DoAfterAll")
		return nil
	})
	fs.PushStepPair(BeforeForward, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeForward")
		return nil
	}, AfterForward, func(c *core.RequestContext) error {
		result = append(result, "DoAfterForward")
		return nil
	})
	fs.PushStepPair(BeforeDispatch, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeDispatch")
		return nil
	}, AfterDispatch, func(c *core.RequestContext) error {
		result = append(result, "DoAfterDispatch")
		return nil
	})
	fs.AddHook(BeforeAll, testHook(&result, "BeforeAll"))
	fs.AddHook(AfterAll, testHook(&result, "AfterAll 1"))
	fs.AddPair(BeforeAll, &testPair{"All", &result, false})
	fs.AddHook(AfterAll, testHook(&result, "AfterAll 2"))

	fs.AddHook(BeforeForward, testHook(&result, "BeforeForward 1"))
	fs.AddPair(BeforeForward, &testPair{"Forward 1", &result, false})
	fs.AddHook(BeforeForward, testHook(&result, "BeforeForward 2"))
	fs.AddHook(AfterForward, testHook(&result, "AfterForward 1"))
	fs.AddPair(BeforeForward, &testPair{"Forward 2", &result, true})
	fs.AddHook(AfterForward, testHook(&result, "AfterForward 2"))

	fs.AddPair(BeforeDispatch, &testPair{"Dispatch", &result, false})
	fs.AddHook(BeforeDispatch, testHook(&result, "BeforeDispatch"))
	fs.AddHook(AfterDispatch, testHook(&result, "AfterDispatch"))
	fs.Do(core.NewRequestContext(nil))
	expect := []string{
		"BeforeAll", "pair.Start All",
		"DoBeforeAll",
		"BeforeForward 1", "pair.Start Forward 1", "BeforeForward 2", "pair.Start Forward 2",
		"pair.End Forward 2", "pair.End Forward 1", "pair.End All",
	}
	assert.Equal(t, strings.Join(expect, ","), strings.Join(result, ","))
}

func TestFilters_error(t *testing.T) {
	fs := NewFilterManager()
	result := []string{}
	fs.PushStepPair(BeforeAll, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeAll")
		return nil
	}, AfterAll, func(c *core.RequestContext) error {
		result = append(result, "DoAfterAll")
		return nil
	})
	fs.PushStepPair(BeforeForward, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeForward")
		return core.ErrServiceUnavailable
	}, AfterForward, func(c *core.RequestContext) error {
		result = append(result, "DoAfterForward")
		return nil
	})
	fs.PushStepPair(BeforeDispatch, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeDispatch")
		return nil
	}, AfterDispatch, func(c *core.RequestContext) error {
		result = append(result, "DoAfterDispatch")
		return nil
	})
	fs.AddHook(BeforeAll, testHook(&result, "BeforeAll"))
	fs.AddHook(AfterAll, testHook(&result, "AfterAll 1"))
	fs.AddPair(BeforeAll, &testPair{"All", &result, false})
	fs.AddHook(AfterAll, testHook(&result, "AfterAll 2"))

	fs.AddHook(BeforeForward, testHook(&result, "BeforeForward 1"))
	fs.AddPair(BeforeForward, &testPair{"Forward 1", &result, false})
	fs.AddHook(BeforeForward, testHook(&result, "BeforeForward 2"))
	fs.AddHook(AfterForward, testHook(&result, "AfterForward 1"))
	fs.AddPair(BeforeForward, &testPair{"Forward 2", &result, false})
	fs.AddHook(AfterForward, testHook(&result, "AfterForward 2"))

	fs.AddPair(BeforeDispatch, &testPair{"Dispatch", &result, true})
	fs.AddHook(BeforeDispatch, testHook(&result, "BeforeDispatch"))
	fs.AddHook(AfterDispatch, testHook(&result, "AfterDispatch"))
	ctx := core.NewRequestContext(nil)
	fs.Do(ctx)
	expect := []string{
		"BeforeAll", "pair.Start All",
		"DoBeforeAll",
		"BeforeForward 1", "pair.Start Forward 1", "BeforeForward 2", "pair.Start Forward 2",
		"DoBeforeForward",
		"pair.End Forward 2", "pair.End Forward 1", "pair.End All",
	}
	assert.Equal(t, strings.Join(expect, ","), strings.Join(result, ","))
	assert.Equal(t, core.ErrServiceUnavailable, ctx.Err)
}

func TestFilters_error2(t *testing.T) {
	fs := NewFilterManager()
	result := []string{}
	fs.PushStepPair(BeforeAll, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeAll")
		return nil
	}, AfterAll, func(c *core.RequestContext) error {
		result = append(result, "DoAfterAll")
		return nil
	})
	fs.PushStepPair(BeforeForward, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeForward")
		return nil
	}, AfterForward, func(c *core.RequestContext) error {
		result = append(result, "DoAfterForward")
		return nil
	})
	fs.PushStepPair(BeforeDispatch, func(c *core.RequestContext) error {
		result = append(result, "DoBeforeDispatch")
		return nil
	}, AfterDispatch, func(c *core.RequestContext) error {
		result = append(result, "DoAfterDispatch")
		return nil
	})
	fs.AddHook(BeforeAll, testHook(&result, "BeforeAll"))
	fs.AddHook(AfterAll, testHook(&result, "AfterAll 1"))
	fs.AddPair(BeforeAll, &testPair{"All", &result, false})
	fs.AddHook(AfterAll, testHook(&result, "AfterAll 2"))

	fs.AddHook(BeforeForward, testHook(&result, "BeforeForward 1"))
	fs.AddPair(BeforeForward, &testPair{"Forward 1", &result, false})
	fs.AddHook(BeforeForward, testHook(&result, "BeforeForward 2"))
	fs.AddHook(AfterForward, testHook(&result, "AfterForward 1"))
	fs.AddPair(BeforeForward, &testPair{"Forward 2", &result, false})
	fs.AddHook(AfterForward, testHook(&result, "AfterForward 2"))

	fs.AddPair(BeforeDispatch, &testPair{"Dispatch", &result, false})
	fs.AddHook(BeforeDispatch, testHook(&result, "BeforeDispatch"))
	fs.AddHook(AfterDispatch, func(c *core.RequestContext) error {
		return ErrAbort
	})
	ctx := core.NewRequestContext(nil)
	fs.Do(ctx)
	expect := []string{
		"BeforeAll", "pair.Start All",
		"DoBeforeAll",
		"BeforeForward 1", "pair.Start Forward 1", "BeforeForward 2", "pair.Start Forward 2",
		"DoBeforeForward",
		"pair.Start Dispatch", "BeforeDispatch",
		"DoBeforeDispatch",
		"DoAfterDispatch",
		"pair.End Dispatch", "pair.End Forward 2", "pair.End Forward 1", "pair.End All",
	}
	assert.Equal(t, strings.Join(expect, ","), strings.Join(result, ","))
	assert.Equal(t, ErrAbort, ctx.Err)
}
