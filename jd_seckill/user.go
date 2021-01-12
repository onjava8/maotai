package jd_seckill

import (
	"errors"
	"fmt"
	"github.com/Albert-Zhan/httpc"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/tidwall/gjson"
	"github.com/unknwon/goconfig"
	"github.com/ztino/jd_seckill/common"
	"github.com/ztino/jd_seckill/log"
	"github.com/ztino/jd_seckill/service"
	"image"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type User struct {
	client *httpc.HttpClient
	conf   *goconfig.ConfigFile
}

func NewUser(client *httpc.HttpClient, conf *goconfig.ConfigFile) *User {
	return &User{client: client, conf: conf}
}

func (this *User) getUserAgent() string {
	return this.conf.MustValue("config", "default_user_agent", "")
}

func (this *User) loginPage() {
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Connection", "keep-alive")
	req.SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	_, _, _ = req.SetUrl("https://passport.jd.com/new/login.aspx").SetMethod("get").Send().End()
}

func (this *User) QrLogin() (string, error) {
	//登录页面
	this.loginPage()
	//二维码登录
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Referer", "https://passport.jd.com/new/login.aspx")
	resp, err := req.SetUrl("https://qr.m.jd.com/show?appid=133&size=300&t="+strconv.Itoa(int(time.Now().Unix()*1000))).SetMethod("get").Send().EndFile(common.SoftDir+"/", "qr_code.png")
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Error("获取二维码失败")
		return "", errors.New("获取二维码失败")
	}
	cookies := resp.Cookies()
	wlfstkSmdl := ""
	for _, cookie := range cookies {
		if cookie.Name == "wlfstk_smdl" {
			wlfstkSmdl = cookie.Value
			break
		}
	}
	log.Println("二维码获取成功，请打开京东APP扫描")
	qrPath := filepath.Join(common.SoftDir, `./qr_code.png`)

	qrcodeShowType := this.conf.MustValue("config", "qrcode_show_type", "open")
	log.Debug("登录二维码展示方式：", qrcodeShowType)
	if qrcodeShowType == "dingtalk" {
		go func() {
			file, _ := os.Open(qrPath)
			defer file.Close()
			img, _, _ := image.Decode(file)
			bmp, _ := gozxing.NewBinaryBitmapFromImage(img)
			qrReader := qrcode.NewQRCodeReader()
			res, _ := qrReader.Decode(bmp, nil)
			qrcodeCreateApi := this.conf.MustValue("config", "qrcode_create_api", "https://api.pwmqr.com/qrcode/create/?url=")
			msg := fmt.Sprintf("二维码获取成功，请打开京东APP扫描\n\n![](%s%s)\n\n", qrcodeCreateApi, url.QueryEscape(res.String()))
			_ = service.SendMessage(this.conf, "京东秒杀通知", msg)
		}()
	} else {
		common.OpenImage(qrPath, qrcodeShowType)
	}

	return wlfstkSmdl, nil
}

func (this *User) QrcodeTicket(wlfstkSmdl string) (string, error) {
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Referer", "https://passport.jd.com/new/login.aspx")
	resp, body, err := req.SetUrl("https://qr.m.jd.com/check?appid=133&callback=jQuery" + strconv.Itoa(common.Rand(1000000, 9999999)) + "&token=" + wlfstkSmdl + "&_=" + strconv.Itoa(int(time.Now().Unix()*1000))).SetMethod("get").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("获取二维码扫描结果异常")
		return "", errors.New("获取二维码扫描结果异常")
	}
	code := gjson.Get(body, "code").String()
	msg := gjson.Get(body, "msg").String()
	//如果二维码过期，直接结束
	if code == "203" {
		log.Errorf("Code: %s, Message: %s", code, msg)
		os.Exit(0)

	} else if code != "200" {
		log.Warnf("Code: %s, Message: %s", code, msg)
		return "", errors.New(fmt.Sprintf("Code: %s, Message: %s", code, msg))
	}
	log.Println("已完成手机客户端确认")
	return gjson.Get(body, "ticket").String(), nil
}

func (this *User) TicketInfo(ticket string) (string, error) {
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Referer", "https://passport.jd.com/uc/login?ltype=logout")
	resp, body, err := req.SetUrl("https://passport.jd.com/uc/qrCodeTicketValidation?t=" + ticket).SetMethod("get").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Error("二维码信息校验失败")
		return "", errors.New("二维码信息校验失败")
	}
	if gjson.Get(body, "returnCode").Int() == 0 {
		log.Info("二维码信息校验成功")
		return "", nil
	} else {
		log.Error("二维码信息校验失败")
		return "", errors.New("二维码信息校验失败")
	}
}

func (this *User) RefreshStatus() error {
	client := httpc.NewHttpClient()
	client.SetCookieJar(common.CookieJar)
	client.SetRedirect(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})
	req := httpc.NewRequest(client)
	req.SetHeader("User-Agent", this.getUserAgent())
	resp, _, err := req.SetUrl("https://order.jd.com/center/list.action?rid=" + strconv.Itoa(int(time.Now().Unix()*1000))).SetMethod("get").Send().End()
	if err == nil && resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New("登录失效")
	}
}

func (this *User) GetUserInfo() (string, error) {
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Referer", "https://order.jd.com/center/list.action")
	errorCount := 5
	nickName := ""
	for {
		if errorCount > 0 {
			_, body, _ := req.SetUrl("https://passport.jd.com/user/petName/getUserInfoForMiniJd.action?callback=" + strconv.Itoa(common.Rand(1000000, 9999999)) + "&_=" + strconv.Itoa(int(time.Now().Unix()*1000))).SetMethod("get").Send().End()
			if gjson.Get(body, "nickName").Exists() {
				nickName = gjson.Get(body, "nickName").String()
				break
			}
			errorCount = errorCount - 1
			time.Sleep(300 * time.Millisecond)
		} else {
			break
		}
	}
	b, _ := common.GbkToUtf8([]byte(nickName))
	return string(b), nil
}

func (this *User) DelQrCode() {
	log.Debug("ViewQrcodePid = ", common.ViewQrcodePid)
	if common.ViewQrcodePid > 0 {
		var err error
		if runtime.GOOS == "windows" {
			err = exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(common.ViewQrcodePid)).Run()
		} else if runtime.GOOS == "darwin" {
			//TODO:MacOS下获取到进程ID不对，直接杀所有进程
			err = exec.Command("pkill", "-f", "Preview").Run()
		} else {
			err = exec.Command("kill", "-9", fmt.Sprint(common.ViewQrcodePid)).Run()
		}
		if err != nil {
			log.Warn(err)
		}
	}
	qrPath := filepath.Join(common.SoftDir, "/qr_code.png")
	if common.Exists(qrPath) {
		_ = os.Remove(qrPath)
	}
}
