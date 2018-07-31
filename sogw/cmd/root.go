package cmd

import (
	_ "net/http/pprof"
	"os"
	"syscall"

	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/go-utils/os/signalx"
	"github.com/recallsong/sogw/sogw/proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func Execute() {
	cobrax.Execute("sogw", options)
}

var (
	proxyCfg proxy.Config
	options  = &cobrax.Options{
		CfgDir:      "conf",
		CfgFileName: "sogw",
		AppConfig:   &proxyCfg,
		Init: func(cmd *cobra.Command) {
			cmd.Short = "sogw is an api gateway"
			cmd.Long = `sogw is an api gateway.`
			// proxy configs
			fs := cmd.Flags()
			fs.StringVar(&proxyCfg.Addr, "addr", "", "addr for proxy listen")
			fs.StringVar(&proxyCfg.TLSAddr, "tls_addr", "", "tls addr for proxy listen (addr,certFile,keyFile)")
			fs.StringVar(&proxyCfg.UnixAddr, "unix_addr", "", "unix addr for proxy listen")
			viper.BindPFlags(fs)
			initMyApiCmd(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			// start http proxy
			pxy := proxy.New()
			err := pxy.Init(&proxyCfg)
			if err != nil {
				os.Exit(1)
			}
			pxy.Start(signalx.Notify(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT))
		},
	}
)
