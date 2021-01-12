package jd_seckill

import (
	"errors"
	"fmt"
	"github.com/Albert-Zhan/httpc"
	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"github.com/unknwon/goconfig"
	"github.com/ztino/jd_seckill/common"
	"github.com/ztino/jd_seckill/log"
	"github.com/ztino/jd_seckill/service"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Seckill struct {
	client *httpc.HttpClient
	conf   *goconfig.ConfigFile
	initInfo string
}

func NewSeckill(client *httpc.HttpClient, conf *goconfig.ConfigFile) *Seckill {
	return &Seckill{client: client, conf: conf,initInfo: ""}
}

func (this *Seckill) SetInitInfo(initInfo string) {
	this.initInfo=initInfo
}

func (this *Seckill) GetInitInfo() string {
	return this.initInfo
}

func (this *Seckill) getUserAgent() string {
	return this.conf.MustValue("config", "default_user_agent", "")
}

func (this *Seckill) SkuTitle() (string, error) {
	skuId := this.conf.MustValue("config", "sku_id", "")
	req := httpc.NewRequest(this.client)
	resp, body, err := req.SetUrl(fmt.Sprintf("https://item.jd.com/%s.html", skuId)).SetMethod("get").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Error("访问商品详情失败")
		return "", errors.New("访问商品详情失败")
	}
	html := strings.NewReader(body)
	doc, _ := goquery.NewDocumentFromReader(html)
	return strings.TrimSpace(doc.Find(".sku-name").Text()), nil
}

func (this *Seckill) GetDiffTime() int64 {
	log.Info("获取本地时间与京东云端时间差")
	localTime := time.Now().UnixNano() / 1e6
	log.Info("本地系统时间：", time.Unix(0, localTime*1e6))

	jdTime := localTime
	req := httpc.NewRequest(common.Client)
	resp, body, err := req.SetUrl("https://a.jd.com//ajax/queryServerData.html").SetMethod("get").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Warn("获取京东服务器时间失败，以本地时间为准")
	} else {
		jdTime = gjson.Get(body, "serverTime").Int()
	}
	log.Info("京东云端时间：", time.Unix(0, jdTime*1e6))

	delayTime := time.Now().UnixNano()/1e6 - localTime
	log.Info("网络请求延时：", delayTime, "ms")

	diffTime := localTime - jdTime + delayTime/2
	log.Info("实际时间误差：", diffTime, "ms (本地时间-京东云端时间+网络请求延时/2)")

	return diffTime
}

