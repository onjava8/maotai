package service

import (
	"github.com/blinkbean/dingtalk"
	"github.com/unknwon/goconfig"
	"github.com/ztino/jd_seckill/log"
	"regexp"
)

type Dingtalk struct {
	conf *goconfig.ConfigFile
}

func NewDingtalk(conf *goconfig.ConfigFile) *Dingtalk {
	return &Dingtalk{conf: conf}
}

func (this *Dingtalk) Send(title, msg string) error {
	cli := dingtalk.InitDingTalkWithSecret(
		this.conf.MustValue("dingtalk", "access_token", ""),
		this.conf.MustValue("dingtalk", "secret", ""),
	)
	markdown := []string{
		"### " + title,
		"---------",
		msg,
	}

	var err error
	at := this.conf.MustValue("dingtalk", "at", "none")

	if at == "all" {
		err = cli.SendMarkDownMessageBySlice(title, markdown, dingtalk.WithAtAll())
	} else {
		reg := regexp.MustCompile(`(\d{11})`)
		mobiles := reg.FindAllString(at, 2)
		if len(mobiles) > 0 {
			err = cli.SendMarkDownMessageBySlice(title, markdown, dingtalk.WithAtMobiles(mobiles))
		} else {
			err = cli.SendMarkDownMessageBySlice(title, markdown)
		}
	}

	if err != nil {
		log.Error(err)
		return err
	}

	log.Println("钉钉机器人推送成功")
	return nil
}
