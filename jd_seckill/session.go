package jd_seckill

import (
	"encoding/json"
	"errors"
	"github.com/Albert-Zhan/httpc"
	"github.com/ztino/jd_seckill/common"
	"github.com/ztino/jd_seckill/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Session struct {
	cookieJar *httpc.CookieJar
}

func NewSession(cookieJar *httpc.CookieJar) *Session {
	return &Session{cookieJar: cookieJar}
}

func (this *Session) SaveCookieToFile(path string) error {
	u,_:=url.Parse("https://jd.com")
	cookies:=this.cookieJar.Cookies(u)
	if len(cookies)==0 {
		log.Error("保存cookie失败，未找到相关cookie")
		return errors.New("保存cookie失败，未找到相关cookie")
	}
	data,_:=json.Marshal(cookies)

	err:=ioutil.WriteFile(path,data,0777)
	if err!=nil{
		log.Error("保存cookie失败，错误信息:"+err.Error())
		return errors.New("保存cookie失败，错误信息:"+err.Error())
	}
	log.Info("保存cookie成功")
	return nil
}

func (this *Session) LoadCookieToJar(path string) error {
	if !common.Exists(path) {
		log.Error("cookie文件不存在")
		return errors.New("cookie文件不存在")
	}
	data,err:=ioutil.ReadFile(path)
	if err!=nil {
		log.Error("读取cookie失败，错误信息:"+err.Error())
		return errors.New("读取cookie失败，错误信息:"+err.Error())
	}

	var cookies []*http.Cookie
	err=json.Unmarshal(data,&cookies)
	if err!=nil {
		log.Error("解析cookie失败，错误信息:"+err.Error())
		return errors.New("解析cookie失败，错误信息:"+err.Error())
	}

	u,_:=url.Parse("https://jd.com")
	this.cookieJar.SetCookies(u,cookies)
	log.Println("加载cookie成功")
	return nil
}

func (this *Session) CheckLoginStatus() error {
	err:=this.LoadCookieToJar(common.SoftDir+"/cookie.txt")
	if err!=nil {
		return err
	}
	//检测cookie会话状态
	user:=NewUser(common.Client,common.Config)
	if status:=user.RefreshStatus();status!=nil {
		_=os.Remove(common.SoftDir+"/cookie.txt")
		return errors.New("登录失效，请重新登录")
	}
	return nil
}