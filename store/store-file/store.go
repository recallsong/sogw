package file

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/recallsong/go-utils/encoding/jsonx"
	"github.com/recallsong/sogw/store/meta"
)

var ErrNotSupportOp = errors.New("not support this operation")

type serviceCache struct {
	Service *meta.Service
	Cfg     *meta.ServiceConfig
	Apis    map[string]*meta.Api
	Svrs    map[string]*meta.Server
}

func newServiceCache(s *meta.Service) *serviceCache {
	return &serviceCache{
		Service: s,
		Apis:    make(map[string]*meta.Api),
		Svrs:    make(map[string]*meta.Server),
	}
}

type storeCache struct {
	Hosts    map[string]*meta.Host
	Auths    map[string]*meta.Auth
	Routes   map[string]*meta.Route
	Services map[string]*serviceCache
}

func newStoreCache() *storeCache {
	return &storeCache{
		Hosts:    make(map[string]*meta.Host),
		Auths:    make(map[string]*meta.Auth),
		Routes:   make(map[string]*meta.Route),
		Services: make(map[string]*serviceCache),
	}
}

type fileContent struct {
	Hosts    []*meta.Host  `mapstructure:"hosts"`
	Auths    []*meta.Auth  `mapstructure:"auths"`
	Routes   []*meta.Route `mapstructure:"routes"`
	Services []*struct {
		Service *meta.Service       `mapstructure:"service"`
		Cfg     *meta.ServiceConfig `mapstructure:"cfg"`
		Apis    []*meta.Api         `mapstructure:"apis"`
		Svrs    []*meta.Server      `mapstructure:"svrs"`
	} `mapstructure:"services"`
}

// FileStore is readonly store
type FileStore struct {
	viper *viper.Viper
	cache *storeCache
	look  sync.RWMutex
}

func NewStore(addrs, prefix string, options map[string]interface{}) (*FileStore, error) {
	idx := strings.LastIndex(prefix, "/")
	if idx >= 0 {
		addrs += prefix[0 : idx+1]
		prefix = prefix[idx+1:]
	} else {
		addrs += "/"
	}
	idx = strings.LastIndex(prefix, ".")
	if idx >= 0 {
		prefix = prefix[0:idx]
	}
	log.Infof("[file] path: %s , name: %s", addrs, prefix)
	viper := viper.New()
	viper.AddConfigPath(addrs)
	viper.SetConfigName(prefix)
	fs := &FileStore{viper: viper}
	cache, err := fs.load()
	if err != nil {
		return nil, err
	}
	fs.cache = cache
	return fs, nil
}

