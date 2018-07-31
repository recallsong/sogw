package core

import (
	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
)

type ApiCondition struct {
	_       lang.NoCopy
	Matcher Matcher
	ApiId   string
}

type Route struct {
	_        lang.NoCopy
	Meta     *meta.Route
	Context  ValueContext
	ApiConds []*ApiCondition
}

func NewRoute(m *meta.Route) *Route {
	var conds []*ApiCondition
	if len(m.ApiConds) > 0 {
		var conds []*ApiCondition
		for _, c := range m.ApiConds {
			if c != nil && c.Matcher != nil {
				conds = append(conds, &ApiCondition{
					Matcher: *NewMatcher(c.Matcher),
					ApiId:   c.ApiId,
				})
			}
		}
	}
	return &Route{
		Meta:     m,
		Context:  ValueContext(m.Context),
		ApiConds: conds,
	}
}

func (r *Route) Dispatch(ctx *RequestContext) *Api {
	service := r.Meta.Service
	if r.Meta.Service == "" && ctx.Host != nil {
		service = ctx.Host.Meta.Service
	}
	if r.ApiConds != nil {
		for _, c := range r.ApiConds {
			if c.Matcher.Match(ctx) {
				if ser, ok := ctx.Services[service]; ok {
					ctx.Service = ser
					api := ser.Apis[c.ApiId]
					if api == nil || api.Meta.Status == meta.Status_Close {
						if cobrax.Flags.Debug {
							if api == nil {
								log.Debugf("[route] api (id=%s) not found", c.ApiId)
							} else {
								log.Debugf("[route] api (id=%s) has been closed", c.ApiId)
							}
						}
						return nil
					}
					if ctx.Service.Config.Context != nil {
						ctx.ValueContexts = append(ctx.ValueContexts, ctx.Service.Config.Context)
					}
					ctx.Api = api
					return api
				}
				if cobrax.Flags.Debug {
					log.Debugf("[route] service (id=%s) not found", r.Meta.Service)
				}
				return nil
			}
		}
	}
	if ser, ok := ctx.Services[service]; ok {
		ctx.Service = ser
		api := ser.Apis[r.Meta.ApiId]
		if api == nil && ctx.Host != nil {
			api = ser.Apis[ctx.Host.Meta.ApiId]
		}
		if api == nil {
			if cobrax.Flags.Debug {
				log.Debugf("[route] api (id=%s) not found", r.Meta.ApiId)
			}
			return nil
		}
		if api.Meta.Status == meta.Status_Close {
			if cobrax.Flags.Debug {
				log.Debugf("[route] api (id=%s) has been closed", r.Meta.ApiId)
			}
			return nil
		}
		if ctx.Service.Config.Context != nil {
			ctx.ValueContexts = append(ctx.ValueContexts, ctx.Service.Config.Context)
		}
		ctx.Api = api
		return api
	}
	if cobrax.Flags.Debug {
		log.Debugf("[route] service (id=%s) not found", r.Meta.Service)
	}
	return nil
}
