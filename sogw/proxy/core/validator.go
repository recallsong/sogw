package core

import (
	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/sogw/store/meta"
)

type Validator struct {
	_       lang.NoCopy
	Meta    *meta.Validator
	Matcher Matcher
}

func NewValidator(m *meta.Validator) *Validator {
	return &Validator{
		Meta:    m,
		Matcher: *NewMatcher(m.Matcher),
	}
}

type Validators []*Validator

func (vs Validators) Validate(ctx *RequestContext) bool {
	for _, v := range vs {
		if !v.Matcher.Match(ctx) {
			ctx.ReqCtx.SetStatusCode(int(v.Meta.Status))
			ctx.ReqCtx.WriteString(v.Meta.ErrorMsg)
			return false
		}
	}
	return true
}