func (fs *FileStore) load() (*storeCache, error) {
	err := fs.viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	cache := newStoreCache()
	content := &fileContent{}
	err = fs.viper.Unmarshal(content)
	if err != nil {
		return nil, err
	}
	for _, v := range content.Hosts {
		if v == nil {
			continue
		}
		if len(v.Id) <= 0 {
			v.InitId()
		}
		if err := v.Valid(); err != nil {
			return nil, fmt.Errorf("invalid host %s, %s", jsonx.Marshal(v), err.Error())
		}
		if _, ok := cache.Hosts[v.Id]; ok {
			return nil, fmt.Errorf("duplicate host %s", jsonx.Marshal(v))
		}
		cache.Hosts[v.Id] = v
	}
	for _, v := range content.Auths {
		if v == nil {
			continue
		}
		if len(v.Id) <= 0 {
			v.InitId()
		}
		if err := v.Valid(); err != nil {
			return nil, fmt.Errorf("invalid auth %s, %s", jsonx.Marshal(v), err.Error())
		}
		if _, ok := cache.Auths[v.Id]; ok {
			return nil, fmt.Errorf("duplicate auth %s", jsonx.Marshal(v))
		}
		cache.Auths[v.Id] = v
	}
	for _, v := range content.Routes {
		if v == nil {
			continue
		}
		if len(v.Id) <= 0 {
			v.InitId()
		}
		if err := v.Valid(); err != nil {
			return nil, fmt.Errorf("invalid route %s, %s", jsonx.Marshal(v), err.Error())
		}
		if _, ok := cache.Routes[v.Id]; ok {
			return nil, fmt.Errorf("duplicate route %s", jsonx.Marshal(v))
		}
		cache.Routes[v.Id] = v
	}
	for _, v := range content.Services {
		if v == nil {
			continue
		}
		if v.Service == nil {
			v.Service = &meta.Service{
				Name: "default",
			}
		}
		if len(v.Service.Id) <= 0 {
			v.Service.InitId()
		}
		if err := v.Service.Valid(); err != nil {
			return nil, fmt.Errorf("invalid service %s, %s", jsonx.Marshal(v.Service), err.Error())
		}
		if _, ok := cache.Services[v.Service.Id]; ok {
			return nil, fmt.Errorf("duplicate service %s", jsonx.Marshal(v.Service))
		}
		sc := newServiceCache(v.Service)
		cache.Services[v.Service.Id] = sc
		if v.Cfg == nil {
			v.Cfg = &meta.ServiceConfig{}
		}
		if len(v.Cfg.Id) <= 0 {
			v.Cfg.InitId()
		}
		if err := v.Cfg.Valid(); err != nil {
			return nil, fmt.Errorf("invalid service config %s, %s", jsonx.Marshal(v.Cfg), err.Error())
		}
		sc.Cfg = v.Cfg
		for _, a := range v.Apis {
			if a == nil {
				continue
			}
			if len(a.Id) <= 0 {
				a.InitId()
			}
			if err := a.Valid(); err != nil {
				return nil, fmt.Errorf("invalid service api %s, %s", jsonx.Marshal(a), err.Error())
			}
			if _, ok := sc.Apis[a.Id]; ok {
				return nil, fmt.Errorf("duplicate api %s", jsonx.Marshal(a))
			}
			sc.Apis[a.Id] = a
		}
		for _, s := range v.Svrs {
			if s == nil {
				continue
			}
			if len(s.Id) <= 0 {
				s.InitId()
			}
			if err := s.Valid(); err != nil {
				return nil, fmt.Errorf("invalid service server %s, %s", jsonx.Marshal(s), err.Error())
			}
			if _, ok := sc.Svrs[s.Id]; ok {
				return nil, fmt.Errorf("duplicate server %s", jsonx.Marshal(s))
			}
			sc.Svrs[s.Id] = s
		}
	}
	return cache, nil
}

func (fs *FileStore) Client() interface{} {
	return fs.viper
}

func (fs *FileStore) Close() error {
	fs.viper = nil
	return nil
}

func (fs *FileStore) Name() string {
	return "file"
}

func (fs *FileStore) Clean() error {
	return ErrNotSupportOp
}
func (fs *FileStore) RegistryGateway(gateway *meta.Gateway, keepAlive time.Duration) error {
	// do nothing
	return nil
}
func (fs *FileStore) GetGateways(handler func(item *meta.Gateway)) error {
	return ErrNotSupportOp
}
func (fs *FileStore) PutHost(host *meta.Host) error {
	return ErrNotSupportOp
}
func (fs *FileStore) RemoveHost(id string) error {
	return ErrNotSupportOp
}

func (fs *FileStore) GetHosts(handler func(item *meta.Host)) error {
	fs.look.RLock()
	defer fs.look.RUnlock()
	for _, v := range fs.cache.Hosts {
		handler(v.Copy())
	}
	return nil
}
func (fs *FileStore) GetHost(id string) (*meta.Host, error) {
	fs.look.RLock()
	defer fs.look.RUnlock()
	if v, ok := fs.cache.Hosts[id]; ok {
		return v.Copy(), nil
	}
	return nil, nil
}

func (fs *FileStore) PutAuth(auth *meta.Auth) error {
	return ErrNotSupportOp
}
func (fs *FileStore) RemoveAuth(id string) error {
	return ErrNotSupportOp
}
func (fs *FileStore) GetAuths(handler func(item *meta.Auth)) error {
	fs.look.RLock()
	defer fs.look.RUnlock()
	for _, v := range fs.cache.Auths {
		handler(v.Copy())
	}
	return nil
}
func (fs *FileStore) GetAuth(id string) (*meta.Auth, error) {
	fs.look.RLock()
	defer fs.look.RUnlock()
	if v, ok := fs.cache.Auths[id]; ok {
		return v.Copy(), nil
	}
	return nil, nil
}

