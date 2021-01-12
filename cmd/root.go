package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ztino/jd_seckill/common"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   common.SoftName,
	Short: "jd_seckill is a Jingdong Moutai seckill script",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_=cmd.Help()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
