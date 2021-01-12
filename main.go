package main

import (
	"github.com/Albert-Zhan/httpc"
	"github.com/unknwon/goconfig"
	"github.com/ztino/jd_seckill/cmd"
	"github.com/ztino/jd_seckill/common"
	"github.com/ztino/jd_seckill/log"
	"os"
	"runtime"
)

func init() {
	//软件目录获取
	common.SoftDir = "."
	if dir, err := os.Getwd(); err == nil {
		common.SoftDir = dir
	}

	//日志初始化
	if !common.IsDir(common.SoftDir + "/logs/") {
		_ = os.Mkdir(common.SoftDir+"/logs/", 0777)
	}

	//客户端设置初始化
	common.Client = httpc.NewHttpClient()
	common.CookieJar = httpc.NewCookieJar()
	common.Client.SetCookieJar(common.CookieJar)

	//配置文件初始化
	confFile := common.SoftDir + "/conf.ini"
	if !common.Exists(confFile) {
		file, err := os.OpenFile(confFile, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("写入配置文件失败，错误：%v", err)
			os.Exit(0)
		}
		defer file.Close()
		if _, err := file.WriteString(common.IniFileContent); err == nil {
			log.Println("初始化配置文件成功")
		}
	}
	var err error
	if common.Config, err = goconfig.LoadConfigFile(confFile); err != nil {
		log.Error("配置文件加载失败，程序退出")
		os.Exit(0)
	}

	//抢购状态管道
	common.SeckillStatus = make(chan bool)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