func (fs *FileStore) PutRoute(route *meta.Route) error {
	return ErrNotSupportOp
}
func (fs *FileStore) RemoveRoute(id string) error {
	return ErrNotSupportOp
}
func (fs *FileStore) GetRoutes(handler func(item *meta.Route)) error {
	fs.look.RLock()
	defer fs.look.RUnlock()
	for _, v := range fs.cache.Routes {
		handler(v.Copy())
	}
	return nil
}
func (fs *FileStore) GetRoute(id string) (*meta.Route, error) {
	fs.look.RLock()
	defer fs.look.RUnlock()
	if v, ok := fs.cache.Routes[id]; ok {
		return v.Copy(), nil
	}
	return nil, nil
}

func (fs *FileStore) PutService(service *meta.Service) error {
	return ErrNotSupportOp
}
func (fs *FileStore) RemoveService(service string, cascade bool) error {
	return ErrNotSupportOp
}
func (fs *FileStore) GetServices(handler func(item *meta.Service)) error {
	fs.look.RLock()
	defer fs.look.RUnlock()
	for _, v := range fs.cache.Services {
		if v.Service != nil {
			handler(v.Service.Copy())
		}
	}
	return nil
}
func (fs *FileStore) GetService(id string) (*meta.Service, error) {
	fs.look.RLock()
	defer fs.look.RUnlock()
	if v, ok := fs.cache.Services[id]; ok {
		if v.Service != nil {
			return v.Service.Copy(), nil
		}
	}
	return nil, nil
}

func (fs *FileStore) PutServiceCfg(service string, cfg *meta.ServiceConfig) error {
	return ErrNotSupportOp
}
func (fs *FileStore) RemoveServiceCfg(service string) error {
	return ErrNotSupportOp
}
func (fs *FileStore) GetServiceCfg(service string) (*meta.ServiceConfig, error) {
	fs.look.RLock()
	defer fs.look.RUnlock()
	if v, ok := fs.cache.Services[service]; ok {
		if v.Cfg != nil {
			return v.Cfg.Copy(), nil
		}
	}
	return nil, nil
}

func (fs *FileStore) PutApi(service string, api *meta.Api) error {
	return ErrNotSupportOp
}
func (fs *FileStore) RemoveApi(service string, id string) error {
	return ErrNotSupportOp
}
func (fs *FileStore) GetApis(service string, handler func(item *meta.Api)) error {
	fs.look.RLock()
	defer fs.look.RUnlock()
	if v, ok := fs.cache.Services[service]; ok {
		for _, v := range v.Apis {
			handler(v.Copy())
		}
	}
	return nil
}
func (fs *FileStore) GetApi(service, id string) (*meta.Api, error) {
	fs.look.RLock()
	defer fs.look.RUnlock()
	if v, ok := fs.cache.Services[service]; ok {
		if v, ok := v.Apis[id]; ok {
			return v.Copy(), nil
		}
	}
	return nil, nil
}

func (fs *FileStore) PutServer(service string, server *meta.Server) error {
	return ErrNotSupportOp
}
func (fs *FileStore) RemoveServer(service string, id string) error {
	return ErrNotSupportOp
}
func (fs *FileStore) GetServers(service string, handler func(item *meta.Server)) error {
	fs.look.RLock()
	defer fs.look.RUnlock()
	if v, ok := fs.cache.Services[service]; ok {
		for _, v := range v.Svrs {
			handler(v.Copy())
		}
	}
	return nil
}
func (fs *FileStore) GetServer(service, id string) (*meta.Server, error) {
	fs.look.RLock()
	defer fs.look.RUnlock()
	if v, ok := fs.cache.Services[service]; ok {
		if v, ok := v.Svrs[id]; ok {
			return v.Copy(), nil
		}
	}
	return nil, nil
}
