package proxy

import (
	"bytes"
	"sync"
	"time"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/sogw/sogw/proxy/core"
	"github.com/recallsong/sogw/sogw/proxy/router"
	"github.com/recallsong/sogw/store"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
)

func (p *HttpProxy) initStore(scfg *StoreConfig) error {
	s, err := store.New(scfg.Url, scfg.Options)
	if err != nil {
		log.Error("[proxy] store.New failed : ", err)
		return err
	}
	p.rtCtx = core.NewRuntimeContext()
	sc := newStoreCache(p, s)
	err = sc.Load()
	if err != nil {
		return err
	}
	if scfg.Watch {
		err = sc.DoWatch()
		if err != nil {
			return err
		}
	}
	return nil
}

type hostEvent struct {
	data *meta.Host
	op   meta.Operation
}

type authEvent struct {
	data *meta.Auth
	op   meta.Operation
}

type routeEvent struct {
	data *meta.Route
	op   meta.Operation
}

type serviceEvent struct {
	id   string
	data *meta.Service
	cfg  *meta.ServiceConfig
	api  *meta.Api
	svr  *meta.Server
	op   meta.Operation
}

type serviceCache struct {
	Meta    *meta.Service           `json:"meta"`
	Cfg     *meta.ServiceConfig     `json:"cfg"`
	Apis    map[string]*meta.Api    `json:"apis"`
	Servers map[string]*meta.Server `json:"svrs"`
}

func newServiceCache(m *meta.Service) *serviceCache {
	return &serviceCache{
		Meta:    m,
		Apis:    make(map[string]*meta.Api),
		Servers: make(map[string]*meta.Server),
	}
}

type storeCache struct {
	store store.Store
	pxy   *HttpProxy

	hosts    map[string]*meta.Host
	auths    map[string]*meta.Auth
	routes   map[string]*meta.Route
	services map[string]*serviceCache

	hostCh    chan *hostEvent
	authCh    chan *authEvent
	routeCh   chan *routeEvent
	serviceCh chan *serviceEvent
}

func newStoreCache(p *HttpProxy, s store.Store) *storeCache {
	return &storeCache{
		pxy:       p,
		store:     s,
		hosts:     make(map[string]*meta.Host),
		auths:     make(map[string]*meta.Auth),
		routes:    make(map[string]*meta.Route),
		services:  make(map[string]*serviceCache),
		hostCh:    make(chan *hostEvent, 512),
		authCh:    make(chan *authEvent, 512),
		routeCh:   make(chan *routeEvent, 1024),
		serviceCh: make(chan *serviceEvent, 1024),
	}
}

func (sc *storeCache) Load() (err error) {
	defer func() {
		if err != nil {
			cerr := sc.store.Close()
			if cerr != nil {
				log.Error("[proxy] fail to close store : ", cerr)
			}
			sc.store = nil
			log.Error("[proxy] fail to load store : ", err)
		}
	}()
	err = sc.store.GetHosts(func(item *meta.Host) {
		if item != nil {
			if err := item.Valid(); err != nil {
				log.Error("[proxy] invalid host, ", err.Error())
				return
			}
			sc.hosts[item.Id] = item
		}
	})
	if err != nil {
		return err
	}
	err = sc.store.GetAuths(func(item *meta.Auth) {
		if item != nil {
			if err := item.Valid(); err != nil {
				log.Error("[proxy] invalid auth, ", err.Error())
				return
			}
			sc.auths[item.Id] = item
		}
	})
	if err != nil {
		return err
	}
	err = sc.store.GetRoutes(func(item *meta.Route) {
		if item != nil {
			if err := item.Valid(); err != nil {
				log.Error("[proxy] invalid route, ", err.Error())
				return
			}
			sc.routes[item.Id] = item
		}
	})
	if err != nil {
		return err
	}
	err = sc.store.GetServices(func(item *meta.Service) {
		if item != nil {
			if err := item.Valid(); err != nil {
				log.Error("[proxy] invalid service, ", err.Error())
				return
			}
			sc.services[item.Id] = newServiceCache(item)
		}
	})
	if err != nil {
		return err
	}
	for _, s := range sc.services {
		cfg, err := sc.store.GetServiceCfg(s.Meta.Id)
		if err != nil {
			return err
		}
		if cfg != nil {
			if err := cfg.Valid(); err != nil {
				log.Error("[proxy] invalid service cfg, ", err.Error())
			} else {
				s.Cfg = cfg
			}
		}
		err = sc.store.GetApis(s.Meta.Id, func(item *meta.Api) {
			if item != nil {
				if err := item.Valid(); err != nil {
					log.Error("[proxy] invalid api, ", err.Error())
					return
				}
				s.Apis[item.Id] = item
			}
		})
		if err != nil {
			return err
		}
		err = sc.store.GetServers(s.Meta.Id, func(item *meta.Server) {
			if item != nil {
				if err := item.Valid(); err != nil {
					log.Error("[proxy] invalid server, ", err.Error())
					return
				}
				s.Servers[item.Id] = item
			}
		})
		if err != nil {
			return err
		}
	}
	start := time.Now()
	sc.SyncRuntimeContext()
	log.Infof("[proxy] build RuntimeContext < cost=%v, hosts=%d, auths=%d, routes=%d, service=%d >",
		time.Now().Sub(start), len(sc.hosts), len(sc.auths), len(sc.routes), len(sc.services))
	return nil
}

