package meta

import (
	"bytes"

	"github.com/pborman/uuid"
	"github.com/recallsong/go-utils/container/slice/strings"
	"github.com/recallsong/go-utils/encoding/md5x"
)

func (h *Host) InitId() {
	h.Id = md5x.SumString(h.Value).String16()
}

func (a *Auth) InitId() {
	a.Id = md5x.Sum([]byte(uuid.NewRandom())).String16()
}

func (r *Route) InitId() {
	buf := &bytes.Buffer{}
	if len(r.Method) <= 0 {
		buf.WriteString("*")
	} else {
		buf.WriteString(r.Method)
	}
	buf.WriteString(":")
	if len(r.Path) <= 0 {
		buf.WriteString("/")
	} else if r.Path[0] != '/' {
		buf.WriteString("/")
	}
	buf.WriteString(r.Path)
	r.Id = md5x.Sum(buf.Bytes()).String16()
}

func (s *Service) InitId() {
	s.Id = md5x.SumString(s.Name).String16()
}

func (a *Api) InitId() {
	buf := &bytes.Buffer{}
	if a.Method == "" {
		buf.WriteString(a.Version)
		buf.WriteString(":*:")
	} else {
		buf.WriteString(a.Version)
		buf.WriteString(":")
		buf.WriteString(a.Method)
		buf.WriteString(":")
	}
	if a.Path == "" {
		buf.WriteString("$*")
	} else {
		buf.WriteString(a.Path)
	}
	a.Id = md5x.Sum(buf.Bytes()).String16()
}

func (s *Server) InitId() {
	s.Id = md5x.SumString(s.Addr).String16()
}

func (c *ServiceConfig) InitId() {
	c.Id = md5x.SumString("/cfg").String16()
}

func (g *Gateway) InitId() {
	g.Id = md5x.SumString(strings.Strings(g.Addrs).Sort().Join(",")).String16()
}
