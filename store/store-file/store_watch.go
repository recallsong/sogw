package file

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
)

func (fs *FileStore) Watch(ln meta.EventListener, stopCh <-chan struct{}, waitStop *sync.WaitGroup) error {
	fs.viper.WatchConfig()
	fs.viper.OnConfigChange(func(event fsnotify.Event) {
		if event.Op&fsnotify.Remove == fsnotify.Remove {
			log.Debugf("[file] [watch] file %s removed", event.Name)
		} else if event.Op&fsnotify.Rename == fsnotify.Rename {
			log.Debugf("[file] [watch] file %s renamed", event.Name)
		} else if (event.Op&fsnotify.Write == fsnotify.Write) ||
			(event.Op&fsnotify.Create == fsnotify.Create) {
			cache, err := fs.load()
			if err != nil {
				log.Errorf("[file] [whath] file load error %s", err.Error())
				return
			}
			fs.sendEvents(ln, cache)
		}
	})
	return nil
}
func (fs *FileStore) sendEvents(ln meta.EventListener, cache *storeCache) {
	fs.look.Lock()
	defer fs.look.Unlock()
	fs.sendHostsEvents(ln, fs.cache.Hosts, cache.Hosts)
	fs.sendAuthsEvents(ln, fs.cache.Auths, cache.Auths)
	fs.sendRoutesEvents(ln, fs.cache.Routes, cache.Routes)
	fs.sendServicesEvents(ln, fs.cache.Services, cache.Services)
	fs.cache = cache
}
func (fs *FileStore) sendHostsEvents(ln meta.EventListener, old, new map[string]*meta.Host) {
	if old == nil {
		if new != nil {
			for _, item := range new {
				log.Infof("[file] [watch] host = %s , %s", item.Id, meta.OperationCreate)
				ln.RecvHost(meta.OperationCreate, item.Copy())
			}
		}
	} else if new == nil {
		for _, item := range old {
			log.Infof("[file] [watch] host = %s , %s", item.Id, meta.OperationDelete)
			ln.RecvHost(meta.OperationDelete, item.Copy())
		}
	} else {
		for id, item := range old {
			if v, ok := new[id]; ok {
				log.Infof("[file] [watch] host = %s , %s", v.Id, meta.OperationUpdate)
				ln.RecvHost(meta.OperationUpdate, v.Copy())
			} else {
				log.Infof("[file] [watch] host = %s , %s", item.Id, meta.OperationDelete)
				ln.RecvHost(meta.OperationDelete, item.Copy())
			}
		}
		for id, item := range new {
			if _, ok := old[id]; !ok {
				log.Infof("[file] [watch] host = %s , %s", item.Id, meta.OperationCreate)
				ln.RecvHost(meta.OperationCreate, item.Copy())
			}
		}
	}
}
func (fs *FileStore) sendAuthsEvents(ln meta.EventListener, old, new map[string]*meta.Auth) {
	if old == nil {
		if new != nil {
			for _, item := range new {
				log.Infof("[file] [watch] auth = %s , %s", item.Id, meta.OperationCreate)
				ln.RecvAuth(meta.OperationCreate, item.Copy())
			}
		}
	} else if new == nil {
		for _, item := range old {
			log.Infof("[file] [watch] auth = %s , %s", item.Id, meta.OperationDelete)
			ln.RecvAuth(meta.OperationDelete, item.Copy())
		}
	} else {
		for id, item := range old {
			if v, ok := new[id]; ok {
				log.Infof("[file] [watch] auth = %s , %s", v.Id, meta.OperationUpdate)
				ln.RecvAuth(meta.OperationUpdate, v.Copy())
			} else {
				log.Infof("[file] [watch] auth = %s , %s", item.Id, meta.OperationDelete)
				ln.RecvAuth(meta.OperationDelete, item.Copy())
			}
		}
		for id, item := range new {
			if _, ok := old[id]; !ok {
				log.Infof("[file] [watch] auth = %s , %s", item.Id, meta.OperationCreate)
				ln.RecvAuth(meta.OperationCreate, item.Copy())
			}
		}
	}
}
func (fs *FileStore) sendRoutesEvents(ln meta.EventListener, old, new map[string]*meta.Route) {
	if old == nil {
		if new != nil {
			for _, item := range new {
				log.Infof("[file] [watch] route = %s , %s", item.Id, meta.OperationCreate)
				ln.RecvRoute(meta.OperationCreate, item.Copy())
			}
		}
	} else if new == nil {
		for _, item := range old {
			log.Infof("[file] [watch] route = %s , %s", item.Id, meta.OperationDelete)
			ln.RecvRoute(meta.OperationDelete, item.Copy())
		}
	} else {
		for id, item := range old {
			if v, ok := new[id]; ok {
				log.Infof("[file] [watch] route = %s , %s", v.Id, meta.OperationUpdate)
				ln.RecvRoute(meta.OperationUpdate, v.Copy())
			} else {
				log.Infof("[file] [watch] route = %s , %s", item.Id, meta.OperationDelete)
				ln.RecvRoute(meta.OperationDelete, item.Copy())
			}
		}
		for id, item := range new {
			if _, ok := old[id]; !ok {
				log.Infof("[file] [watch] route = %s , %s", item.Id, meta.OperationCreate)
				ln.RecvRoute(meta.OperationCreate, item.Copy())
			}
		}
	}
}

