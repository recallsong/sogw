package cmd

import (
	"fmt"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/encoding/jsonx"
	"github.com/recallsong/sogw/sogw-ctl/cmd/common"
	"github.com/recallsong/sogw/store"
	"github.com/recallsong/sogw/store/meta"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func InitShowCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all information about sogw in store",
	Run: func(cmd *cobra.Command, args []string) {
		cobrax.InitCommand(&common.Config)
		s := common.InitStore()
		err := showStoreInfo(s)
		if err != nil {
			s.Close()
			log.Fatal("[show] ", err)
			return
		}
		s.Close()
		log.Info("[show] ok")
	},
}

func showStoreInfo(s store.Store) error {
	type Service struct {
		*meta.Service
		Cfg     *meta.ServiceConfig `json:"cfg"`
		Apis    []*meta.Api         `json:"apis"`
		Servers []*meta.Server      `json:"svrs"`
	}
	type StoreData struct {
		Hosts    []*meta.Host  `json:"hosts"`
		Auths    []*meta.Auth  `json:"auths"`
		Routes   []*meta.Route `json:"routes"`
		Services []*Service    `json:"services"`
	}
	sd := &StoreData{}
	err := s.GetHosts(func(item *meta.Host) {
		if item != nil {
			if err := item.Valid(); err != nil {
				log.Warn("[show] [hosts] : ", err.Error())
				return
			}
			sd.Hosts = append(sd.Hosts, item)
		}
	})
	if err != nil {
		return err
	}
	err = s.GetAuths(func(item *meta.Auth) {
		if item != nil {
			if err := item.Valid(); err != nil {
				log.Warn("[show] [auths] : ", err.Error())
				return
			}
			sd.Auths = append(sd.Auths, item)
		}
	})
	if err != nil {
		return err
	}
	err = s.GetRoutes(func(item *meta.Route) {
		if item != nil {
			if err := item.Valid(); err != nil {
				log.Warn("[show] [routes] : ", err.Error())
				return
			}
			sd.Routes = append(sd.Routes, item)
		}
	})
	if err != nil {
		return err
	}
	err = s.GetServices(func(item *meta.Service) {
		if item != nil {
			if err := item.Valid(); err != nil {
				log.Warn("[show] [services] : ", err.Error())
				return
			}
			sd.Services = append(sd.Services, &Service{
				Service: item,
			})
		}
	})
	if err != nil {
		return err
	}
	for _, ser := range sd.Services {
		cfg, err := s.GetServiceCfg(ser.Id)
		if err != nil {
			return err
		}
		if cfg != nil {
			if err := cfg.Valid(); err != nil {
				log.Warn("[show] [service.cfg] : ", err.Error())
			} else {
				ser.Cfg = cfg
			}
		}
		err = s.GetApis(ser.Id, func(item *meta.Api) {
			if item != nil {
				if err := item.Valid(); err != nil {
					log.Warn("[show] [service.apis] : ", err.Error())
					return
				}
				ser.Apis = append(ser.Apis, item)
			}
		})
		if err != nil {
			return err
		}
		err = s.GetServers(ser.Id, func(item *meta.Server) {
			if item != nil {
				if err := item.Valid(); err != nil {
					log.Warn("[show] [service.svrs] : ", err.Error())
					return
				}
				ser.Servers = append(ser.Servers, item)
			}
		})
		if err != nil {
			return err
		}
	}
	fmt.Println(jsonx.MarshalAndIntend(sd))
	return nil
}
