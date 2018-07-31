package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/recallsong/go-utils/encoding/jsonx"
	"github.com/recallsong/sogw/store/meta"
	file "github.com/recallsong/sogw/store/store-file"
)

func exit(err error) {
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(1)
}

func main() {
	s, err := file.NewStore(".", "/meta.json")
	if err != nil {
		exit(err)
	}
	err = s.GetHosts(func(item *meta.Host) {
		fmt.Println("hosts: ", jsonx.Marshal(item))
	})
	if err != nil {
		exit(err)
	}
	err = s.GetAuths(func(item *meta.Auth) {
		fmt.Println("auths: ", jsonx.Marshal(item))
	})
	if err != nil {
		exit(err)
	}
	err = s.GetRoutes(func(item *meta.Route) {
		fmt.Println("routes: ", jsonx.Marshal(item))
	})
	if err != nil {
		exit(err)
	}
	err = s.GetServices(func(item *meta.Service) {
		fmt.Println("services: ", jsonx.Marshal(item))
	})
	if err != nil {
		exit(err)
	}
	wg := sync.WaitGroup{}
	stopCh := make(chan struct{})
	ln := Ln{}
	s.Watch(ln, stopCh, &wg)
	<-stopCh
	fmt.Println("ok")
}

type Ln struct{}

func (ln Ln) RecvHost(op meta.Operation, data *meta.Host) {
	fmt.Println(op, "host: ", jsonx.Marshal(data))
}
func (ln Ln) RecvAuth(op meta.Operation, data *meta.Auth) {
	fmt.Println(op, "auth: ", jsonx.Marshal(data))
}
func (ln Ln) RecvRoute(op meta.Operation, data *meta.Route) {
	fmt.Println(op, "routes: ", jsonx.Marshal(data))
}
func (ln Ln) RecvService(op meta.Operation, data *meta.Service) {
	fmt.Println(op, "services: ", jsonx.Marshal(data))
}
func (ln Ln) RecvServiceConfig(op meta.Operation, service string, data *meta.ServiceConfig) {
	fmt.Println(op, "service config: ", service, jsonx.Marshal(data))
}
func (ln Ln) RecvApi(op meta.Operation, service string, data *meta.Api) {
	fmt.Println(op, "service api: ", service, jsonx.Marshal(data))
}
func (ln Ln) RecvServer(op meta.Operation, service string, data *meta.Server) {
	fmt.Println(op, "service server: ", service, jsonx.Marshal(data))
}
