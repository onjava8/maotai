package service

import (
	"errors"
	"fmt"
	"github.com/Albert-Zhan/httpc"
	"github.com/tidwall/gjson"
	"github.com/unknwon/goconfig"
	"github.com/ztino/jd_seckill/log"
)

type Wechat struct {
	conf *goconfig.ConfigFile
}

func NewWechat(conf *goconfig.ConfigFile) *Wechat {
	return &Wechat{conf: conf}
}

func (this *Wechat) Send(title,msg string) error {
	client:=httpc.NewHttpClient()
	req:=httpc.NewRequest(client)
	url:=fmt.Sprintf("http://sc.ftqq.com/%s.send",this.conf.MustValue("messenger","server_chan_sckey",""))
	req.SetHeader("User-Agent","Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")
	log.Warn("正在发送通知...")
	_,body,_:=req.SetUrl(url+"?text="+title+"&desp="+msg).SetMethod("get").Send().End()
	if gjson.Get(body,"errno").Int()!=0 {
		log.Error("微信推送失败，返回错误:"+gjson.Get(body,"errmsg").String())
		return errors.New("微信推送失败，返回错误:"+gjson.Get(body,"errmsg").String())
	}
	log.Println("微信推送成功")
	return nil
}