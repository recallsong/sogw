package myapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/recallsong/go-utils/container/dic"
	"github.com/recallsong/go-utils/net/echox"
	"github.com/recallsong/go-utils/net/servegrp"
	"github.com/recallsong/sogw/store"
	log "github.com/sirupsen/logrus"
)

var Debug bool

type Config struct {
	HttpAddr string `mapstructure:"api_addr"`
	GrpcAddr string `mapstructure:"grpc_store"`
	// k/v store
	Store StoreConfig `mapstructure:"store"`
}

type StoreConfig struct {
	Url     string                 `mapstructure:"url"`
	Options map[string]interface{} `mapstructure:"options"`
}

type ApiServer struct {
	cfg     *Config
	store   store.Store
	serGrp  *servegrp.ServeGroup
	httpSvr *echox.EchoServer
}

func NewApiServer() *ApiServer {
	s := &ApiServer{
		serGrp:  servegrp.NewServeGroup(),
		httpSvr: echox.New(),
	}
	s.httpSvr.Echo.Logger = LogrusToEchoLogger{}
	return s
}

func (s *ApiServer) Init(cfg *Config) (err error) {
	s.cfg = cfg
	store, err := store.New(cfg.Store.Url, cfg.Store.Options)
	if err != nil {
		log.Error("[apisvr] store.New failed : ", err)
		return err
	}
	s.store = store
	return s.initHttpApis()
}

func (s *ApiServer) Start(closeCh <-chan os.Signal) error {
	log.Infof("[apisvr] start servers number : %d", s.serGrp.Num())
	defer s.Close()
	return s.serGrp.Serve(closeCh, func(err error, addr string, svr servegrp.ServeItem) {
		if err != nil {
			log.Errorf("[apisvr] server [ %s ] exit error: %v", addr, err)
		} else {
			log.Infof("[apisvr] server [ %s ] exit ok", addr)
		}
	})
}

func (s *ApiServer) Close() error {
	log.Info("[apisvr] stop api server")
	return s.serGrp.Close()
}

func (s *ApiServer) ReadJSON(ctx echo.Context, out interface{}) error {
	body, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &out)
	if err != nil {
		s.WriteError(ctx, http.StatusBadRequest, "body is not json format")
		return err
	}
	return nil
}

func (s *ApiServer) WriteError(ctx echo.Context, code int, msg string) {
	ctx.JSON(code, dic.Dic{
		"code": code,
		"msg":  msg,
	})
}

func (s *ApiServer) WriteData(ctx echo.Context, data interface{}) {
	ctx.JSON(http.StatusOK, dic.Dic{
		"code": http.StatusOK,
		"data": data,
	})
}
