package core

import (
	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type Server struct {
	_              lang.NoCopy
	Meta           *meta.Server
	client         *fasthttp.Client
	checkFailTimes int64
	checkSum       int64
	// healthCk       *HealthChecker
}

func NewServer(m *meta.Server) *Server {
	svr := &Server{
		Meta:   m,
		client: httpClient,
	}
	return svr
}

func (s *Server) Forward(freq *fasthttp.Request, fresp *fasthttp.Response) error {
	freq.SetHost(s.Meta.Addr)
	if len(s.Meta.Host) > 0 {
		freq.Header.SetHost(s.Meta.Host)
	}
	err := s.client.Do(freq, fresp)
	if err != nil {
		log.Errorf("[server] forward %s -> %s", reflectx.BytesToString(freq.URI().FullURI()), err.Error())
	}
	return err
}

func (s *Server) Check() error {
	/*
		freq.SetHost(s.Meta.Addr)
		if len(s.Meta.Host) > 0 {
			freq.Header.SetHost(s.Meta.Host)
		}
		err := client.Do(freq, fresp)
		if err != nil {
			log.Errorf("[server] forward %s -> %s", reflectx.BytesToString(freq.URI().FullURI()), err.Error())
		}
		return err*/
	return nil
}

func (s *Server) Close() error {
	// if s.healthCk != nil {
	// 	s.healthCk.Stop()
	// }
	return nil
}