func (sc *storeCache) SyncRuntimeContext() {
	hosts := make(map[string]*core.Host)
	for _, item := range sc.hosts {
		hosts[item.Value] = core.NewHost(item)
	}
	auths := make(map[string]*core.Auth)
	for _, item := range sc.auths {
		auths[item.Id] = core.NewAuth(item)
	}
	services := make(map[string]*core.Service)
	router := sc.MakeRouter()
	for _, item := range sc.services {
		ser := core.NewService(item.Meta)
		ser.Init(item.Cfg)
		for _, s := range item.Servers {
			ser.Servers[s.Id] = core.NewServer(s)
		}
		ser.ServerList = make([]*core.Server, 0, len(ser.Servers))
		for _, s := range ser.Servers {
			ser.ServerList = append(ser.ServerList, s)
		}
		for _, a := range item.Apis {
			ser.Apis[a.Id] = core.NewApi(a, ser)
		}
		services[item.Meta.Id] = ser
	}
	sc.pxy.rtCtx.Update(hosts, auths, router, services)
}

func (sc *storeCache) MakeRouter() *router.Router {
	rt := router.New()
	for _, r := range sc.routes {
		if len(r.Method) <= 0 || len(r.Path) <= 0 {
			log.Errorf("[proxy] invalid route, method=%s, path=%s", r.Method, r.Path)
			continue
		}
		route := core.NewRoute(r)
		if r.Method == "*" {
			for _, m := range router.Methods {
				if cobrax.Flags.Debug {
					log.Debugf("[proxy] add route %s, %s", m, r.Path)
				}
				rt.Add(m, r.Path, route)
			}
		} else {
			if cobrax.Flags.Debug {
				log.Debugf("[proxy] add route %s, %s", r.Method, r.Path)
			}
			rt.Add(r.Method, r.Path, route)
		}
	}
	return rt
}

func (sc *storeCache) DoWatch() (err error) {
	err = sc.store.Watch(sc, sc.pxy.waitClose.CloseCh, &sc.pxy.waitClose.WaitGroup)
	if err != nil {
		cerr := sc.store.Close()
		if cerr != nil {
			log.Error("[proxy] fail to close store : ", cerr)
		}
		sc.store = nil
		log.Error("[proxy] fail to watch store : ", err)
		return err
	}
	go sc.doFetch(sc.pxy.waitClose.CloseCh, &sc.pxy.waitClose.WaitGroup)
	return err
}

func (sc *storeCache) RecvHost(op meta.Operation, data *meta.Host) {
	sc.hostCh <- &hostEvent{
		op:   op,
		data: data,
	}
}
func (sc *storeCache) RecvAuth(op meta.Operation, data *meta.Auth) {
	sc.authCh <- &authEvent{
		op:   op,
		data: data,
	}
}
func (sc *storeCache) RecvRoute(op meta.Operation, data *meta.Route) {
	sc.routeCh <- &routeEvent{
		op:   op,
		data: data,
	}
}
func (sc *storeCache) RecvService(op meta.Operation, data *meta.Service) {
	sc.serviceCh <- &serviceEvent{
		op:   op,
		id:   data.Id,
		data: data,
	}
}
func (sc *storeCache) RecvServiceConfig(op meta.Operation, service string, data *meta.ServiceConfig) {
	sc.serviceCh <- &serviceEvent{
		op:  op,
		id:  service,
		cfg: data,
	}
}
func (sc *storeCache) RecvApi(op meta.Operation, service string, data *meta.Api) {
	sc.serviceCh <- &serviceEvent{
		op:  op,
		id:  service,
		api: data,
	}
}
func (sc *storeCache) RecvServer(op meta.Operation, service string, data *meta.Server) {
	sc.serviceCh <- &serviceEvent{
		op:  op,
		id:  service,
		svr: data,
	}
}

