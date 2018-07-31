package core

import (
	"strconv"
	"strings"
	"time"

	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/sogw/proxy/router"
	"github.com/valyala/fasthttp"
)

type RequestContext struct {
	_           lang.NoCopy
	HttpC       *fasthttp.Client
	Start       time.Time
	Err         error
	ReqCtx      *fasthttp.RequestCtx
	ForwardReq  *fasthttp.Request
	ForwardResp *fasthttp.Response
	Attrs       map[string]interface{}
	PathNames   []string
	PathValues  []string

	Hosts    Hosts
	Auths    map[string]*Auth
	Services map[string]*Service
	Router   *router.Router

	Host          *Host
	ValueContexts ValueContexts
	Service       *Service
	Api           *Api
	Server        *Server
}

func NewRequestContext(reqc *fasthttp.RequestCtx) *RequestContext {
	return &RequestContext{
		Start:  time.Now(),
		ReqCtx: reqc,
	}
}
func ReleaseRequestContext(ctx *RequestContext) {
	if ctx.ForwardReq != nil {
		fasthttp.ReleaseRequest(ctx.ForwardReq)
	}
	if ctx.ForwardResp != nil {
		fasthttp.ReleaseResponse(ctx.ForwardResp)
	}
}

func (c *RequestContext) StartOn() time.Time {
	return c.Start
}

func (c *RequestContext) OriginRequestCtx() *fasthttp.RequestCtx {
	return c.ReqCtx
}

func (c *RequestContext) ForwardRequest() *fasthttp.Request {
	return c.ForwardReq
}

func (c *RequestContext) ForwardResponse() *fasthttp.Response {
	return c.ForwardResp
}

func (c *RequestContext) GetRealClientAddr() string {
	xforward := c.ReqCtx.Request.Header.Peek("X-Forwarded-For")
	if nil == xforward {
		return strings.SplitN(c.ReqCtx.RemoteAddr().String(), ":", 2)[0]
	}
	return strings.SplitN(reflectx.BytesToString(xforward), ",", 2)[0]
}

func (c *RequestContext) Error() error {
	return c.Err
}

func (c *RequestContext) SetAttr(key string, value interface{}) {
	if c.Attrs == nil {
		c.Attrs = make(map[string]interface{})
	}
	c.Attrs[key] = value
}

func (c *RequestContext) GetAttr(key string) interface{} {
	if c.Attrs == nil {
		val, ok := c.ValueContexts.Get(c, key)
		if ok {
			return val
		}
		return nil
	}
	val, ok := c.Attrs[key]
	if ok {
		return val
	} else {
		val, ok := c.ValueContexts.Get(c, key)
		if ok {
			return val
		}
	}
	return nil
}

func (c *RequestContext) PathParam(name string) string {
	for i, n := range c.PathNames {
		if n == name {
			return c.PathValues[i]
		}
	}
	return ""
}

func (c *RequestContext) WriteError(statusCode int) {
	c.ReqCtx.Response.Reset()
	c.ReqCtx.SetStatusCode(statusCode)
	c.ReqCtx.SetBodyString(strconv.Itoa(statusCode) + " " + fasthttp.StatusMessage(statusCode))
}
