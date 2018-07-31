package cmd

import (
	"os"

	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/sogw/sogw-ctl/cmd/common"
	"github.com/recallsong/sogw/sogw-ctl/cmd/swagger"
	"github.com/recallsong/sogw/store"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func InitPubCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(pubCmd)
}

var pubCmd = &cobra.Command{
	Use:   "pub",
	Short: "Publish something to store",
	Run: func(cmd *cobra.Command, args []string) {
		cobrax.InitCommand(&common.Config)
		pubs, err := swagger.ReadAndParseConfig()
		if err != nil {
			log.Fatal("[pub] ", err)
			return
		}
		if cobrax.Flags.Debug {
			log.Debug("\n", pubs, " \nContinue [y/n]: ")
			c := []byte{0}
			if _, err := os.Stdin.Read(c[:]); err != nil {
				log.Fatal("[pub] ", err)
				return
			}
			if c[0] == 'n' || c[0] == 'N' {
				return
			}
		}
		s := common.InitStore()
		err = doPublish(s, pubs)
		if err != nil {
			s.Close()
			log.Fatal("[pub] ", err)
			return
		}
		s.Close()
		log.Info("[pub] ok")
	},
}

func doPublish(s store.Store, pubs *swagger.PublishInfo) error {
	insert := 0
	err := s.PutHost(&pubs.Host)
	if err != nil {
		return err
	}
	log.Info("[pub] host: ", jsonx.Marshal(pubs.Host))
	insert++
	for _, ser := range pubs.Services {
		err = s.PutService(&ser.Meta)
		if err != nil {
			return err
		}
		log.Info("[pub] service: ", jsonx.Marshal(ser.Meta))
		insert++
		if ser.Cfg != nil {
			err = s.PutServiceCfg(ser.Meta.Id, ser.Cfg)
			if err != nil {
				return err
			}
			log.Infof("[pub] %s cfg: %s", ser.Meta.Name, jsonx.Marshal(ser.Cfg))
		}
		for _, api := range ser.Apis {
			err = s.PutApi(ser.Meta.Id, api)
			if err != nil {
				return err
			}
			log.Infof("[pub] %s api: %s", ser.Meta.Name, jsonx.Marshal(api))
			insert++
		}
		for _, svr := range ser.Svrs {
			err = s.PutServer(ser.Meta.Id, svr)
			if err != nil {
				return err
			}
			log.Infof("[pub] %s server: %s", ser.Meta.Name, jsonx.Marshal(svr))
			insert++
		}
	}
	for _, r := range pubs.Routes {
		err = s.PutRoute(r)
		if err != nil {
			return err
		}
		log.Info("[pub] route: ", jsonx.Marshal(r))
		insert++
	}
	log.Info("[pub] inserted: ", insert)
	return nil
}
