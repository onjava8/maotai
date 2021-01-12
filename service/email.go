package service

import (
	"github.com/unknwon/goconfig"
	"github.com/ztino/jd_seckill/log"
	"gopkg.in/gomail.v2"
	"strconv"
)

type Email struct {
	host string
	port string
	user string
	pass string
}

func NewEmail(conf *goconfig.ConfigFile) *Email {
	host:=conf.MustValue("smtp","email_host","")
	port:=conf.MustValue("smtp","port","")
	user:=conf.MustValue("smtp","email_user","")
	pass:=conf.MustValue("smtp","email_pwd","")
	return &Email{host: host,port: port,user: user,pass: pass}
}

func (this *Email) Send(mailTo []string,subject,body string) error {
	port, _ := strconv.Atoi(this.port)
	m:=gomail.NewMessage()
	m.SetHeader("From", "<" + this.user + ">")
	m.SetHeader("To", mailTo...)
	m.SetHeader("Subject",subject)
	m.SetBody("text/html",body)
	d := gomail.NewDialer(this.host,port,this.user,this.pass)
	log.Warn("正在发送通知...")
	err:=d.DialAndSend(m)
	if err!=nil {
		log.Error("邮件发送失败，返回错误:"+err.Error())
	}else{
		log.Println("邮件发送成功")
	}
	return err
}