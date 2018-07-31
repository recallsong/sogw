package meta

func (h *Host) Copy() *Host {
	val := *h
	return &val
}

func (a *Auth) Copy() *Auth {
	val := *a
	cfg := make(map[string]string)
	for k, v := range a.Config {
		cfg[k] = v
	}
	val.Config = cfg
	return &val
}

func (r *Route) Copy() *Route {
	val := *r
	if r.Context != nil {
		ctx := make(map[string]*ValueItem)
		for k, v := range r.Context {
			val := *v
			ctx[k] = &val
		}
		val.Context = ctx
	}
	if r.ApiConds != nil {
		conds := make([]*ApiCondition, len(r.ApiConds))
		for i, v := range r.ApiConds {
			val := *v
			m := *val.Matcher
			val.Matcher = &m
			conds[i] = &val
		}
		val.ApiConds = conds
	}
	return &val
}

func (s *Service) Copy() *Service {
	val := *s
	return &val
}

func (a *Api) Copy() *Api {
	val := *a
	if a.Context != nil {
		ctx := make(map[string]*ValueItem)
		for k, v := range a.Context {
			val := *v
			ctx[k] = &val
		}
		val.Context = ctx
	}
	if a.Headers != nil {
		headers := *a.Headers
		if a.Headers.ToClient != nil {
			toClient := make([]*HeaderItem, len(a.Headers.ToClient))
			for i, v := range a.Headers.ToClient {
				val := *v
				toClient[i] = &val
			}
			headers.ToClient = toClient
		}
		if a.Headers.ToBackend != nil {
			toBackend := make([]*HeaderItem, len(a.Headers.ToBackend))
			for i, v := range a.Headers.ToBackend {
				val := *v
				toBackend[i] = &val
			}
			headers.ToBackend = toBackend
		}
		val.Headers = &headers
	}
	if a.Cookies != nil {
		cookies := *a.Cookies
		if a.Cookies.ToClient != nil {
			toClient := make([]*CookieItem, len(a.Cookies.ToClient))
			for i, v := range a.Cookies.ToClient {
				val := *v
				toClient[i] = &val
			}
			cookies.ToClient = toClient
		}
		if a.Cookies.ToBackend != nil {
			toBackend := make([]*CookieItem, len(a.Cookies.ToBackend))
			for i, v := range a.Cookies.ToBackend {
				val := *v
				toBackend[i] = &val
			}
			cookies.ToBackend = toBackend
		}
		val.Cookies = &cookies
	}
	if a.Validators != nil {
		valids := make([]*Validator, len(a.Validators))
		for i, v := range a.Validators {
			val := *v
			m := *val.Matcher
			val.Matcher = &m
			valids[i] = &val
		}
		val.Validators = valids
	}
	return &val
}

func (s *Server) Copy() *Server {
	val := *s
	if s.HealthCheck != nil {
		hc := *s.HealthCheck
		val.HealthCheck = &hc
	}
	return &val
}
func (c *ServiceConfig) Copy() *ServiceConfig {
	val := *c
	if c.Context != nil {
		ctx := make(map[string]*ValueItem)
		for k, v := range c.Context {
			val := *v
			ctx[k] = &val
		}
		val.Context = ctx
	}
	return &val
}

func (g *Gateway) Copy() *Gateway {
	val := *g
	if g.Addrs != nil {
		addrs := make([]string, len(g.Addrs))
		for i, v := range g.Addrs {
			addrs[i] = v
		}
		val.Addrs = addrs
	}
	return &val
}
