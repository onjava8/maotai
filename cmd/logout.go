package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ztino/jd_seckill/common"
	"github.com/ztino/jd_seckill/log"
	"os"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Open JD’s simulated logout",
	Run: startLogout,
}

func startLogout(cmd *cobra.Command, args []string)  {
	if common.Exists(common.SoftDir+"/cookie.txt") {
		_=os.Remove(common.SoftDir+"/cookie.txt")
		log.Println("退出成功")
	}else{
		log.Error("退出失败，未登录")
	}
}