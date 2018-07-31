package etcd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultDialTimeout      = time.Second * 3
	DefaultSlowTxnTimeToLog = time.Second * 1
	DefaultRequestTimeout   = 10 * time.Second
)

type EtcdStore struct {
	Prefix         string
	HostPath       string
	AuthPath       string
	RoutePath      string
	ServicePath    string
	ServicePrefix  string
	ApiPath        string
	ServerPath     string
	ServiceCfgPath string
	GatewayPath    string
	ConfigPath     string
	client         *clientv3.Client
}

func NewStore(addrs []string, prefix string, options map[string]interface{}) (*EtcdStore, error) {
	var endps []string
	for _, addr := range addrs {
		endps = append(endps, fmt.Sprintf("http://%s", addr))
	}
	log.Info("[etcd] addrs ", endps)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endps,
		DialTimeout: DefaultDialTimeout,
	})
	if err != nil {
		return nil, err
	}
	if len(prefix) > 0 && prefix[0] != '/' {
		prefix = "/" + prefix
	}
	return &EtcdStore{
		Prefix:         prefix + "/",
		ConfigPath:     fmt.Sprintf("%s/config/", prefix),
		GatewayPath:    fmt.Sprintf("%s/gateways/", prefix),
		HostPath:       fmt.Sprintf("%s/hosts/", prefix),
		AuthPath:       fmt.Sprintf("%s/auths/", prefix),
		RoutePath:      fmt.Sprintf("%s/routes/", prefix),
		ServicePath:    fmt.Sprintf("%s/services/", prefix),
		ServicePrefix:  fmt.Sprintf("%s/space/", prefix),
		ServerPath:     "/svrs/",
		ApiPath:        "/apis/",
		ServiceCfgPath: "/cfg",
		client:         cli,
	}, nil
}

func (s *EtcdStore) Client() interface{} {
	return s.client
}

func (s *EtcdStore) Close() error {
	return s.client.Close()
}

func (s *EtcdStore) Name() string {
	return "etcd"
}

func (s *EtcdStore) Clean() error {
	return s.delete(s.Prefix, clientv3.WithPrefix())
}

func (s *EtcdStore) RegistryGateway(gateway *meta.Gateway, keepAlive time.Duration) error {
	gateway.InitId()
	if len(gateway.Addrs) <= 0 {
		return errors.New("invalid gateway' Addrs")
	}
	data, err := gateway.Marshal()
	if err != nil {
		return err
	}
	return s.putWithKeepAlive(s.key(s.GatewayPath, gateway.Id), reflectx.BytesToString(data), keepAlive)
}
func (s *EtcdStore) GetGateways(handler func(item *meta.Gateway)) error {
	return s.gets(s.GatewayPath, func() meta.Serializable { return &meta.Gateway{} }, func(sb meta.Serializable) {
		handler(sb.(*meta.Gateway))
	})
}

