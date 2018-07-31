package proxy

import (
	"net/http"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/sogw/proxy/core"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

func (p *HttpProxy) Handler(reqc *fasthttp.RequestCtx) {
	ctx := core.NewRequestContext(reqc)
	p.rtCtx.Lock.RLock()
	ctx.Router = p.rtCtx.Router
	ctx.Hosts = p.rtCtx.Hosts
	ctx.Auths = p.rtCtx.Auths
	ctx.Services = p.rtCtx.Services
	p.rtCtx.Lock.RUnlock()
	p.filters.Do(ctx)
	p.FinishRequest(ctx)
}

func (p *HttpProxy) FinishRequest(ctx *core.RequestContext) {
	if ctx.Err == nil {
		status := ctx.ReqCtx.Response.StatusCode()
		var backend string
		if ctx.ForwardReq != nil {
			backend = " -> " + reflectx.BytesToString(ctx.ForwardReq.URI().FullURI())
		}
		if status >= 400 {
			if ctx.ForwardResp != nil && ctx.ForwardResp.StatusCode() >= 400 {
				log.Errorf("[proxy] %s%s %d (by backend)", reflectx.BytesToString(ctx.ReqCtx.RequestURI()), backend, status)
			} else {
				log.Errorf("[proxy] %s%s %d", reflectx.BytesToString(ctx.ReqCtx.RequestURI()), backend, status)
			}
		} else {
			log.Infof("[proxy] %s%s %d", reflectx.BytesToString(ctx.ReqCtx.RequestURI()), backend, status)
		}
	} else {
		log.Error("[proxy] ", ctx.Err)
	}
	core.ReleaseRequestContext(ctx)
}

func doRoute(ctx *core.RequestContext) (err error) {
	reqc := ctx.ReqCtx
	if !ctx.Hosts.ValidateHost(ctx) {
		ctx.WriteError(fasthttp.StatusNotFound)
		return core.ErrHostNotAllow
	}
	url := reflectx.BytesToString(reqc.Path())
	method := reflectx.BytesToString(reqc.Method())
	rt := ctx.Router
	result := rt.NewResult()
	if !rt.Find(method, url, result) {
		ctx.WriteError(fasthttp.StatusNotFound)
		return core.ErrRouteNotFound
	}
	if result.MethodNotAllow {
		ctx.WriteError(fasthttp.StatusMethodNotAllowed)
		return core.ErrMethodNotAllow
	}
	route := result.Dest.(*core.Route)
	if route.Meta.Status == meta.Status_Close {
		if cobrax.Flags.Debug {
			log.Debugf("[handle] %s %s route closed", route.Meta.Method, route.Meta.Path)
		}
		ctx.WriteError(fasthttp.StatusNotFound)
		return core.ErrRouteNotFound
	}
	if route.Context != nil {
		ctx.ValueContexts = append(ctx.ValueContexts, route.Context)
	}
	ctx.PathNames, ctx.PathValues = result.PathParams, result.PathValues
	api := route.Dispatch(ctx)
	if api == nil {
		ctx.WriteError(fasthttp.StatusNotFound)
		return core.ErrRouteNotFound
	}
	return
}

func doForward(ctx *core.RequestContext) error {
	a := ctx.Api
	if a.Context != nil {
		ctx.ValueContexts = append(ctx.ValueContexts, a.Context)
	}
	if a.Validators.Validate(ctx) == false {
		return core.ErrApiValidateFailed
	}
	err := a.DoAuth(ctx)
	if err != nil {
		return err
	}
	reqc := ctx.ReqCtx
	freq := fasthttp.AcquireRequest()
	freq.Reset()
	reqc.Request.CopyTo(freq)
	ctx.ForwardReq = freq
	err = a.RewriteURL(ctx)
	if err != nil {
		return err
	}
	a.SetForwardHeader(ctx)
	a.SetForwardCookie(ctx)
	if len(a.Meta.Lambda) > 0 {
		return nil
	}
	var svr *core.Server
	if a.Server != nil {
		svr = a.Server
	} else if ctx.Host != nil && len(ctx.Host.Meta.SvrId) > 0 {
		svr = ctx.Service.Servers[ctx.Host.Meta.SvrId]
	} else {
		svr = ctx.Service.LB.Select(ctx, ctx.Service.ServerList)
	}
	if svr == nil {
		if cobrax.Flags.Debug {
			log.Debugf("[handle] no server available for api(%s) service(%s)", a.Meta.Id, ctx.Service.Meta.Name)
		}
		ctx.WriteError(fasthttp.StatusServiceUnavailable)
		return core.ErrServiceUnavailable
	}
	ctx.Server = svr
	return nil
}

func doDispatch(ctx *core.RequestContext) error {
	if len(ctx.Api.Meta.Lambda) > 0 {
		a := ctx.Api
		return a.EvalLambda(ctx, a.Meta.Lambda)
	} else {
		fresp := fasthttp.AcquireResponse()
		ctx.ForwardResp = fresp
		err := ctx.Server.Forward(ctx.ForwardReq, fresp)
		if err != nil {
			ctx.WriteError(http.StatusInternalServerError)
			return err
		}
	}
	return nil
}

func finishForward(ctx *core.RequestContext) (err error) {
	a := ctx.Api
	a.SetResponseHeader(ctx)
	a.SetResponseCookie(ctx)
	dst := &ctx.ReqCtx.Response
	ctx.ForwardResp.Header.CopyTo(&dst.Header)
	err = ctx.ForwardResp.BodyWriteTo(dst.BodyWriter())
	if err != nil {
		log.Error("[handle] write to response error : ", err)
		ctx.WriteError(http.StatusInternalServerError)
		return err
	}
	return nil
}

func finishDispatch(ctx *core.RequestContext) error {
	if ctx.Err != nil {
		ctx.WriteError(fasthttp.StatusInternalServerError)
	}
	return nil
}