func (this *Seckill) GetWareBusiness() ([]string, []string, error) {
	log.Info("获取商品的预约时间、抢购时间")
	skuId := this.conf.MustValue("config", "sku_id", "")                                                    //商品ID
	cat := "12259,12260,9435"                                                                               //分类路径，TODO:适配其他商品时要调整
	area := "16_1303_3484_0"                                                                                //配送至，TODO:适配其他商品时要调整成购买者实际地区
	shopId := "1000085463"                                                                                  //卖家ID
	venderId := "1000085463"                                                                                //供应商ID
	paramJson := "{\"platform2\":\"1\",\"specialAttrStr\":\"p0pp1pppppppppppppppp\",\"skuMarkStr\":\"00\"}" //TODO:不知道干嘛用的?
	num := this.conf.MustValue("config", "seckill_num", "1")                                                //购买数量
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Referer", fmt.Sprintf("https://item.jd.com/%s.html", skuId))
	resp, body, err := req.SetUrl(fmt.Sprintf("https://item-soa.jd.com/getWareBusiness?callback=jQuery%s&skuId=%s&cat=%s&area=%s&shopId=%s&venderId=%s&paramJson=%s&num=%s&_=%s",
		common.RandomNumber(7),
		skuId,
		cat,
		area,
		shopId,
		venderId,
		url.QueryEscape(paramJson),
		num,
		strconv.Itoa(int(time.Now().Unix()*1000)),
	)).SetMethod("get").Send().End()

	var yuyueTimeArr []string
	var buyTimeArr []string
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Error("获取商品详情失败", resp, body, err)
		return yuyueTimeArr, buyTimeArr, errors.New("访问商品详情失败")
	}
	if !gjson.Get(body, "yuyueInfo").Exists() || !gjson.Get(body, "yuyueInfo.yuyueTime").Exists() || !gjson.Get(body, "yuyueInfo.buyTime").Exists() {
		log.Error("获取商品预约信息失败", body)
		return yuyueTimeArr, buyTimeArr, errors.New("获取商品预约信息失败")
	}

	yuyueTime := gjson.Get(body, "yuyueInfo.yuyueTime").String()
	buyTime := gjson.Get(body, "yuyueInfo.buyTime").String()
	log.Debug("预约起止时间：", yuyueTime)
	log.Debug("购买起止时间：", buyTime)

	reg := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s\d{2}:\d{2})`)
	yuyueTimeArr = reg.FindAllString(yuyueTime, 2)
	buyTimeArr = reg.FindAllString(buyTime, 2)

	return yuyueTimeArr, buyTimeArr, nil
}

func (this *Seckill) MakeReserve() {
	yuyueTimeArr, buyTimeArr, err := this.GetWareBusiness()
	if err == nil && len(yuyueTimeArr) == 2 {
		diffTime := this.GetDiffTime()
		loc, _ := time.LoadLocation("Local")
		yuyueTimeBegin, _ := time.ParseInLocation(common.DateTimeFormatStr, yuyueTimeArr[0]+":00", loc)
		yuyueTimeEnd, _ := time.ParseInLocation(common.DateTimeFormatStr, yuyueTimeArr[1]+":59", loc)

		beginTime := yuyueTimeBegin.UnixNano()/1e6 + diffTime
		endTime := yuyueTimeEnd.UnixNano()/1e6 + diffTime

		diffTime = beginTime - time.Now().UnixNano()/1e6
		if diffTime > 0 {
			log.Warn("还没到预约时间，等待", diffTime, "ms 后开始预约")
			time.Sleep(time.Duration(diffTime) * time.Millisecond)
		}

		diffTime = time.Now().UnixNano()/1e6 - endTime
		if diffTime > 0 {
			log.Error("您已经错过预约时间，下次请早！")
			os.Exit(0)
		}
	} else {
		log.Warn("预约起始时间获取失败，立即尝试预约：", err, yuyueTimeArr)
	}

	user := NewUser(this.client, this.conf)
	userInfo, _ := user.GetUserInfo()
	log.Debug("用户:" + userInfo)
	shopTitle, err := this.SkuTitle()
	if err != nil {
		log.Error("获取商品信息失败")
	} else {
		log.Debug("商品名称:" + shopTitle)
	}
	skuId := this.conf.MustValue("config", "sku_id", "")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Referer", fmt.Sprintf("https://item.jd.com/%s.html", skuId))
	resp, body, err := req.SetUrl("https://yushou.jd.com/youshouinfo.action?callback=fetchJSON&sku=" + skuId + "&_=" + strconv.Itoa(int(time.Now().Unix()*1000))).SetMethod("get").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Error("预约商品失败")
	} else {
		reserveUrl := gjson.Get(body, "url").String()
		req = httpc.NewRequest(this.client)
		_, _, _ = req.SetUrl("https:" + reserveUrl).SetMethod("get").Send().End()
		msg := "商品名称《" + shopTitle + "》预约成功，已获得抢购资格 / 您已成功预约过了，无需重复预约！\n\n[我的预约](https://yushou.jd.com/member/qualificationList.action)"
		_ = service.SendMessage(this.conf, "京东秒杀通知", msg)
		log.Debug(msg)

		//更新购买时间
		if len(buyTimeArr) == 2 {
			confFile := common.SoftDir+"/conf.ini"
			cfg, err := goconfig.LoadConfigFile(confFile)
			if err != nil {
				log.Error("配置文件不存在，程序退出")
				os.Exit(0)
			}
			buyTime := buyTimeArr[0] + ":00"
			cfg.SetValue("config", "buy_time", buyTime)
			if err := goconfig.SaveConfigFile(cfg, confFile); err != nil {
				log.Error("保存配置文件失败，请手动修改conf.ini，buy_time =", buyTime)
			}

			log.Warn("下一次抢购开始时间设定已经更新:", buyTime)
		}
	}
}

func (this *Seckill) getSeckillUrl() (string, error) {
	skuId := this.conf.MustValue("config", "sku_id", "")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Host", "itemko.jd.com")
	req.SetHeader("Referer", fmt.Sprintf("https://item.jd.com/%s.html", skuId))
	req.SetUrl("https://itemko.jd.com/itemShowBtn?callback=jQuery" + strconv.Itoa(common.Rand(1000000, 9999999)) + "&skuId=" + skuId + "&from=pc&_=" + strconv.Itoa(int(time.Now().Unix()*1000))).SetMethod("get")
	url := ""
	for {
		_, body, _ := req.Send().End()
		if gjson.Get(body, "url").Exists() && gjson.Get(body, "url").String() != "" {
			url = gjson.Get(body, "url").String()
			break
		}
		log.Error("抢购链接获取失败，稍后自动重试")
		time.Sleep(300 * time.Millisecond)
	}
	url = "https:"+url
	//https://divide.jd.com/user_routing?skuId=8654289&sn=c3f4ececd8461f0e4d7267e96a91e0e0&from=pc
	url = strings.ReplaceAll(url, "divide", "marathon")
	//https://marathon.jd.com/captcha.html?skuId=8654289&sn=c3f4ececd8461f0e4d7267e96a91e0e0&from=pc
	url = strings.ReplaceAll(url, "user_routing", "captcha.html")
	log.Println("抢购链接获取成功:" + url)
	return url, nil
}

func (this *Seckill) RequestSeckillUrl() {
	user := NewUser(this.client, this.conf)
	userInfo, _ := user.GetUserInfo()
	log.Info("用户:" + userInfo)
	shopTitle, err := this.SkuTitle()
	if err != nil {
		log.Error("获取商品信息失败")
	} else {
		log.Info("商品名称:" + shopTitle)
	}
	url, _ := this.getSeckillUrl()
	skuId := this.conf.MustValue("config", "sku_id", "")
	log.Info("访问商品的抢购连接...")
	client:=httpc.NewHttpClient()
	client.SetCookieJar(common.CookieJar)
	client.SetRedirect(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})
	req := httpc.NewRequest(client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Host", "marathon.jd.com")
	req.SetHeader("Referer", fmt.Sprintf("https://item.jd.com/%s.html", skuId))
	_, _, _ = req.SetUrl(url).SetMethod("get").Send().End()
}

func (this *Seckill) SeckillPage() {
	log.Info("访问抢购订单结算页面...")
	skuId := this.conf.MustValue("config", "sku_id", "")
	seckillNum := this.conf.MustValue("config", "seckill_num", "2")
	client:=httpc.NewHttpClient()
	client.SetCookieJar(common.CookieJar)
	client.SetRedirect(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	})
	req := httpc.NewRequest(client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Host", "marathon.jd.com")
	req.SetHeader("Referer", fmt.Sprintf("https://item.jd.com/%s.html", skuId))
	_, _, _ = req.SetUrl("https://marathon.jd.com/seckill/seckill.action?skuId=" + skuId + "&num=" + seckillNum + "&rid=" + strconv.Itoa(int(time.Now().Unix()))).SetMethod("get").Send().End()
}

func (this *Seckill) SeckillInitInfo() (string, error) {
	log.Info("获取秒杀初始化信息...")
	skuId := this.conf.MustValue("config", "sku_id", "")
	seckillNum := this.conf.MustValue("config", "seckill_num", "2")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Host", "marathon.jd.com")
	req.SetData("sku", skuId)
	req.SetData("num", seckillNum)
	req.SetData("isModifyAddress", "false")
	req.SetUrl("https://marathon.jd.com/seckillnew/orderService/pc/init.action").SetMethod("post")
	//尝试获取三次
	errorCount:=3
	errorMsg:=""
	for errorCount > 0 {
		_, body, _ := req.Send().End()
		if body!="null" && gjson.Valid(body) {
			log.Warn("获取秒杀初始化信息成功")
			return body,nil
		}else{
			log.Error("获取秒杀初始化信息失败,返回信息:"+body)
			errorMsg=body
		}
		errorCount=errorCount-1
		time.Sleep(300*time.Millisecond)
	}
	return "", errors.New(errorMsg)
}

func (this *Seckill) SubmitSeckillOrder() bool {
	eid := this.conf.MustValue("config", "eid", "")
	fp := this.conf.MustValue("config", "fp", "")
	skuId := this.conf.MustValue("config", "sku_id", "")
	seckillNum := this.conf.MustValue("config", "seckill_num", "2")
	paymentPwd := this.conf.MustValue("account", "payment_pwd", "")
	//如果提前获取秒杀初始化信息失败，提交订单时自动重试一次
	if this.initInfo=="" {
		initInfo, _ := this.SeckillInitInfo()
		this.SetInitInfo(initInfo)
	}
	if this.initInfo=="" {
		log.Error(fmt.Sprintf("抢购失败，无法获取生成订单的基本信息，接口返回:【%s】", this.initInfo))
		return false
	}
	address := gjson.Get(this.initInfo, "addressList").Array()
	if !gjson.Get(this.initInfo, "addressList").Exists() || len(address)<1 {
		log.Error("抢购失败，返回信息:"+this.initInfo)
		return false
	}
	defaultAddress := address[0]
	isinvoiceInfo := gjson.Get(this.initInfo, "invoiceInfo").Exists()
	invoiceTitle := "-1"
	invoiceContentType := "-1"
	invoicePhone := ""
	invoicePhoneKey := ""
	invoiceInfo := "false"
	if isinvoiceInfo {
		invoiceTitle = gjson.Get(this.initInfo, "invoiceInfo.invoiceTitle").String()
		invoiceContentType = gjson.Get(this.initInfo, "invoiceInfo.invoiceContentType").String()
		invoicePhone = gjson.Get(this.initInfo, "invoiceInfo.invoicePhone").String()
		invoicePhoneKey = gjson.Get(this.initInfo, "invoiceInfo.invoicePhoneKey").String()
		invoiceInfo = "true"
	}
	token := gjson.Get(this.initInfo, "token").String()
	log.Info("提交抢购订单...")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.getUserAgent())
	req.SetHeader("Host", "marathon.jd.com")
	req.SetHeader("Referer", fmt.Sprintf("https://marathon.jd.com/seckill/seckill.action?skuId=%s&num=%s&rid=%d", skuId, seckillNum, int(time.Now().Unix())))
	req.SetData("skuId", skuId)
	req.SetData("num", seckillNum)
	req.SetData("addressId", defaultAddress.Get("id").String())
	req.SetData("yuShou", "true")
	req.SetData("isModifyAddress", "false")
	req.SetData("name", defaultAddress.Get("name").String())
	req.SetData("provinceId", defaultAddress.Get("provinceId").String())
	req.SetData("cityId", defaultAddress.Get("cityId").String())
	req.SetData("countyId", defaultAddress.Get("countyId").String())
	req.SetData("townId", defaultAddress.Get("townId").String())
	req.SetData("addressDetail", defaultAddress.Get("addressDetail").String())
	req.SetData("mobile", defaultAddress.Get("mobile").String())
	req.SetData("mobileKey", defaultAddress.Get("mobileKey").String())
	req.SetData("email", defaultAddress.Get("email").String())
	req.SetData("postCode", "")
	req.SetData("invoiceTitle", invoiceTitle)
	req.SetData("invoiceCompanyName", "")
	req.SetData("invoiceContent", invoiceContentType)
	req.SetData("invoiceTaxpayerNO", "")
	req.SetData("invoiceEmail", "")
	req.SetData("invoicePhone", invoicePhone)
	req.SetData("invoicePhoneKey", invoicePhoneKey)
	req.SetData("invoice", invoiceInfo)
	req.SetData("password", paymentPwd)
	req.SetData("codTimeType", "3")
	req.SetData("paymentType", "4")
	req.SetData("areaCode", "")
	req.SetData("overseas", "0")
	req.SetData("phone", "")
	req.SetData("eid", eid)
	req.SetData("fp", fp)
	req.SetData("token", token)
	req.SetData("pru", "")
	resp, body, err := req.SetUrl("https://marathon.jd.com/seckillnew/orderService/pc/submitOrder.action?skuId=" + skuId).SetMethod("post").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Error("抢购失败，网络错误")
		_ = service.SendMessage(this.conf, "京东秒杀通知", "抢购失败，网络错误")
		return false
	}

	if !gjson.Valid(body) {
		log.Error("抢购失败，返回信息:" + body)
		_ = service.SendMessage(this.conf, "京东秒杀通知", "抢购失败，返回信息:"+body)
		return false
	}
	if gjson.Get(body, "success").Bool() {
		orderId := gjson.Get(body, "orderId").String()
		totalMoney := gjson.Get(body, "totalMoney").String()
		payUrl := "https:" + gjson.Get(body, "pcUrl").String()
		log.Println(fmt.Sprintf("抢购成功，订单号:%s, 总价:%s, 电脑端付款链接:%s", orderId, totalMoney, payUrl))
		_ = service.SendMessage(this.conf, "京东秒杀通知", fmt.Sprintf("抢购成功，订单号:%s, 总价:%s, 电脑端付款链接:%s", orderId, totalMoney, payUrl))
		return true
	} else {
		log.Error("抢购失败，返回信息:" + body)
		_ = service.SendMessage(this.conf, "京东秒杀通知", "抢购失败，返回信息:"+body)
		return false
	}
}
