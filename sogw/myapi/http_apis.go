package myapi

import (
	"errors"
	"net/http"

	"github.com/labstack/echo"
	"github.com/recallsong/go-utils/conv"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
)

func (s *ApiServer) initHttpApis() error {
	svr := s.httpSvr.Echo
	svr.POST("/hosts", s.putHost)
	svr.DELETE("/hosts/:id", s.removeHost)
	svr.GET("/hosts/:id", s.getHost)
	svr.GET("/hosts", s.getHosts)

	svr.POST("/auths", s.putAuth)
	svr.DELETE("/auths/:id", s.removeAuth)
	svr.GET("/auths/:id", s.getAuth)
	svr.GET("/auths", s.getAuths)

	svr.POST("/routes", s.putRoute)
	svr.DELETE("/routes/:id", s.removeRoute)
	svr.GET("/routes/:id", s.getRoute)
	svr.GET("/routes", s.getRoutes)

	svr.POST("/services", s.putService)
	svr.DELETE("/services/:id", s.removeService)
	svr.GET("/services/:id", s.getService)
	svr.GET("/services", s.getServices)

	svr.POST("/services/:sid/cfg", s.putServiceCfg)
	svr.DELETE("/services/:sid/cfg", s.removeServiceCfg)
	svr.GET("/services/:sid/cfg", s.getServiceCfg)

	svr.POST("/services/:sid/svrs", s.putServer)
	svr.DELETE("/services/:sid/svrs/:id", s.removeServer)
	svr.GET("/services/:sid/svrs/:id", s.getServer)
	svr.GET("/services/:sid/svrs", s.getServers)

	svr.POST("/services/:sid/apis", s.putApi)
	svr.DELETE("/services/:sid/apis/:id", s.removeApi)
	svr.GET("/services/:sid/apis/:id", s.getApi)
	svr.GET("/services/:sid/apis", s.getApis)

	if s.cfg.HttpAddr == "" {
		err := errors.New("http addr should not be empty")
		log.Error("[apisvr] ", err)
		return err
	} else {
		_, svr := s.httpSvr.GetHttpServer(s.cfg.HttpAddr)
		err := s.serGrp.Put(s.cfg.HttpAddr, svr)
		if err != nil {
			log.Error("[apisvr] ", err)
			return err
		}
	}
	return nil
}

func (s *ApiServer) putHost(ctx echo.Context) error {
	data := &meta.Host{}
	err := s.ReadJSON(ctx, &data)
	if err != nil {
		return nil
	}
	if data.Id, err = "-", data.Valid(); err != nil {
		s.WriteError(ctx, http.StatusBadRequest, err.Error())
		return nil
	}
	data.Id = ""
	err = s.store.PutHost(data)
	if err != nil {
		log.Error("[apisvr] fail to put host ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to put host")
		return nil
	}
	return nil
}
func (s *ApiServer) removeHost(ctx echo.Context) error {
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "host id should not be empty")
		return nil
	}
	err := s.store.RemoveHost(id)
	if err != nil {
		log.Error("[apisvr] fail to remove host ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to remove host")
		return nil
	}
	return nil
}
func (s *ApiServer) getHost(ctx echo.Context) error {
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "host id should not be empty")
		return nil
	}
	data, err := s.store.GetHost(id)
	if err != nil {
		log.Errorf("[apisvr] fail to get host %s , %s", id, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get host")
		return nil
	}
	s.WriteData(ctx, data)
	return nil
}
func (s *ApiServer) getHosts(ctx echo.Context) error {
	var hosts []*meta.Host
	err := s.store.GetHosts(func(item *meta.Host) {
		hosts = append(hosts, item)
	})
	if err != nil {
		log.Error("[apisvr] fail to get hosts ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get hosts")
		return nil
	}
	s.WriteData(ctx, hosts)
	return nil
}

