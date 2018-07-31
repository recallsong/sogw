package core

import (
	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/go-utils/reflectx"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
)

type Hosts map[string]*Host

func (hs Hosts) ValidateHost(ctx *RequestContext) bool {
	host := reflectx.BytesToString(ctx.ReqCtx.Host())
	if h, ok := hs[host]; ok {
		if h.Meta.Kind == meta.HostKind_Deny {
			if cobrax.Flags.Debug {
				log.Debug("[Hosts] host %s denied", host)
			}
			return false
		}
		ctx.Host = h
		return true
	} else {
		if h, ok := hs["*"]; ok {
			if h.Meta.Kind == meta.HostKind_Deny {
				if cobrax.Flags.Debug {
					log.Debugf("[Hosts] host %s denied (by *)", host)
				}
				return false
			}
			ctx.Host = h
			return true
		}
	}
	if cobrax.Flags.Debug {
		log.Debugf("[Hosts] host %s not defined", host)
	}
	return true
}

type Host struct {
	_    lang.NoCopy
	Meta *meta.Host
}

func NewHost(m *meta.Host) *Host {
	return &Host{Meta: m}
}
