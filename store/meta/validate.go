package meta

import "errors"

func (h *Host) Valid() error {
	if h.Id == "" {
		return errors.New("host id should not be empty")
	}
	if _, ok := HostKind_name[int32(h.Kind)]; !ok {
		return errors.New("invalid host kind value")
	}
	if h.Value == "" {
		return errors.New("host value should not be empty")
	}
	return nil
}

func (a *Auth) Valid() error {
	if a.Id == "" {
		return errors.New("auth id should not be empty")
	}
	if _, ok := AuthKind_name[int32(a.Kind)]; !ok {
		return errors.New("invalid auth kind value")
	}
	return nil
}

func (r *Route) Valid() error {
	if r.Id == "" {
		return errors.New("route id should not be empty")
	}
	if _, ok := Status_name[int32(r.Status)]; !ok {
		return errors.New("invalid route status value")
	}
	return nil
}

func (s *Service) Valid() error {
	if s.Id == "" {
		return errors.New("service id should not be empty")
	}
	if s.Name == "" {
		return errors.New("service name should not be empty")
	}
	return nil
}

func (a *Api) Valid() error {
	if a.Id == "" {
		return errors.New("api id should not be empty")
	}
	if _, ok := Status_name[int32(a.Status)]; !ok {
		return errors.New("invalid api status value")
	}
	return nil
}

func (s *Server) Valid() error {
	if s.Id == "" {
		return errors.New("server id should not be empty")
	}
	if s.Addr == "" {
		return errors.New("server addr should not be empty")
	}
	if _, ok := Status_name[int32(s.Status)]; !ok {
		return errors.New("invalid server status value")
	}
	return nil
}

func (c *ServiceConfig) Valid() error {
	if c.Id == "" {
		return errors.New("service config id should not be empty")
	}
	if _, ok := LoadBalance_name[int32(c.LoadBlance)]; !ok {
		return errors.New("invalid service loadblance value")
	}
	if _, ok := Status_name[int32(c.Status)]; !ok {
		return errors.New("invalid service status value")
	}
	return nil
}

func (g *Gateway) Valid() error {
	if g.Id == "" {
		return errors.New("gateway id should not be empty")
	}
	if len(g.Addrs) <= 0 {
		return errors.New("gateway addrs should not be empty")
	}
	return nil
}
