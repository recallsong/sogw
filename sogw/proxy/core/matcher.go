package core

import (
	"regexp"

	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/store/meta"
)

type Matcher struct {
	_     lang.NoCopy
	Meta  *meta.Matcher
	Regex *regexp.Regexp
	Func  MatcherFunc
}

func NewMatcher(m *meta.Matcher) *Matcher {
	if m.Kind == meta.MatcherKind_Regex {
		re, err := regexp.Compile(m.Value)
		if err != nil {
			return &Matcher{
				Meta: m,
				Func: matcherInvalid,
			}
		}
		return &Matcher{
			Meta:  m,
			Regex: re,
		}
	} else {
		fn, ok := matcherFuncs[m.Kind]
		if !ok {
			fn = matcherInvalid
		}
		return &Matcher{
			Meta: m,
			Func: fn,
		}
	}
}

func (m *Matcher) Match(ctx *RequestContext) bool {
	val, _ := ctx.ValueContexts.Get(ctx, m.Meta.Key)
	if m.Regex != nil {
		return m.Regex.Match(reflectx.StringToBytes(val))
	}
	return m.Func(val, m.Meta.Value)
}

type MatcherFunc func(source, value string) bool

var matcherFuncs = map[meta.MatcherKind]MatcherFunc{
	meta.MatcherKind_EQ:    matchByEQ,
	meta.MatcherKind_NE:    matchByNE,
	meta.MatcherKind_LT:    matchByLT,
	meta.MatcherKind_LE:    matchByLE,
	meta.MatcherKind_GT:    matchByGT,
	meta.MatcherKind_GE:    matchByGE,
	meta.MatcherKind_Regex: matchByRegex,
}

func matcherInvalid(source, value string) bool {
	return false
}

func matchByRegex(source, value string) bool {
	re, err := regexp.Compile(value)
	if err != nil {
		return false
	}
	return re.Match(reflectx.StringToBytes(source))
}

func matchByEQ(source, value string) bool {
	return source == value
}

func matchByNE(source, value string) bool {
	return source != value
}

func matchByLT(source, value string) bool {
	return source < value
}

func matchByLE(source, value string) bool {
	return source <= value
}

func matchByGT(source, value string) bool {
	return source > value
}

func matchByGE(source, value string) bool {
	return source >= value
}
