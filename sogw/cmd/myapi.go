package cmd

import (
	"syscall"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/os/signalx"
	"github.com/recallsong/sogw/sogw/myapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var apiCfg myapi.Config

var myApiCmd = &cobra.Command{
	Use:   "myapi",
	Short: "start gateway api server",
	Long:  `start gateway api server`,
	Run: func(cmd *cobra.Command, args []string) {
		cobrax.InitCommand(&apiCfg)
		// init server
		svr := myapi.NewApiServer()
		err := svr.Init(&apiCfg)
		if err != nil {
			return
		}
		svr.Start(signalx.Notify(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT))
	},
}

func initMyApiCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(myApiCmd)
	fs := myApiCmd.Flags()
	myApiCmd.Flags().StringVar(&apiCfg.HttpAddr, "api_addr", "", "address for http server")
	viper.BindPFlags(fs)
}
