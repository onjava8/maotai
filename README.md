helloworld
=======

## 特别声明:

* 本仓库发布的`helloworld`项目中涉及的任何脚本，仅用于测试和学习研究，禁止用于商业用途，不能保证其合法性，准确性，完整性和有效性，请根据情况自行判断。

* 本项目内所有资源文件，禁止任何公众号、自媒体进行任何形式的转载、发布。

* `ztino` 对任何脚本问题概不负责，包括但不限于由任何脚本错误导致的任何损失或损害.

* 间接使用脚本的任何用户，包括但不限于建立VPS或在某些行为违反国家/地区法律或相关法规的情况下进行传播, `ztino` 对于由此引起的任何隐私泄漏或其他后果概不负责。

* 请勿将`helloworld`项目的任何内容用于商业或非法目的，否则后果自负。

* 如果任何单位或个人认为该项目的脚本可能涉嫌侵犯其权利，则应及时通知并提供身份证明，所有权证明，我们将在收到认证文件后删除相关脚本。

* 以任何方式查看此项目的人或直接或间接使用`helloworld`项目的任何脚本的使用者都应仔细阅读此声明。`ztino` 保留随时更改或补充此免责声明的权利。一旦使用并复制了任何相关脚本或`helloworld`项目，则视为您已接受此免责声明。
  
* 您必须在下载后的24小时内从计算机或手机中完全删除以上内容。  
  
* 本项目遵循`GPL-3.0 License`协议，如果本特别声明与`GPL-3.0 License`协议有冲突之处，以本特别声明为准。

> ***您使用或者复制了本仓库且本人制作的任何代码或项目，则视为`已接受`此声明，请仔细阅读***  
> ***您在本声明未发出之时点使用或者复制了本仓库且本人制作的任何代码或项目且此时还在使用，则视为`已接受`此声明，请仔细阅读***

> ⚠ 此项目是[python jd_seckill](https://github.com/huanghyw/jd_seckill) 的go版本实现，旨在降低使用门槛和相互学习而创建。

**go版本的jd_seckill，支持跨平台，使用者请在发布页下载可执行文件，欢迎pr。**

## 支持系统

>目前编译好的可执行文件有Windows,MacOS,Linux,arm,mips平台。

## 安装(开发者)

方式一(推荐):

```shell
git clone https://github.com/ztino/jd_seckill.git
cd jd_seckill
go get
```

方式二:

```shell
go get github.com/ztino/jd_seckill
```

## 使用

> [下载](https://github.com/ztino/jd_seckill/releases) 对应平台的可执行文件，解压，终端进入该目录。

### 登录
执行以下命令按照提示操作:
```shell
jd_seckill login
```

### 自动获取eid,fp

> ⚠依赖谷歌浏览器，请安装谷歌浏览器，windows下请将安装目录加入系统变量Path

执行以下命令按照提示操作:
```shell
#参数--good_url商品链接必须设置，链接地址是一个可以加入购物车的商品
jd_seckill jdTdudfp --good_url https://item.jd.com/100007959916.html
```
> ⚠获取成功后会将获取到的eid和fp写入到配置文件中

### 预约
执行以下命令按照提示操作:
```shell
jd_seckill reserve
```

### 抢购
执行以下命令按照提示操作:
```shell
#支持--run参数，将跳过抢购等待时间，直接执行抢购任务，适合10点左右未设置抢购时间的使用
jd_seckill seckill
```

### 退出登录
```shell
jd_seckill logout
```

### 获取版本号
```shell
jd_seckill version
```

> ⚠ 以上命令并不是每次都需要执行的，都是可选的，具体使用请参考提示。

## 使用教程

#### 1. 推荐Chrome浏览器
#### 2. 网页扫码登录，或者账号密码登录
#### 3. 填写config.ini配置信息

> ⚠ 按照下方获取不到的，可以点击进入付款界面(输入支付密码页面)，尝试下方步骤进行获取

(1)`eid`和`fp`找个普通商品随便下单,然后抓包就能看到,这两个值可以填固定的
> 随便找一个商品下单，然后进入结算页面，打开浏览器的调试窗口，切换到控制台Tab页，在控制台中输入变量`_JdTdudfp`，即可从输出的Json中获取`eid`和`fp`。  
> 不会的话参考issue https://github.com/ztino/jd_seckill/issues/2

(2)`sku_id`,`default_user_agent`
> `sku_id`已经按照茅台的填好。
> `default_user_agent` 可以用默认的。谷歌浏览器也可以浏览器地址栏中输入about:version 查看`USER_AGENT`替换

(3)配置一下时间
> 现在不强制要求同步最新时间了，程序会自动同步京东时间
> 但要是电脑时间快慢了好几分钟的，最好还是同步一下吧

以上都是必须的.
> tips：
> 在程序开始运行后，会检测本地时间与京东服务器时间，输出的差值为本地时间-京东服务器时间，即-50为本地时间比京东服务器时间慢50ms。
> 本代码的执行的抢购时间以本地电脑/服务器时间为准

> ⚠ 京东每月限购两瓶，如果本月已抢到两瓶，一个月后再抢吧，有的抢到1瓶的，使用脚本记得需要修改参数

(4)修改抢购瓶数
> 可在配置文件中找到seckill_num进行修改，默认值2瓶

(5)抢购总时间
> 可在配置文件中找到seckill_time进行修改，单位:分钟，默认两分钟

(6)抢购任务数量
> 可在配置文件中找到task_num进行修改，默认5个

(7)每次抢购间隔时间
> 可在配置文件中找到ticker_time进行修改，单位:毫秒，默认1500毫秒，每1000毫秒等于1秒

(8)通知配置
> 目前支持email，wechat，dingtalk，具体可查看配置文件

## Linux 无图形界面获取 eid 与 fp 方法参考
(1) 安装无头 chrome
```shell
sudo apt install ./google-chrome-stable_current_amd64.deb
sudo apt-get -y install xorg xvfb gtk2-engines-pixbuf
sudo apt-get -y install dbus-x11 xfonts-base xfonts-100dpi xfonts-75dpi xfonts-cyrillic xfonts-scalable
sudo apt-get install -y xvfb

Xvfb -ac :99 -screen 0 1280x1024x16 & export DISPLAY=:99
```
(2) 执行获取 eid 与 fp
```shell
#参数--good_url商品链接必须设置，链接地址是一个可以加入购物车的商品
jd_seckill jdTdudfp --good_url https://item.jd.com/100007959916.html
```
## docker 运行
（1）构建镜像
```shell
docker-compose build
```
(2) 运行
修改 conf.ini 配置文件后，直接运行 TODO: 暂不知环境变量
```shell
docker-compose up -d
```
随后界面上会打印二维码，扫描登陆完成即可：
```shell
docker-compose logs -f
```

## 感谢
##### 非常感谢原作者 https://github.com/zhou-xiaojun/jd_mask 提供的代码
##### 也非常感谢 https://github.com/wlwwu/jd_maotai 进行的优化
