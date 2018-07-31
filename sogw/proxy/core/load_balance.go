package core

import (
	"net"
	"sync/atomic"

	"github.com/recallsong/go-utils/lang"
	"github.com/recallsong/sogw/store/meta"
)

type LoadBalance interface {
	Select(ctx *RequestContext, servers []*Server) *Server
}

var loadBalanceGetter = map[meta.LoadBalance]func() LoadBalance{
	meta.LoadBalance_RoundRobin: NewRoundRobinLB,
	meta.LoadBalance_IPHash:     NewIPHashLB,
}

type RoundRobinLB struct {
	_     lang.NoCopy
	index uint64
}

func NewRoundRobinLB() LoadBalance {
	return &RoundRobinLB{}
}

func (lb *RoundRobinLB) Select(ctx *RequestContext, servers []*Server) *Server {
	num := uint64(len(servers))
	if 0 >= num {
		return nil
	}
	return servers[atomic.AddUint64(&lb.index, 1)%num]
}

type IPHashLB struct{}

func NewIPHashLB() LoadBalance {
	return &IPHashLB{}
}

func (lb *IPHashLB) Select(ctx *RequestContext, servers []*Server) *Server {
	num := uint64(len(servers))
	if 0 >= num {
		return nil
	}
	addr, err := net.ResolveIPAddr("ip", ctx.GetRealClientAddr())
	if err != nil {
		return servers[uint64(len(err.Error()))%num]
	}
	var sum uint64
	for i, c := len(addr.IP)-1, 0; i >= 0 && c < 4; c, i = c+1, i-1 {
		sum += uint64(addr.IP[i])
	}
	if sum <= 0 {
		return servers[0]
	}
	return servers[sum%num]
}