func (sc *storeCache) doFetch(stop <-chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	var hflg, aflg, rflg bool
	for {
		services := make(map[string]*core.Service)
		select {
		case evt, ok := <-sc.hostCh:
			if !ok {
				return
			}
			sc.updateHost(evt)
			hflg = true
		case evt, ok := <-sc.authCh:
			if !ok {
				return
			}
			sc.updateAuth(evt)
			aflg = true
		case evt, ok := <-sc.routeCh:
			if !ok {
				return
			}
			sc.updateRoute(evt)
			rflg = true
		case evt, ok := <-sc.serviceCh:
			if !ok {
				return
			}
			sc.updateService(evt)
			services[evt.id] = nil
		case <-stop:
			return
		}
		afterCh := time.After(5 * time.Second)
	loop:
		for {
			select {
			case evt, ok := <-sc.hostCh:
				if !ok {
					return
				}
				sc.updateHost(evt)
				hflg = true
			case evt, ok := <-sc.authCh:
				if !ok {
					return
				}
				sc.updateAuth(evt)
				aflg = true
			case evt, ok := <-sc.routeCh:
				if !ok {
					return
				}
				sc.updateRoute(evt)
				rflg = true
			case evt, ok := <-sc.serviceCh:
				if !ok {
					return
				}
				sc.updateService(evt)
				services[evt.id] = nil
			case <-stop:
				return
			case <-afterCh:
				break loop
			}
		}
		var (
			hosts  core.Hosts
			auths  map[string]*core.Auth
			routes *router.Router
		)
		rc := sc.pxy.rtCtx
		start := time.Now()
		if hflg {
			hosts = make(core.Hosts)
			for _, item := range sc.hosts {
				hosts[item.Value] = core.NewHost(item)
			}
		}
		if aflg {
			auths = make(map[string]*core.Auth)
			for _, item := range sc.auths {
				auths[item.Id] = core.NewAuth(item)
			}
		}
		if rflg {
			routes = sc.MakeRouter()
		}
		if len(services) > 0 {
			for id, _ := range services {
				item, ok := sc.services[id]
				if !ok {
					services[id] = nil
					continue
				}
				ser := core.NewService(item.Meta)
				ser.Init(item.Cfg)
				for _, s := range item.Servers {
					ser.Servers[s.Id] = core.NewServer(s)
				}
				ser.ServerList = make([]*core.Server, 0, len(ser.Servers))
				for _, s := range ser.Servers {
					ser.ServerList = append(ser.ServerList, s)
					// sc.pxy.hcw.AddServer(s)
					// TODO
				}
				for _, a := range item.Apis {
					ser.Apis[a.Id] = core.NewApi(a, ser)
				}
				services[id] = ser
			}
		}
		var oldServices []*core.Service
		rc.Lock.Lock()
		if hflg {
			rc.Hosts = hosts
		}
		if aflg {
			rc.Auths = auths
		}
		if rflg {
			rc.Router = routes
		}
		if services != nil {
			for id, ser := range services {
				service := rc.Services[id]
				if ser == nil {
					delete(rc.Services, id)
				} else {
					rc.Services[id] = ser
				}
				if service != nil {
					oldServices = append(oldServices, service)
				}
			}
		}
		sc.pxy.jobs.Update(rc)
		rc.Lock.Unlock()
		if oldServices != nil {
			for _, item := range oldServices {
				err := item.Close()
				if err != nil {
					log.Error("fail to close service ", err.Error())
				}
			}
		}
		msg := bytes.Buffer{}
		msg.WriteString("[proxy] update RuntimeContext < cost=%v, ")
		if hflg {
			msg.WriteString("*")
		}
		msg.WriteString("hosts=%d, ")
		if aflg {
			msg.WriteString("*")
		}
		msg.WriteString("auths=%d, ")
		if rflg {
			msg.WriteString("*")
		}
		msg.WriteString("routes=%d, ")
		if len(services) > 0 {
			msg.WriteString("*")
		}
		msg.WriteString("service=%d >")
		log.Infof(msg.String(), time.Now().Sub(start), len(sc.hosts), len(sc.auths), len(sc.routes), len(sc.services))
		services = nil
		hflg, aflg, rflg = false, false, false
	}
}

func (sc *storeCache) updateHost(evt *hostEvent) {
	data := evt.data
	if data == nil {
		return
	}
	if evt.op == meta.OperationDelete {
		delete(sc.hosts, data.Id)
	} else if err := data.Valid(); err != nil {
		log.Errorf("invalid host, %s", err.Error())
	} else if evt.op == meta.OperationUpdate || evt.op == meta.OperationCreate {
		sc.hosts[data.Id] = data
	}
}

