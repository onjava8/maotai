package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ztino/jd_seckill/common"
	"runtime"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of jd_seckill",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("%s version %s %s %s/%s", common.SoftName, common.SoftName, common.Version, runtime.GOOS, runtime.GOARCH))
	},
}