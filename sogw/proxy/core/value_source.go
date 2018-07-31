package core

import (
	"net"
	"strconv"
	"time"

	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/store/meta"
	"github.com/valyala/fasthttp"
)

type ValueContexts []ValueContext

func (vcs ValueContexts) Get(ctx *RequestContext, name string) (string, bool) {
	for i := len(vcs) - 1; i >= 0; i-- {
		if val, ok := vcs[i].Get(ctx, name); ok {
			return val, ok
		}
	}
	for i, n := range ctx.PathNames {
		if n == name {
			return ctx.PathValues[i], true
		}
	}
	return "", false
}

type ValueContext map[string]*meta.ValueItem

func (vc ValueContext) Get(ctx *RequestContext, name string) (string, bool) {
	if item, ok := vc[name]; ok {
		fn, ok := valueSourceToGetter[item.Source]
		if ok {
			return fn(ctx, item.Name)
		}
	}
	return "", false
}

type valueGetterFunc func(ctx *RequestContext, name string) (string, bool)

var valueSourceToGetter = map[meta.ValueSource]valueGetterFunc{
	meta.ValueSource_Fixed:         getValueByName,
	meta.ValueSource_ReqHeader:     getValueFromReqHeader,
	meta.ValueSource_ReqCookie:     getValueFromReqCookie,
	meta.ValueSource_ReqPathParam:  getValueFromReqPath,
	meta.ValueSource_ReqQueryParam: getValueFromReqQuery,
	meta.ValueSource_ReqFormData:   getValueFromReqFormData,
	meta.ValueSource_ReqJSONBody:   getValueFromReqJSONBody,
	meta.ValueSource_ReqXMLBody:    getValueFromReqXMLBody,
	meta.ValueSource_Request:       getValueFromRequest,

	meta.ValueSource_RespHeader:   getValueFromRespHeader,
	meta.ValueSource_RespCookie:   getValueFromRespCookie,
	meta.ValueSource_RespJSONBody: getValueFromRespJSONBody,
	meta.ValueSource_RespXMLBody:  getValueFromRespXMLBody,
	meta.ValueSource_Response:     getValueFromResponse,

	meta.ValueSource_System: getValueFromSystem,
}

func getValueByName(ctx *RequestContext, name string) (string, bool) {
	return name, true
}

func getValueFromReqHeader(ctx *RequestContext, name string) (string, bool) {
	val := ctx.ReqCtx.Request.Header.Peek(name)
	if val == nil {
		return "", false
	}
	return reflectx.BytesToString(val), true
}

func getValueFromReqCookie(ctx *RequestContext, name string) (string, bool) {
	val := ctx.ReqCtx.Request.Header.Cookie(name)
	if val == nil {
		return "", false
	}
	return reflectx.BytesToString(val), true
}

func getValueFromReqPath(ctx *RequestContext, name string) (string, bool) {
	for i, n := range ctx.PathNames {
		if n == name {
			return ctx.PathValues[i], true
		}
	}
	return "", false
}

func getValueFromReqQuery(ctx *RequestContext, name string) (string, bool) {
	val := ctx.ReqCtx.URI().QueryArgs().Peek(name)
	if val == nil {
		return "", false
	}
	return reflectx.BytesToString(val), true
}

func getValueFromReqFormData(ctx *RequestContext, name string) (string, bool) {
	// TODO
	return "", false
}

func getValueFromReqJSONBody(ctx *RequestContext, name string) (string, bool) {
	// TODO
	return "", false
}

func getValueFromReqXMLBody(ctx *RequestContext, name string) (string, bool) {
	// TODO
	return "", false
}

func getValueFromRequest(ctx *RequestContext, name string) (string, bool) {
	switch name {
	case "ip":
		return ctx.ReqCtx.RemoteIP().String(), true
	case "port":
		addr := ctx.ReqCtx.RemoteAddr()
		if ta, ok := addr.(*net.TCPAddr); ok {
			return strconv.Itoa(ta.Port), true
		} else if ua, ok := addr.(*net.UDPAddr); ok {
			return strconv.Itoa(ua.Port), true
		}
	case "addr":
		return ctx.ReqCtx.RemoteAddr().String(), true
	case "body":
		return reflectx.BytesToString(ctx.ReqCtx.Request.Body()), true
	}
	return "", false
}

func getValueFromRespHeader(ctx *RequestContext, name string) (string, bool) {
	if ctx.ForwardResp != nil {
		val := ctx.ForwardResp.Header.Peek(name)
		if val != nil {
			return reflectx.BytesToString(val), true
		}
	}
	return "", false
}

func getValueFromRespCookie(ctx *RequestContext, name string) (string, bool) {
	if ctx.ForwardResp != nil {
		ck := fasthttp.AcquireCookie()
		ck.SetKey(name)
		if ctx.ForwardResp.Header.Cookie(ck) {
			fasthttp.ReleaseCookie(ck)
			return reflectx.BytesToString(ck.Value()), true
		}
		fasthttp.ReleaseCookie(ck)
	}
	return "", false
}

func getValueFromRespJSONBody(ctx *RequestContext, name string) (string, bool) {
	// TODO
	return "", false
}

func getValueFromRespXMLBody(ctx *RequestContext, name string) (string, bool) {
	// TODO
	return "", false
}

func getValueFromResponse(ctx *RequestContext, name string) (string, bool) {
	switch name {
	case "status":
		if ctx.ForwardResp != nil {
			return strconv.Itoa(ctx.ForwardResp.StatusCode()), true
		}
	case "body":
		if ctx.ForwardResp != nil {
			return reflectx.BytesToString(ctx.ForwardResp.Body()), true
		}
	}
	return "", false
}

var systemFuncMap = map[string]func(ctx *RequestContext) string{
	"now()": func(ctx *RequestContext) string {
		return strconv.FormatInt(time.Now().Unix(), 10)
	},
	"now_ms()": func(ctx *RequestContext) string {
		return strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	},
}

func getValueFromSystem(ctx *RequestContext, name string) (string, bool) {
	if fn, ok := systemFuncMap[name]; ok {
		return fn(ctx), true
	}
	return "", false
}