func (sc *storeCache) updateAuth(evt *authEvent) {
	data := evt.data
	if data == nil {
		return
	}
	if evt.op == meta.OperationDelete {
		delete(sc.auths, data.Id)
	} else if err := data.Valid(); err != nil {
		log.Errorf("invalid auth, %s", err.Error())
	} else if evt.op == meta.OperationUpdate || evt.op == meta.OperationCreate {
		sc.auths[data.Id] = data
	}
}

func (sc *storeCache) updateRoute(evt *routeEvent) {
	data := evt.data
	if data == nil {
		return
	}
	if evt.op == meta.OperationDelete {
		delete(sc.routes, data.Id)
	} else if err := data.Valid(); err != nil {
		log.Errorf("invalid route, %s", err.Error())
	} else if evt.op == meta.OperationUpdate || evt.op == meta.OperationCreate {
		sc.routes[data.Id] = data
	}
}

func (sc *storeCache) updateService(evt *serviceEvent) {
	if evt.api != nil {
		sc.updateApi(evt.op, evt.id, evt.api)
	} else if evt.svr != nil {
		sc.updateServer(evt.op, evt.id, evt.svr)
	} else if evt.cfg != nil {
		sc.updateServiceCfg(evt.op, evt.id, evt.cfg)
	} else if evt.data != nil {
		if ser, ok := sc.services[evt.id]; ok {
			if evt.op == meta.OperationDelete {
				delete(sc.services, evt.id)
			} else if err := evt.data.Valid(); err != nil {
				log.Errorf("invalid service, %s", err.Error())
			} else if evt.op == meta.OperationUpdate || evt.op == meta.OperationCreate {
				ser.Meta = evt.data
			}
		} else if err := evt.data.Valid(); err != nil {
			log.Errorf("invalid service, %s", err.Error())
		} else if evt.op == meta.OperationCreate || evt.op == meta.OperationUpdate {
			sc.services[evt.id] = newServiceCache(evt.data)
		}
	}
}

func (sc *storeCache) updateServer(op meta.Operation, service string, data *meta.Server) {
	if data == nil {
		return
	}
	if ser, ok := sc.services[service]; ok {
		if op == meta.OperationDelete {
			delete(ser.Servers, data.Id)
		} else if err := data.Valid(); err != nil {
			log.Errorf("invalid server, %s", err.Error())
		} else if op == meta.OperationUpdate || op == meta.OperationCreate {
			ser.Servers[data.Id] = data
		}
	} else if err := data.Valid(); err != nil {
		log.Errorf("invalid server, %s", err.Error())
	} else if op == meta.OperationUpdate || op == meta.OperationCreate {
		ser := newServiceCache(&meta.Service{Id: service})
		sc.services[service] = ser
		ser.Servers[data.Id] = data
	}
}

func (sc *storeCache) updateServiceCfg(op meta.Operation, service string, data *meta.ServiceConfig) {
	if data == nil {
		return
	}
	if ser, ok := sc.services[service]; ok {
		if op == meta.OperationDelete {
			ser.Cfg = core.NewServiceConfig()
		} else if err := data.Valid(); err != nil {
			log.Errorf("invalid service cfg, %s", err.Error())
		} else if op == meta.OperationUpdate || op == meta.OperationCreate {
			ser.Cfg = data
		}
	} else if err := data.Valid(); err != nil {
		log.Errorf("invalid service cfg, %s", err.Error())
	} else if op == meta.OperationUpdate || op == meta.OperationCreate {
		ser := newServiceCache(&meta.Service{Id: service})
		sc.services[service] = ser
		ser.Cfg = data
	}
}

func (sc *storeCache) updateApi(op meta.Operation, service string, data *meta.Api) {
	if data == nil {
		return
	}
	if ser, ok := sc.services[service]; ok {
		if op == meta.OperationDelete {
			delete(ser.Apis, data.Id)
		} else if err := data.Valid(); err != nil {
			log.Errorf("invalid api, %s", err.Error())
		} else if op == meta.OperationUpdate || op == meta.OperationCreate {
			ser.Apis[data.Id] = data
		}
	} else if err := data.Valid(); err != nil {
		log.Errorf("invalid api, %s", err.Error())
	} else if op == meta.OperationUpdate || op == meta.OperationCreate {
		ser := newServiceCache(&meta.Service{Id: service})
		sc.services[service] = ser
		ser.Apis[data.Id] = data
	}
}