func (s *EtcdStore) PutHost(host *meta.Host) error {
	if len(host.Value) <= 0 {
		return errors.New("host.Value shoud not be empty")
	}
	host.InitId()
	data, err := host.Marshal()
	if err != nil {
		return err
	}
	return s.put(s.key(s.HostPath, host.Id), reflectx.BytesToString(data))
}
func (s *EtcdStore) RemoveHost(id string) error {
	return s.delete(s.key(s.HostPath, id))
}
func (s *EtcdStore) GetHosts(handler func(item *meta.Host)) error {
	return s.gets(s.HostPath, func() meta.Serializable { return &meta.Host{} }, func(sb meta.Serializable) {
		handler(sb.(*meta.Host))
	})
}
func (s *EtcdStore) GetHost(id string) (*meta.Host, error) {
	resp, err := s.get(s.key(s.HostPath, id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if resp.Count <= 0 {
		return nil, nil
	}
	kv := resp.Kvs[0]
	m := &meta.Host{Id: id}
	err = m.Unmarshal(kv.Value)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *EtcdStore) PutAuth(auth *meta.Auth) error {
	auth.InitId()
	data, err := auth.Marshal()
	if err != nil {
		return err
	}
	return s.put(s.key(s.AuthPath, auth.Id), reflectx.BytesToString(data))
}
func (s *EtcdStore) RemoveAuth(id string) error {
	return s.delete(s.key(s.AuthPath, id))
}
func (s *EtcdStore) GetAuths(handler func(item *meta.Auth)) error {
	return s.gets(s.AuthPath, func() meta.Serializable { return &meta.Auth{} }, func(sb meta.Serializable) {
		handler(sb.(*meta.Auth))
	})
}
func (s *EtcdStore) GetAuth(id string) (*meta.Auth, error) {
	resp, err := s.get(s.key(s.AuthPath, id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if resp.Count <= 0 {
		return nil, nil
	}
	kv := resp.Kvs[0]
	m := &meta.Auth{Id: id}
	err = m.Unmarshal(kv.Value)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *EtcdStore) PutRoute(route *meta.Route) error {
	if len(route.Path) <= 0 {
		route.Path = "/"
	}
	if len(route.Method) <= 0 {
		route.Method = "*"
	}
	if route.Path[0] != '/' {
		route.Path += "/"
	}
	route.InitId()
	data, err := route.Marshal()
	if err != nil {
		return err
	}
	return s.put(s.key(s.RoutePath, route.Id), reflectx.BytesToString(data))
}
func (s *EtcdStore) RemoveRoute(id string) error {
	return s.delete(s.key(s.RoutePath, id))
}
func (s *EtcdStore) GetRoutes(handler func(item *meta.Route)) error {
	return s.gets(s.RoutePath, func() meta.Serializable { return &meta.Route{} }, func(sb meta.Serializable) {
		handler(sb.(*meta.Route))
	})
}
func (s *EtcdStore) GetRoute(id string) (*meta.Route, error) {
	resp, err := s.get(s.key(s.RoutePath, id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if resp.Count <= 0 {
		return nil, nil
	}
	kv := resp.Kvs[0]
	m := &meta.Route{Id: id}
	err = m.Unmarshal(kv.Value)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *EtcdStore) PutServiceCfg(service string, cfg *meta.ServiceConfig) error {
	if _, ok := meta.LoadBalance_name[int32(cfg.LoadBlance)]; !ok {
		cfg.LoadBlance = meta.LoadBalance_RoundRobin
	}
	cfg.InitId()
	data, err := cfg.Marshal()
	if err != nil {
		return err
	}
	return s.put(s.skey(service, s.ServiceCfgPath, ""), reflectx.BytesToString(data))
}
func (s *EtcdStore) RemoveServiceCfg(service string) error {
	return s.delete(s.skey(service, s.ServiceCfgPath, ""))
}
func (s *EtcdStore) GetServiceCfg(service string) (*meta.ServiceConfig, error) {
	resp, err := s.get(s.skey(service, s.ServiceCfgPath, ""), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if resp.Count <= 0 {
		return nil, nil
	}
	kv := resp.Kvs[0]
	cfg := &meta.ServiceConfig{}
	err = cfg.Unmarshal(kv.Value)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *EtcdStore) PutService(ser *meta.Service) error {
	if len(ser.Name) <= 0 {
		return errors.New("service.Name shoud not be empty")
	}
	ser.InitId()
	data, err := ser.Marshal()
	if err != nil {
		return err
	}
	return s.put(s.key(s.ServicePath, ser.Id), reflectx.BytesToString(data))
}
func (s *EtcdStore) RemoveService(service string, cascade bool) error {
	if cascade {
		err := s.delete(s.ServicePrefix+service+"/", clientv3.WithPrefix())
		if err != nil {
			return err
		}
	}
	return s.delete(s.key(s.ServicePath, service))
}
func (s *EtcdStore) GetServices(handler func(item *meta.Service)) error {
	return s.gets(s.ServicePath, func() meta.Serializable { return &meta.Service{} }, func(sb meta.Serializable) {
		handler(sb.(*meta.Service))
	})
}
func (s *EtcdStore) GetService(id string) (*meta.Service, error) {
	resp, err := s.get(s.key(s.ServicePath, id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if resp.Count <= 0 {
		return nil, nil
	}
	kv := resp.Kvs[0]
	m := &meta.Service{Id: id}
	err = m.Unmarshal(kv.Value)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *EtcdStore) PutApi(service string, api *meta.Api) error {
	api.InitId()
	data, err := api.Marshal()
	if err != nil {
		return err
	}
	return s.put(s.skey(service, s.ApiPath, api.Id), reflectx.BytesToString(data))
}
func (s *EtcdStore) RemoveApi(service, id string) error {
	return s.delete(s.skey(service, s.ApiPath, id))
}
func (s *EtcdStore) GetApis(service string, handler func(item *meta.Api)) error {
	return s.gets(s.skey(service, s.ApiPath, ""), func() meta.Serializable {
		return &meta.Api{
			Headers: &meta.ApiHeaders{},
			Cookies: &meta.ApiCookies{}}
	}, func(sb meta.Serializable) {
		handler(sb.(*meta.Api))
	})
}
func (s *EtcdStore) GetApi(service, id string) (*meta.Api, error) {
	resp, err := s.get(s.skey(service, s.ApiPath, id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if resp.Count <= 0 {
		return nil, nil
	}
	kv := resp.Kvs[0]
	m := &meta.Api{Id: id}
	err = m.Unmarshal(kv.Value)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *EtcdStore) PutServer(service string, server *meta.Server) error {
	if len(server.Addr) <= 0 {
		return errors.New("server.Addr shoud not be empty")
	}
	server.InitId()
	data, err := server.Marshal()
	if err != nil {
		return err
	}
	return s.put(s.skey(service, s.ServerPath, server.Id), reflectx.BytesToString(data))
}
func (s *EtcdStore) RemoveServer(service, id string) error {
	return s.delete(s.skey(service, s.ServerPath, id))
}
func (s *EtcdStore) GetServers(service string, handler func(item *meta.Server)) error {
	return s.gets(s.skey(service, s.ServerPath, ""), func() meta.Serializable { return &meta.Server{} }, func(sb meta.Serializable) {
		handler(sb.(*meta.Server))
	})
}
func (s *EtcdStore) GetServer(service, id string) (*meta.Server, error) {
	resp, err := s.get(s.skey(service, s.ServerPath, id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if resp.Count <= 0 {
		return nil, nil
	}
	kv := resp.Kvs[0]
	m := &meta.Server{Id: id}
	err = m.Unmarshal(kv.Value)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *EtcdStore) put(key, value string, opts ...clientv3.OpOption) error {
	_, err := s.txn().Then(clientv3.OpPut(key, value, opts...)).Commit()
	return err
}

func (s *EtcdStore) putWithTTL(key, value string, ttl time.Duration) error {
	lease := clientv3.NewLease(s.client)
	defer lease.Close()
	ctx, cancel := context.WithTimeout(s.client.Ctx(), DefaultRequestTimeout)
	lgResp, err := lease.Grant(ctx, int64(ttl/time.Second))
	cancel()
	if err != nil {
		return err
	}
	_, err = s.txn().Then(clientv3.OpPut(key, value, clientv3.WithLease(lgResp.ID))).Commit()
	return err
}

func (s *EtcdStore) putWithKeepAlive(key, value string, ttl time.Duration) error {
	lease := clientv3.NewLease(s.client)
	defer lease.Close()
	ctx, cancel := context.WithTimeout(s.client.Ctx(), DefaultRequestTimeout)
	lgResp, err := lease.Grant(ctx, int64(ttl/time.Second))
	cancel()
	if err != nil {
		return err
	}
	_, err = s.client.KeepAlive(s.client.Ctx(), lgResp.ID)
	if err != nil {
		return err
	}
	_, err = s.txn().Then(clientv3.OpPut(key, value, clientv3.WithLease(lgResp.ID))).Commit()
	return err
}

func (s *EtcdStore) delete(key string, opts ...clientv3.OpOption) error {
	_, err := s.txn().Then(clientv3.OpDelete(key, opts...)).Commit()
	return err
}

func (s *EtcdStore) get(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(s.client.Ctx(), DefaultRequestTimeout)
	defer cancel()
	return clientv3.NewKV(s.client).Get(ctx, key, opts...)
}

func (s *EtcdStore) gets(prefix string, factory func() meta.Serializable, handler func(sb meta.Serializable)) error {
	var limit int64 = 1
	withRange := clientv3.WithRange(s.nextKey([]byte(prefix)))
	withLimit := clientv3.WithLimit(limit)
	for {
		resp, err := s.get(prefix, withRange, withLimit)
		if err != nil {
			return err
		}
		var nextKey []byte
		for _, item := range resp.Kvs {
			value := factory()
			err := value.Unmarshal(item.Value)
			if err != nil {
				return err
			}
			if len(value.GetId()) <= 0 {
				continue
			}
			if value.Valid() != nil {
				continue
			}
			handler(value)
			nextKey = item.Key
		}
		if len(resp.Kvs) < int(limit) {
			return nil
		}
		if len(nextKey) <= 0 {
			return nil
		}
		prefix = s.nextKey(nextKey)
	}
}

func (s *EtcdStore) key(prefix, id string) string {
	return prefix + id
}

func (s *EtcdStore) skey(service, part, id string) string {
	return s.ServicePrefix + service + part + id
}

func (s *EtcdStore) nextKey(key []byte) string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] < 0xFF {
			key[i]++
			return string(key[:i+1])
		}
	}
	return "\x00"
}