func (fs *FileStore) sendServicesEvents(ln meta.EventListener, old, new map[string]*serviceCache) {
	for service, item := range old {
		if v, ok := new[service]; ok {
			log.Infof("[file] [watch] service = %s , %s", v.Service.Id, meta.OperationUpdate)
			ln.RecvService(meta.OperationUpdate, v.Service.Copy())
			fs.sendServiceCfgEvents(ln, service, item.Cfg, v.Cfg)
			fs.sendApisEvents(ln, service, item.Apis, v.Apis)
			fs.sendSvrsEvents(ln, service, item.Svrs, v.Svrs)
		} else {
			log.Infof("[file] [watch] service = %s , %s", item.Service.Id, meta.OperationDelete)
			fs.sendServiceCfgEvents(ln, service, item.Cfg, nil)
			fs.sendApisEvents(ln, service, item.Apis, nil)
			fs.sendSvrsEvents(ln, service, item.Svrs, nil)
			ln.RecvService(meta.OperationDelete, item.Service.Copy())
		}
	}
	for service, item := range new {
		if _, ok := old[service]; !ok {
			log.Infof("[file] [watch] service = %s , %s", item.Service.Id, meta.OperationCreate)
			ln.RecvService(meta.OperationCreate, item.Service.Copy())
			fs.sendServiceCfgEvents(ln, service, nil, item.Cfg)
			fs.sendApisEvents(ln, service, nil, item.Apis)
			fs.sendSvrsEvents(ln, service, nil, item.Svrs)
		}
	}
}
func (fs *FileStore) sendApisEvents(ln meta.EventListener, service string, old, new map[string]*meta.Api) {
	if old == nil {
		if new != nil {
			for _, item := range new {
				log.Infof("[file] [watch] api = %s in service = %s , %s", item.Id, service, meta.OperationCreate)
				ln.RecvApi(meta.OperationCreate, service, item.Copy())
			}
		}
	} else if new == nil {
		for _, item := range old {
			log.Infof("[file] [watch] api = %s in service = %s , %s", item.Id, service, meta.OperationDelete)
			ln.RecvApi(meta.OperationDelete, service, item.Copy())
		}
	} else {
		for id, item := range old {
			if v, ok := new[id]; ok {
				log.Infof("[file] [watch] api = %s in service = %s , %s", v.Id, service, meta.OperationUpdate)
				ln.RecvApi(meta.OperationUpdate, service, v.Copy())
			} else {
				log.Infof("[file] [watch] api = %s in service = %s , %s", item.Id, service, meta.OperationDelete)
				ln.RecvApi(meta.OperationDelete, service, item.Copy())
			}
		}
		for id, item := range new {
			if _, ok := old[id]; !ok {
				log.Infof("[file] [watch] api = %s in service = %s , %s", item.Id, service, meta.OperationCreate)
				ln.RecvApi(meta.OperationCreate, service, item.Copy())
			}
		}
	}
}
func (fs *FileStore) sendSvrsEvents(ln meta.EventListener, service string, old, new map[string]*meta.Server) {
	if old == nil {
		if new != nil {
			for _, item := range new {
				log.Infof("[file] [watch] svr = %s in service = %s , %s", item.Id, service, meta.OperationCreate)
				ln.RecvServer(meta.OperationCreate, service, item.Copy())
			}
		}
	} else if new == nil {
		for _, item := range old {
			log.Infof("[file] [watch] svr = %s in service = %s , %s", item.Id, service, meta.OperationDelete)
			ln.RecvServer(meta.OperationDelete, service, item.Copy())
		}
	} else {
		for id, item := range old {
			if v, ok := new[id]; ok {
				log.Infof("[file] [watch] svr = %s in service = %s , %s", item.Id, service, meta.OperationUpdate)
				ln.RecvServer(meta.OperationUpdate, service, v.Copy())
			} else {
				log.Infof("[file] [watch] svr = %s in service = %s , %s", item.Id, service, meta.OperationDelete)
				ln.RecvServer(meta.OperationDelete, service, item.Copy())
			}
		}
		for id, item := range new {
			if _, ok := old[id]; !ok {
				log.Infof("[file] [watch] svr = %s in service = %s , %s", item.Id, service, meta.OperationCreate)
				ln.RecvServer(meta.OperationCreate, service, item.Copy())
			}
		}
	}
}
func (fs *FileStore) sendServiceCfgEvents(ln meta.EventListener, service string, old, new *meta.ServiceConfig) {
	if old == nil {
		if new != nil {
			log.Infof("[file] [watch] cfg = %s in service = %s , %s", new.Id, service, meta.OperationCreate)
			ln.RecvServiceConfig(meta.OperationCreate, service, new.Copy())
		}
	} else if new == nil {
		log.Infof("[file] [watch] cfg = %s in service = %s , %s", old.Id, service, meta.OperationDelete)
		ln.RecvServiceConfig(meta.OperationDelete, service, old.Copy())
	} else {
		log.Infof("[file] [watch] cfg = %s in service = %s , %s", new.Id, service, meta.OperationUpdate)
		ln.RecvServiceConfig(meta.OperationUpdate, service, new.Copy())
	}
}
