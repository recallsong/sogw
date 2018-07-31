package core

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type Api struct {
	_          lang.NoCopy
	Meta       *meta.Api
	Context    ValueContext
	Validators Validators
	URLRewrite []string
	Server     *Server
}

func NewApi(m *meta.Api, service *Service) *Api {
	if m.Headers == nil {
		m.Headers = &meta.ApiHeaders{}
	}
	if m.Cookies == nil {
		m.Cookies = &meta.ApiCookies{}
	}
	var svr *Server
	if m.ServerId != "" {
		svr = service.Servers[m.ServerId]
	}
	var vs Validators
	for _, v := range m.Validators {
		if v != nil && v.Matcher != nil {
			vs = append(vs, NewValidator(v))
		}
	}
	return &Api{
		Meta:       m,
		Server:     svr,
		Validators: vs,
		Context:    ValueContext(m.Context),
		URLRewrite: makeURLRewrite(strings.TrimSpace(m.Path)),
	}
}

func makeURLRewrite(path string) (parts []string) {
	i, j, l := 0, 0, len(path)
	if l == 0 {
		path = "$*"
	}
	for i < l {
		if path[i] == ':' {
			if path[:i] != "" {
				parts = append(parts, path[:i])
			}
			j = i
			for i++; i < l && path[i] != '/' && path[i] != '.' && path[i] != '?' &&
				path[i] != '&' && path[i] != '#' && path[i] != '=' &&
				path[i] != '{' && path[i] != '*'; i++ {
			}
			parts = append(parts, path[j:i])
			path = path[i:]
			i, l = 0, len(path)
			continue
		} else if path[i] == '{' {
			if path[:i] != "" {
				parts = append(parts, path[:i])
			}
			j = i
			i++
			if i >= l {
				parts = append(parts, "{")
				return
			}
			for ; i < l && path[i] != '}'; i++ {
			}
			if i >= l {
				parts = append(parts, path[j:])
				return
			}
			parts = append(parts, ":"+path[j+1:i])
			i++
			if i >= l {
				return
			}
			path = path[i:]
			i, l = 0, len(path)
			continue
		} else if path[i] == '*' {
			if path[:i] != "" {
				parts = append(parts, path[:i])
			}
			parts = append(parts, "*")
			path = path[i+1:]
			i, l = 0, len(path)
			continue
		} else if path[i] == '$' {
			if path[:i] != "" {
				parts = append(parts, path[:i])
			}
			i++
			if i >= l {
				return
			}
			if path[i] == '*' || path[i] == '?' {
				parts = append(parts, path[i-1:i+1])
				path = path[i+1:]
				i, l = 0, len(path)
				continue
			} else {
				path = path[i+1:]
				i, l = 0, len(path)
				continue
			}
		}
		i++
	}
	if path != "" {
		parts = append(parts, path)
	}
	return
}

func (a *Api) DoAuth(ctx *RequestContext) error {
	authId := ctx.Service.Config.AuthId
	if a.Meta.AuthId != "" {
		authId = a.Meta.AuthId
	}
	if len(authId) <= 0 {
		return nil
	}
	if auth, ok := ctx.Auths[authId]; ok {
		if auth.Fn(auth.Meta.Config, ctx) {
			return nil
		}
	}
	if cobrax.Flags.Debug {
		log.Debug("[api] auth not found ", authId)
	}
	ctx.WriteError(http.StatusUnauthorized)
	return ErrAuthFailed
}

func (a *Api) RewriteURL(ctx *RequestContext) error {
	if len(a.URLRewrite) <= 0 {
		return nil
	}

	v, _ := ctx.ValueContexts.Get(ctx, "*")
	fmt.Println(a.URLRewrite, v)

	buf := bytes.Buffer{}
	for _, part := range a.URLRewrite {
		if part[0] == ':' {
			v, _ := ctx.ValueContexts.Get(ctx, part[1:])
			buf.WriteString(v)
		} else if part[0] == '*' {
			v, _ := ctx.ValueContexts.Get(ctx, "*")
			buf.WriteString(v)
		} else if len(part) == 2 {
			if part[0] == '$' {
				switch part[1] {
				case '?':
					buf.WriteString(reflectx.BytesToString(ctx.ReqCtx.URI().QueryString()))
				case '*':
					buf.WriteString(reflectx.BytesToString(ctx.ReqCtx.URI().RequestURI()))
				case '&':
					buf.WriteRune('$')
				}
			} else {
				buf.WriteString(part)
			}
		} else {
			buf.WriteString(part)
		}
	}
	bytes := buf.Bytes()
	lastURL := reflectx.BytesToString(bytes)
	if cobrax.Flags.Debug {
		log.Debugf("[api] url rewrite %s -> %s", reflectx.BytesToString(ctx.ReqCtx.URI().RequestURI()), lastURL)
	}
	ctx.ForwardReq.SetRequestURI(lastURL)
	return nil
}

func (a *Api) SetForwardHeader(ctx *RequestContext) {
	if len(a.Meta.Method) > 0 && a.Meta.Method != "*" {
		ctx.ForwardReq.Header.SetMethod(a.Meta.Method)
	}
	for _, h := range a.Meta.Headers.ToBackend {
		v, _ := ctx.ValueContexts.Get(ctx, h.Key)
		ctx.ForwardReq.Header.Add(h.Name, v)
	}
}

func (a *Api) SetForwardCookie(ctx *RequestContext) {
	for _, c := range a.Meta.Cookies.ToBackend {
		v, _ := ctx.ValueContexts.Get(ctx, c.Key)
		ctx.ForwardReq.Header.SetCookie(c.Name, v)
	}
}

func (a *Api) SetResponseHeader(ctx *RequestContext) {
	for _, h := range a.Meta.Headers.ToClient {
		v, _ := ctx.ValueContexts.Get(ctx, h.Key)
		ctx.ReqCtx.Response.Header.Add(h.Name, v)
	}
}

func (a *Api) SetResponseCookie(ctx *RequestContext) {
	if len(a.Meta.Cookies.ToClient) > 0 {
		ck := fasthttp.AcquireCookie()
		now := time.Now()
		for _, c := range a.Meta.Cookies.ToClient {
			v, _ := ctx.ValueContexts.Get(ctx, c.Key)
			if c.Expire > 0 {
				ck.SetExpire(now.Add(time.Duration(c.Expire) * time.Second))
			}
			ck.SetKey(c.Name)
			ck.SetValue(v)
			ctx.ReqCtx.Response.Header.SetCookie(ck)
		}
		fasthttp.ReleaseCookie(ck)
	}
}
