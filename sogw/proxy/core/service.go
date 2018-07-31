package core

import (
	"github.com/recallsong/go-utils/errorx"
	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	_          lang.NoCopy
	Meta       *meta.Service
	Config     *meta.ServiceConfig
	Apis       map[string]*Api
	Servers    map[string]*Server
	ServerList []*Server
	LB         LoadBalance
}

func NewService(m *meta.Service) *Service {
	return &Service{
		Meta:    m,
		Apis:    make(map[string]*Api),
		Servers: make(map[string]*Server),
	}
}

func (s *Service) Init(m *meta.ServiceConfig) *Server {
	if m == nil {
		m = NewServiceConfig()
	}
	s.Config = m
	lb_new, ok := loadBalanceGetter[m.LoadBlance]
	if !ok {
		log.Warnf("[service] load balance %v not found, use RoundRobin", m.LoadBlance)
		lb_new = NewRoundRobinLB
	}
	s.LB = lb_new()
	return nil
}

func (s *Service) Close() error {
	errs := errorx.Errors{}
	for _, svr := range s.Servers {
		err := svr.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs.MaybeUnwrap()
}

func NewServiceConfig() *meta.ServiceConfig {
	return &meta.ServiceConfig{
		Status:     meta.Status_Open,
		LoadBlance: meta.LoadBalance_RoundRobin,
	}
}
