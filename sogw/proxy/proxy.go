package proxy

import (
	"fmt"
	"os"
	"strings"

	"github.com/recallsong/go-utils/ioutil"
	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/go-utils/net/fasthttpx"
	"github.com/recallsong/go-utils/net/servegrp"
	"github.com/recallsong/sogw/sogw/proxy/core"
	"github.com/recallsong/sogw/sogw/proxy/filters"
	"github.com/recallsong/sogw/sogw/proxy/jobs"
	"github.com/recallsong/sogw/sogw/proxy/jobs/healthchecker"
	log "github.com/sirupsen/logrus"
)

type HttpProxy struct {
	_         lang.NoCopy
	cfg       *Config
	svrGrp    *servegrp.ServeGroup
	waitClose *servegrp.WaitClose
	filters   *filters.FilterManager
	jobs      *jobs.JobManager
	rtCtx     *core.RuntimeContext
}

func New() *HttpProxy {
	return &HttpProxy{
		svrGrp:    servegrp.NewServeGroup(),
		waitClose: servegrp.NewWaitClose(),
	}
}

func (p *HttpProxy) Init(c *Config) error {
	p.cfg = c
	if err := p.initStore(&c.Store); err != nil {
		return err
	}
	if err := p.initJobs(p.cfg.Jobs); err != nil {
		return err
	}
	if err := p.initFilters(p.cfg.Filters); err != nil {
		return err
	}
	if err := p.initServers(p.cfg); err != nil {
		return err
	}
	return nil
}

func (p *HttpProxy) initJobs(cfg map[string]interface{}) error {
	// jobs.RegisterJobFactory() , TODO
	p.jobs = jobs.NewJobManager()
	p.jobs.Add("healthchecker", healthchecker.JobFactory())
	err := p.jobs.SetupAll(cfg)
	if err != nil {
		log.Errorf("[proxy] set jobs error : %v", err)
		return err
	}
	p.jobs.Run(p.rtCtx, p.waitClose.CloseCh, &p.waitClose.WaitGroup)
	return nil
}

func (p *HttpProxy) initServers(c *Config) error {
	if c.Addr != "" {
		svr := fasthttpx.NewTcpServer(p.Handler)
		err := svr.Listen(c.Addr)
		if err != nil {
			log.Errorf("[proxy] %v", err)
			return err
		}
		err = p.svrGrp.Put(c.Addr, svr)
		if err != nil {
			log.Errorf("[proxy] %v", err)
			return err
		}
		log.Infof("[proxy] listen tcp [ %s ] ok", c.Addr)
	}
	if c.TLSAddr != "" {
		parts := strings.Split(c.TLSAddr, ",")
		if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			err := fmt.Errorf("tls address format is invalid")
			log.Error("[proxy] ", err)
			return err
		}
		svr := fasthttpx.NewTLSTcpServer(p.Handler, parts[1], parts[2])
		err := svr.Listen(parts[0])
		if err != nil {
			log.Errorf("[proxy] %v", err)
			return err
		}
		err = p.svrGrp.Put(parts[0], svr)
		if err != nil {
			log.Errorf("[proxy] %v", err)
			return err
		}
		log.Infof("[proxy] listen tcp (tls) [ %s ] ok", parts[0])
	}
	if c.UnixAddr != "" {
		svr := fasthttpx.NewUnixServer(p.Handler)
		err := svr.Listen(c.UnixAddr)
		if err != nil {
			log.Errorf("[proxy] %v", err)
			return err
		}
		err = p.svrGrp.Put(c.UnixAddr, svr)
		if err != nil {
			log.Errorf("[proxy] %v", err)
			return err
		}
		log.Infof("[proxy] listen unix socket [ %s ] ok", c.UnixAddr)
	}
	if p.svrGrp.Num() <= 0 {
		err := fmt.Errorf("no address to listen")
		log.Errorf("[proxy] %v", err)
		return err
	}
	return nil
}

func (p *HttpProxy) initFilters(cfg map[string]interface{}) error {
	p.filters = filters.NewFilterManager()
	p.filters.PushStepPair(filters.BeforeAll, doRoute, filters.AfterAll, nil)
	p.filters.PushStepPair(filters.BeforeForward, doForward, filters.AfterForward, finishForward)
	p.filters.PushStepPair(filters.BeforeDispatch, doDispatch, filters.AfterDispatch, finishDispatch)
	return p.filters.Init(cfg)
}

func (p *HttpProxy) Start(closeCh <-chan os.Signal) error {
	log.Infof("[proxy] start servers number : %d", p.svrGrp.Num())
	defer p.Close()
	return p.svrGrp.Serve(closeCh, func(err error, addr string, svr servegrp.ServeItem) {
		if err != nil {
			log.Errorf("[proxy] server [ %s ] exit error: %v", addr, err)
		} else {
			log.Infof("[proxy] server [ %s ] exit ok", addr)
		}
	})
}

func (p *HttpProxy) Close() error {
	log.Info("[proxy] stop servers")
	return ioutil.CloseMulti(p.svrGrp, p.jobs, p.waitClose)
}