func (s *ApiServer) putAuth(ctx echo.Context) error {
	data := &meta.Auth{}
	err := s.ReadJSON(ctx, &data)
	if err != nil {
		return nil
	}
	if data.Id, err = "-", data.Valid(); err != nil {
		s.WriteError(ctx, http.StatusBadRequest, err.Error())
		return nil
	}
	data.Id = ""
	err = s.store.PutAuth(data)
	if err != nil {
		log.Error("[apisvr] fail to put auth ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to put auth")
		return nil
	}
	return nil
}
func (s *ApiServer) removeAuth(ctx echo.Context) error {
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "auth id should not be empty")
		return nil
	}
	err := s.store.RemoveAuth(id)
	if err != nil {
		log.Error("[apisvr] fail to remove auth ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to remove auth")
		return nil
	}
	return nil
}
func (s *ApiServer) getAuth(ctx echo.Context) error {
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "auth id should not be empty")
		return nil
	}
	data, err := s.store.GetAuth(id)
	if err != nil {
		log.Errorf("[apisvr] fail to get auth %s , %s", id, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get auth")
		return nil
	}
	s.WriteData(ctx, data)
	return nil
}
func (s *ApiServer) getAuths(ctx echo.Context) error {
	var auths []*meta.Auth
	err := s.store.GetAuths(func(item *meta.Auth) {
		auths = append(auths, item)
	})
	if err != nil {
		log.Error("[apisvr] fail to get auths ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get auths")
		return nil
	}
	s.WriteData(ctx, auths)
	return nil
}

func (s *ApiServer) putRoute(ctx echo.Context) error {
	data := &meta.Route{}
	err := s.ReadJSON(ctx, &data)
	if err != nil {
		return nil
	}
	if data.Id, err = "-", data.Valid(); err != nil {
		s.WriteError(ctx, http.StatusBadRequest, err.Error())
		return nil
	}
	data.Id = ""
	err = s.store.PutRoute(data)
	if err != nil {
		log.Error("[apisvr] fail to put route ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to put route")
		return nil
	}
	return nil
}
func (s *ApiServer) removeRoute(ctx echo.Context) error {
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "route id should not be empty")
		return nil
	}
	err := s.store.RemoveRoute(id)
	if err != nil {
		log.Error("[apisvr] fail to remove route ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to remove route")
		return nil
	}
	return nil
}
func (s *ApiServer) getRoute(ctx echo.Context) error {
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "route id should not be empty")
		return nil
	}
	data, err := s.store.GetRoute(id)
	if err != nil {
		log.Errorf("[apisvr] fail to get route %s , %s", id, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get route")
		return nil
	}
	s.WriteData(ctx, data)
	return nil
}
func (s *ApiServer) getRoutes(ctx echo.Context) error {
	var routes []*meta.Route
	err := s.store.GetRoutes(func(item *meta.Route) {
		routes = append(routes, item)
	})
	if err != nil {
		log.Error("[apisvr] fail to get routes ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get routes")
		return nil
	}
	s.WriteData(ctx, routes)
	return nil
}

