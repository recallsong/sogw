package etcd

import (
	"strings"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
)

var watchPathHandlers = map[string]func(ln meta.EventListener, op meta.Operation, key string, kv *mvccpb.KeyValue){
	"hosts": func(ln meta.EventListener, op meta.Operation, key string, kv *mvccpb.KeyValue) {
		m := &meta.Host{Id: key}
		err := m.Unmarshal(kv.Value)
		if err != nil || (op != meta.OperationDelete && m.Valid() != nil) {
			log.Errorf("[etcd] [watch] recv invalid host = %s , %s", key, op)
			return
		}
		log.Infof("[etcd] [watch] host = %s , %s", key, op)
		ln.RecvHost(op, m)
	},
	"auths": func(ln meta.EventListener, op meta.Operation, key string, kv *mvccpb.KeyValue) {
		m := &meta.Auth{Id: key}
		err := m.Unmarshal(kv.Value)
		if err != nil || (op != meta.OperationDelete && m.Valid() != nil) {
			log.Errorf("[etcd] [watch] recv invalid auth = %s , %s", key, op)
			return
		}
		log.Infof("[etcd] [watch] auth = %s , %s", key, op)
		ln.RecvAuth(op, m)
	},
	"routes": func(ln meta.EventListener, op meta.Operation, key string, kv *mvccpb.KeyValue) {
		m := &meta.Route{Id: key}
		err := m.Unmarshal(kv.Value)
		if err != nil || (op != meta.OperationDelete && m.Valid() != nil) {
			log.Errorf("[etcd] [watch] recv invalid route id = %s , %s", key, op)
			return
		}
		log.Infof("[etcd] [watch] route = %s , %s", key, op)
		ln.RecvRoute(op, m)
	},
	"services": func(ln meta.EventListener, op meta.Operation, key string, kv *mvccpb.KeyValue) {
		m := &meta.Service{Id: key}
		err := m.Unmarshal(kv.Value)
		if err != nil || (op != meta.OperationDelete && m.Valid() != nil) {
			log.Errorf("[etcd] [watch] recv invalid service = %s , %s", key, op)
			return
		}
		log.Infof("[etcd] [watch] service = %s , name = %s , %s", key, m.Name, op)
		ln.RecvService(op, m)
	},
	"space": func(ln meta.EventListener, op meta.Operation, key string, kv *mvccpb.KeyValue) {
		idx := strings.IndexByte(key, '/')
		if idx < 0 {
			log.Errorf("[etcd] [watch] recv invalid space key = %s , %s", key, op)
			return
		}
		service := key[:idx]
		key = key[idx+1:]
		part := key
		idx = strings.IndexByte(key, '/')
		if idx >= 0 {
			part = key[:idx]
			key = key[idx+1:]
		}
		h, ok := watchSpacePathHandlers[part]
		if !ok {
			log.Debugf("recv invalid part = %s in service = %s , %s", part, service, op)
			return
		}
		h(ln, service, op, key, kv)
	},
}

var watchSpacePathHandlers = map[string]func(ln meta.EventListener, service string, op meta.Operation, key string, kv *mvccpb.KeyValue){
	"svrs": func(ln meta.EventListener, service string, op meta.Operation, key string, kv *mvccpb.KeyValue) {
		m := &meta.Server{Id: key}
		err := m.Unmarshal(kv.Value)
		if err != nil {
			log.Errorf("[etcd] [watch] recv invalid server = %s in service = %s , %s", key, service, op)
			return
		}
		log.Infof("[etcd] [watch] service = %s , server = %s , %s", service, key, op)
		ln.RecvServer(op, service, m)
	},
	"apis": func(ln meta.EventListener, service string, op meta.Operation, key string, kv *mvccpb.KeyValue) {
		m := &meta.Api{
			Id:      key,
			Headers: &meta.ApiHeaders{},
			Cookies: &meta.ApiCookies{},
		}
		err := m.Unmarshal(kv.Value)
		if err != nil || (op != meta.OperationDelete && m.Valid() != nil) {
			log.Errorf("[etcd] [watch] recv invalid api = %s in service = %s , %s", key, service, op)
			return
		}
		log.Infof("[etcd] [watch] service = %s , api = %s , %s", service, key, op)
		ln.RecvApi(op, service, m)
	},
	"cfg": func(ln meta.EventListener, service string, op meta.Operation, key string, kv *mvccpb.KeyValue) {
		m := &meta.ServiceConfig{
			Id: key,
		}
		err := m.Unmarshal(kv.Value)
		if err != nil || (op != meta.OperationDelete && m.Valid() != nil) {
			log.Errorf("[etcd] [watch] recv invalid config %s in service = %s , %s", key, service, op)
			return
		}
		log.Infof("[etcd] [watch] service %s config , %s )", service, op)
		ln.RecvServiceConfig(op, service, m)
	},
}

func (s *EtcdStore) Watch(ln meta.EventListener, stopCh <-chan struct{}, waitStop *sync.WaitGroup) error {
	go func(prefix string, ln meta.EventListener) {
		waitStop.Add(1)
		defer waitStop.Done()
		log.Infof("[etcd] watch events (%s*)", prefix)
		watcher := clientv3.NewWatcher(s.client)
		ctx := s.client.Ctx()
		defer watcher.Close()
		for {
			rch := watcher.Watch(ctx, prefix, clientv3.WithPrefix())
			for {
				select {
				case <-stopCh:
					return
				case resp := <-rch:
					if resp.Canceled {
						return
					}
					for _, ev := range resp.Events {
						op := meta.OperationNone
						switch ev.Type {
						case mvccpb.DELETE:
							op = meta.OperationDelete
						case mvccpb.PUT:
							if ev.IsCreate() {
								op = meta.OperationCreate
							} else if ev.IsModify() {
								op = meta.OperationUpdate
							}
						}
						key := reflectx.BytesToString(ev.Kv.Key)
						if !strings.HasPrefix(key, prefix) {
							log.Debugf("recv invalid a key = %s", key)
							continue
						}
						key = strings.Replace(key, prefix, "", 1)
						idx := strings.IndexByte(key, '/')
						if idx < 0 {
							log.Debugf("recv invalid a key = %s%s", prefix, key)
							continue
						}
						typ := key[:idx]
						key = key[idx+1:]
						h, ok := watchPathHandlers[typ]
						if !ok {
							log.Debugf("recv invalid key = %s ,type = %s", key, typ)
							continue
						}
						h(ln, op, key, ev.Kv)
					}
				}
			}
		}
	}(s.Prefix, ln)
	return nil
}
