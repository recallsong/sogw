package core

import (
	"sync"

	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/sogw/sogw/proxy/router"
	"github.com/valyala/fasthttp"
)

var httpClient *fasthttp.Client = &fasthttp.Client{Dial: fasthttp.Dial}

type RuntimeContext struct {
	_          lang.NoCopy
	Lock       sync.RWMutex
	Hosts      Hosts
	Auths      map[string]*Auth
	Router     *router.Router
	Services   map[string]*Service
	HttpClient *fasthttp.Client
}

func NewRuntimeContext() *RuntimeContext {
	return &RuntimeContext{
		Hosts:    make(map[string]*Host),
		Auths:    make(map[string]*Auth),
		Services: make(map[string]*Service),
	}
}

func (rt *RuntimeContext) Update(
	hosts Hosts, auths map[string]*Auth,
	router *router.Router, services map[string]*Service) {
	rt.Lock.Lock()
	rt.Hosts = hosts
	rt.Auths = auths
	rt.Router = router
	rt.Services = services
	rt.Lock.Unlock()
}
