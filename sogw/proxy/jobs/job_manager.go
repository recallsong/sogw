package jobs

import (
	"fmt"
	"sync"

	"github.com/recallsong/go-utils/errorx"
	"github.com/recallsong/sogw/sogw/proxy/core"
	log "github.com/sirupsen/logrus"
)

type JobFactory func() Job

var supportJobs map[string]JobFactory

type JobManager struct {
	jobs map[string]Job
	lock sync.RWMutex
}

func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]Job),
	}
}

func (jm *JobManager) Add(name string, j Job) error {
	if j == nil || len(name) <= 0 {
		return nil
	}
	jm.lock.Lock()
	defer jm.lock.Unlock()
	if _, ok := jm.jobs[name]; ok {
		err := fmt.Errorf("%s job already exist", name)
		log.Error(err)
		return err
	}
	jm.jobs[name] = j
	return nil
}

func (jm *JobManager) SetupAll(cfg map[string]interface{}) error {
	if cfg == nil {
		return nil
	}
	jm.lock.Lock()
	defer jm.lock.Unlock()
	var j Job
	for name, cfg := range cfg {
		if v, ok := jm.jobs[name]; ok {
			j = v
		} else {
			if jf, ok := supportJobs[name]; !ok {
				j = jf()
			} else {
				err := fmt.Errorf("%s job not support", name)
				log.Error(err)
				return err
			}
		}
		if cfg != nil {
			if cfg, ok := cfg.(map[string]interface{}); ok {
				j.Setup(cfg)
			} else {
				err := fmt.Errorf("invalid config of %s job", name)
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func (jm *JobManager) Run(ctx *core.RuntimeContext, stopCh <-chan struct{}, wg *sync.WaitGroup) {
	jm.lock.RLock()
	defer jm.lock.RUnlock()
	wg.Add(len(jm.jobs))
	for n, j := range jm.jobs {
		go func(n string, j Job) {
			defer wg.Done()
			log.Info("%s job start.")
			err := j.Run(ctx, stopCh)
			if err != nil {
				log.Errorf("%s job run error : %s", n, err.Error())
			}
		}(n, j)
	}
}

func (jm *JobManager) Close() error {
	jm.lock.RLock()
	defer jm.lock.RUnlock()
	errs := errorx.Errors{}
	for n, j := range jm.jobs {
		err := j.Close()
		if err != nil {
			log.Errorf("%s job close error : %s", n, err.Error())
		}
		errs = append(errs, err)
	}
	return errs.MaybeUnwrap()
}

func (jm *JobManager) Update(ctx *core.RuntimeContext) error {
	jm.lock.RLock()
	defer jm.lock.RUnlock()
	errs := errorx.Errors{}
	for n, j := range jm.jobs {
		err := j.Update(ctx)
		if err != nil {
			log.Errorf("%s job update error : %s", n, err.Error())
		}
		errs = append(errs, err)
	}
	return errs.MaybeUnwrap()
}

func RegisterJobFactory(name string, jf JobFactory) {
	supportJobs[name] = jf
}
