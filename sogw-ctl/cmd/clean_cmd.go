package cmd

import (
	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/sogw/sogw-ctl/cmd/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func InitCleanCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(cleanCmd)
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean store",
	Run: func(cmd *cobra.Command, args []string) {
		cobrax.InitCommand(&common.Config)
		s := common.InitStore()
		err := s.Clean()
		if err != nil {
			s.Close()
			log.Fatal("[clean] ", err)
			return
		}
		s.Close()
		log.Info("[clean] ok")
	},
}