func (s *ApiServer) putService(ctx echo.Context) error {
	data := &meta.Service{}
	err := s.ReadJSON(ctx, &data)
	if err != nil {
		return nil
	}
	if data.Id, err = "-", data.Valid(); err != nil {
		s.WriteError(ctx, http.StatusBadRequest, err.Error())
		return nil
	}
	data.Id = ""
	err = s.store.PutService(data)
	if err != nil {
		log.Error("[apisvr] fail to put service ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to put service")
		return nil
	}
	return nil
}
func (s *ApiServer) removeService(ctx echo.Context) error {
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	err := s.store.RemoveService(id, conv.ParseBool(ctx.QueryParam("cascade"), false))
	if err != nil {
		log.Error("[apisvr] fail to remove service ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to remove service")
		return nil
	}
	return nil
}
func (s *ApiServer) getService(ctx echo.Context) error {
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	data, err := s.store.GetService(id)
	if err != nil {
		log.Errorf("[apisvr] fail to get service %s , %s", id, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get service")
		return nil
	}
	s.WriteData(ctx, data)
	return nil
}
func (s *ApiServer) getServices(ctx echo.Context) error {
	var services []*meta.Service
	err := s.store.GetServices(func(item *meta.Service) {
		services = append(services, item)
	})
	if err != nil {
		log.Error("[apisvr] fail to get services ", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get services")
		return nil
	}
	s.WriteData(ctx, services)
	return nil
}

func (s *ApiServer) putServiceCfg(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	data := &meta.ServiceConfig{}
	err := s.ReadJSON(ctx, &data)
	if err != nil {
		return nil
	}

	if data.Id, err = "-", data.Valid(); err != nil {
		s.WriteError(ctx, http.StatusBadRequest, err.Error())
		return nil
	}
	data.Id = ""
	err = s.store.PutServiceCfg(sid, data)
	if err != nil {
		log.Error("[apisvr] fail to put service config", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to put service config")
		return nil
	}
	return nil
}
func (s *ApiServer) removeServiceCfg(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	err := s.store.RemoveServiceCfg(sid)
	if err != nil {
		log.Errorf("[apisvr] fail to remove service %s config, %s", sid, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to remove service config")
		return nil
	}
	return nil
}
func (s *ApiServer) getServiceCfg(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	data, err := s.store.GetServiceCfg(sid)
	if err != nil {
		log.Errorf("[apisvr] fail to get service %s config, %s", sid, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get service")
		return nil
	}
	s.WriteData(ctx, data)
	return nil
}

func (s *ApiServer) putServer(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	data := &meta.Server{}
	err := s.ReadJSON(ctx, &data)
	if err != nil {
		return nil
	}

	if data.Id, err = "-", data.Valid(); err != nil {
		s.WriteError(ctx, http.StatusBadRequest, err.Error())
		return nil
	}
	data.Id = ""
	err = s.store.PutServer(sid, data)
	if err != nil {
		log.Errorf("[apisvr] fail to put server in service %s, %s", sid, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to put server")
		return nil
	}
	return nil
}
func (s *ApiServer) removeServer(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "server id should not be empty")
		return nil
	}
	err := s.store.RemoveServer(sid, id)
	if err != nil {
		log.Errorf("[apisvr] fail to remove server %s in service %s, %s", id, sid, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to remove server")
		return nil
	}
	return nil
}
func (s *ApiServer) getServer(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "server id should not be empty")
		return nil
	}
	data, err := s.store.GetServer(sid, id)
	if err != nil {
		log.Errorf("[apisvr] fail to get server %s in service %s, %s", id, sid, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get server")
		return nil
	}
	s.WriteData(ctx, data)
	return nil
}
func (s *ApiServer) getServers(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	var svrs []*meta.Server
	err := s.store.GetServers(sid, func(item *meta.Server) {
		svrs = append(svrs, item)
	})
	if err != nil {
		log.Errorf("[apisvr] fail to get servers in service %s, %s", sid, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get servers")
		return nil
	}
	s.WriteData(ctx, svrs)
	return nil
}

func (s *ApiServer) putApi(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	data := &meta.Api{}
	err := s.ReadJSON(ctx, &data)
	if err != nil {
		return nil
	}

	if data.Id, err = "-", data.Valid(); err != nil {
		s.WriteError(ctx, http.StatusBadRequest, err.Error())
		return nil
	}
	data.Id = ""
	err = s.store.PutApi(sid, data)
	if err != nil {
		log.Error("[apisvr] fail to put api", err)
		s.WriteError(ctx, http.StatusInternalServerError, "fail to put api")
		return nil
	}
	return nil
}
func (s *ApiServer) removeApi(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "api id should not be empty")
		return nil
	}
	err := s.store.RemoveApi(sid, id)
	if err != nil {
		log.Errorf("[apisvr] fail to remove api %s in service %s, %s", id, sid, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to remove api")
		return nil
	}
	return nil
}
func (s *ApiServer) getApi(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	id := ctx.Param("id")
	if len(id) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "api id should not be empty")
		return nil
	}
	data, err := s.store.GetApi(sid, id)
	if err != nil {
		log.Errorf("[apisvr] fail to get api %s in service %s, %s", id, sid, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get api")
		return nil
	}
	s.WriteData(ctx, data)
	return nil
}
func (s *ApiServer) getApis(ctx echo.Context) error {
	sid := ctx.Param("sid")
	if len(sid) <= 0 {
		s.WriteError(ctx, http.StatusBadRequest, "service id should not be empty")
		return nil
	}
	var apis []*meta.Api
	err := s.store.GetApis(sid, func(item *meta.Api) {
		apis = append(apis, item)
	})
	if err != nil {
		log.Errorf("[apisvr] fail to get apis in service %s, %s", sid, err.Error())
		s.WriteError(ctx, http.StatusInternalServerError, "fail to get apis")
		return nil
	}
	s.WriteData(ctx, apis)
	return nil
}
