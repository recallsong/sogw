package main

import (
	"github.com/recallsong/cliframe/cobrax"
	"github.com/recallsong/sogw/sogw-ctl/cmd"
	"github.com/spf13/cobra"
)

func main() {
	cobrax.Execute("sogw-ctl", &cobrax.Options{
		CfgDir:      ".",
		CfgFileName: "sogw-ctl",
		Init: func(rootCmd *cobra.Command) {
			cmd.InitPubCmd(rootCmd)
			cmd.InitShowCmd(rootCmd)
			cmd.InitCleanCmd(rootCmd)
		},
	})
}
