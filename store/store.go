package store

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/recallsong/sogw/store/meta"
	"github.com/recallsong/sogw/store/store-etcd"
	"github.com/recallsong/sogw/store/store-file"
)

type Store interface {
	Client() interface{}
	Close() error
	Name() string
	Clean() error

	RegistryGateway(gateway *meta.Gateway, keepAlive time.Duration) error
	GetGateways(handler func(item *meta.Gateway)) error

	PutHost(host *meta.Host) error
	RemoveHost(id string) error
	GetHosts(handler func(item *meta.Host)) error
	GetHost(id string) (*meta.Host, error)

	PutAuth(auth *meta.Auth) error
	RemoveAuth(id string) error
	GetAuths(handler func(item *meta.Auth)) error
	GetAuth(id string) (*meta.Auth, error)

	PutRoute(route *meta.Route) error
	RemoveRoute(id string) error
	GetRoutes(handler func(item *meta.Route)) error
	GetRoute(id string) (*meta.Route, error)

	PutService(service *meta.Service) error
	RemoveService(service string, cascade bool) error
	GetServices(handler func(item *meta.Service)) error
	GetService(id string) (*meta.Service, error)

	PutServiceCfg(service string, cfg *meta.ServiceConfig) error
	RemoveServiceCfg(service string) error
	GetServiceCfg(service string) (*meta.ServiceConfig, error)

	PutApi(service string, api *meta.Api) error
	RemoveApi(service string, id string) error
	GetApis(service string, handler func(item *meta.Api)) error
	GetApi(service, id string) (*meta.Api, error)

	PutServer(service string, server *meta.Server) error
	RemoveServer(service string, id string) error
	GetServers(service string, handler func(item *meta.Server)) error
	GetServer(service, id string) (*meta.Server, error)

	Watch(ln meta.EventListener, stopCh <-chan struct{}, waitStop *sync.WaitGroup) error
}

var storeNewMap = map[string]func(addrs, prefix string, options map[string]interface{}) (Store, error){
	"etcd": func(addrs, prefix string, options map[string]interface{}) (Store, error) {
		return etcd.NewStore(strings.Split(addrs, ","), prefix, options)
	},
	"file": func(addrs, prefix string, options map[string]interface{}) (Store, error) {
		return file.NewStore(addrs, prefix, options)
	},
}

func New(store_url string, options map[string]interface{}) (Store, error) {
	store, addrs, prefix, err := ParseUrl(store_url)
	if err != nil {
		return nil, err
	}
	if len(addrs) <= 0 {
		return nil, errors.New("invalid store addrs")
	}
	if new_fn, ok := storeNewMap[store]; ok {
		return new_fn(addrs, prefix, options)
	}
	return nil, fmt.Errorf("store %s not support yet.", store)
}

func ParseUrl(store_url string) (name, addrs, prefix string, err error) {
	u, err := url.Parse(store_url)
	if err != nil {
		err := fmt.Errorf("fail to parse store address: %v", err)
		return "", "", "", err
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme, u.Host, u.Path, nil
}
