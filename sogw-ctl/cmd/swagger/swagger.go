package swagger

import (
	"bytes"
	"fmt"
	gpath "path"
	"regexp"
	"strings"

	"github.com/recallsong/go-utils/encoding/jsonx"
	"github.com/recallsong/sogw/sogw-ctl/cmd/common"
	"github.com/recallsong/sogw/store/meta"
	"github.com/spf13/viper"
)

type SwaggerConfig struct {
	Swagger string `mapstructure:"swagger"`
	Info    struct {
		Description string `mapstructure:"description"`
		Version     string `mapstructure:"version"`
		// ...
	} `mapstructure:"info"`
	Host     string `mapstructure:"host"`
	BasePath string `mapstructure:"basePath"`
	Tags     []*struct {
		Name        string `mapstructure:"name"`
		Description string `mapstructure:"description"`
	} `mapstructure:"tags"`
	Schemes []string `mapstructure:"schemes"`
	Paths   map[string]map[string]*struct {
		Tags        []string `mapstructure:"tags"`
		Description string   `mapstructure:"description"`
		// ... Others
	} `mapstructure:"paths"`
}

func ReadAndParseConfig() (pubs *PublishInfo, err error) {
	sc, err := readSwaggerConfig()
	if err != nil {
		return nil, err
	}
	pubs, err = parseSwaggerConfig(sc)
	if err != nil {
		return nil, err
	}
	return pubs, nil
}

func readSwaggerConfig() (*SwaggerConfig, error) {
	viper := viper.New()
	viper.SetConfigFile(common.Config.Publish.Swagger.File)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	sc := &SwaggerConfig{}
	err = viper.Unmarshal(sc)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

type PublishInfo struct {
	Host     meta.Host
	Routes   []*meta.Route
	Services map[string]*Service
}

type Service struct {
	Meta meta.Service
	Apis []*meta.Api
	Svrs []*meta.Server
	Cfg  *meta.ServiceConfig
}

func NewPublishInfo() *PublishInfo {
	return &PublishInfo{
		Services: make(map[string]*Service),
	}
}

func parseSwaggerConfig(sc *SwaggerConfig) (pubs *PublishInfo, err error) {
	var regx *regexp.Regexp
	if common.Config.Publish.Mapper.PathPattern != "" {
		regx, err = regexp.Compile(common.Config.Publish.Mapper.PathPattern)
		if err != nil {
			return pubs, err
		}
	}
	pubs = NewPublishInfo()
	pubs.Host.Value = sc.Host
	pubs.Host.Kind = meta.HostKind_Allow
	for path, info := range sc.Paths {
		path = gpath.Join(sc.BasePath, path)
		rewritePath := path
		if regx != nil {
			subs := regx.FindAllStringSubmatch(path, -1)
			if len(subs) <= 0 {
				return nil, fmt.Errorf("path %s not match pattern %s", path, common.Config.Publish.Mapper.PathPattern)
			}
			strs := subs[len(subs)-1]
			if len(strs) <= 0 {
				return nil, fmt.Errorf("path %s not match pattern %s", path, common.Config.Publish.Mapper.PathPattern)
			}
			rewritePath = strs[len(strs)-1]
			rewritePath, err = pathToRoute(rewritePath)
			if err != nil {
				return nil, err
			}
		}
		for method, info := range info {
			method = strings.ToUpper(method)
			if len(info.Tags) != 1 {
				return nil, fmt.Errorf("muti tags %v in path %s", info.Tags, path)
			}
			tag := info.Tags[0]
			ser, ok := pubs.Services[tag]
			if !ok {
				ser = &Service{}
				ser.Meta.Name = tag
				pubs.Services[tag] = ser
			}
			// api
			api := &meta.Api{
				Status:  meta.Status_Open,
				Method:  method,
				Version: sc.Info.Version,
				Path:    rewritePath,
			}
			api.InitId()
			ser.Apis = append(ser.Apis, api)
			// ruote
			path, err = pathToRoute(path)
			if err != nil {
				return nil, err
			}
			ser.Meta.InitId()
			rt := &meta.Route{
				Path:    path,
				Method:  method,
				Status:  meta.Status_Open,
				Service: ser.Meta.Id,
				ApiId:   api.Id,
			}
			pubs.Routes = append(pubs.Routes, rt)
		}
	}
	// server
	for name, s := range common.Config.Publish.Mapper.Services {
		if ser, ok := pubs.Services[name]; ok {
			for _, svr := range s.Servers {
				ser.Svrs = append(ser.Svrs, &meta.Server{
					Status: meta.Status_Open,
					Name:   svr.Addr,
					Addr:   svr.Addr,
				})
			}
			ser.Cfg = &meta.ServiceConfig{
				Status:     meta.Status_Open,
				LoadBlance: meta.LoadBalance_RoundRobin,
			}
			if s.Config != nil {
				if s.Config.LoadBlance != "" {
					if lb, ok := meta.LoadBalance_value[s.Config.LoadBlance]; ok {
						ser.Cfg.LoadBlance = meta.LoadBalance(lb)
					} else {
						return nil, fmt.Errorf("invalid LoadBlance %s", s.Config.LoadBlance)
					}
				}
				if s.Config.Status != "" {
					if st, ok := meta.Status_value[s.Config.Status]; ok {
						ser.Cfg.Status = meta.Status(st)
					} else {
						return nil, fmt.Errorf("invalid Status %s", s.Config.Status)
					}
				}
			}
		}
	}
	return pubs, err
}

func pathToRoute(path string) (string, error) {
	i, j, l := 0, 0, len(path)
	if l == 0 {
		return "", fmt.Errorf("path should not be empty")
	}
	buf := bytes.Buffer{}
	for i < l {
		if path[i] == '{' {
			if path[:i] != "" {
				buf.WriteString(path[:i])
			}
			j = i
			i++
			if i >= l {
				buf.WriteString("{")
				return buf.String(), nil
			}
			for ; i < l && path[i] != '}'; i++ {
			}
			if i >= l {
				buf.WriteString(path[j:])
				return buf.String(), nil
			}
			buf.WriteString(":" + path[j+1:i])
			i++
			if i >= l {
				return buf.String(), nil
			}
			path = path[i:]
			i, l = 0, len(path)
			continue
		}
		i++
	}
	if path != "" {
		buf.WriteString(path)
	}
	return buf.String(), nil
}

func (p *PublishInfo) String() string {
	return jsonx.MarshalAndIntend(p)
}
