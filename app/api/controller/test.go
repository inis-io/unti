package controller

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"inis/app/model"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	url2 "net/url"
	"os"
	"strings"
	"time"
)

type Test struct {
	// 继承
	base
}

// IGET - GET请求本体
func (this *Test) IGET(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"request": this.request,
		"save":    this.save,
		"proxy":   this.proxy,
		"chrome":  this.chrome,
		"tcping":  this.tcping,
		"alipay":  this.alipay,
		"play":    this.play,
		"channel": this.channel,
		"one":     this.one,
		"two":     this.one,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPOST - POST请求本体
func (this *Test) IPOST(ctx *gin.Context) {

	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"save":       this.save,
		"return-url": this.returnUrl,
		"notify-url": this.notifyUrl,
		"request":    this.request,
		"upload":     this.upload,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IPUT - PUT请求本体
func (this *Test) IPUT(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"request": this.request,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// IDEL - DELETE请求本体
func (this *Test) IDEL(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))

	allow := map[string]any{
		"request": this.request,
	}
	err := this.call(allow, method, ctx)

	if err != nil {
		this.json(ctx, nil, facade.Lang(ctx, "方法调用错误：%v", err.Error()), 405)
		return
	}
}

// INDEX - GET请求本体
func (this *Test) INDEX(ctx *gin.Context) {

	// params := this.params(ctx)

	value := "这是一个测试，{name}，{age}，这是第二次测试，{name}，{age}，好的{{ok}}"
	item  := utils.Replace(value, map[string]any{
		"{name}": "李四",
		"{age}":  18,
	})

	this.json(ctx, item, "ok", 200)

	// item := facade.SMS.VerifyCode(params["phone"])
	// if item.Error != nil {
	// 	this.json(ctx, nil, item.Error.Error(), 400)
	// 	return
	// }
	//
	// this.json(ctx, item, "ok", 200)
}

func (this *Test) one(ctx *gin.Context) {

	params := this.params(ctx)

	this.json(ctx, params, "ok", 200)
}

// INDEX - GET请求本体
func (this *Test) upload(ctx *gin.Context) {

	params := this.params(ctx)

	// 上传文件
	file, err := ctx.FormFile("file")
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	// 文件数据
	bytes, err := file.Open()
	if err != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}
	defer func(bytes multipart.File) {
		err := bytes.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(bytes)

	// 文件后缀
	suffix := file.Filename[strings.LastIndex(file.Filename, "."):]
	params["suffix"] = suffix

	item := facade.Storage.Upload(facade.Storage.Path() + suffix, bytes)
	if item.Error != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}

	params["item"] = item

	fmt.Println("url: ", item.Domain + item.Path)

	this.json(ctx, params, "ok", 200)
}

func (this *Test) channel(ctx *gin.Context) {

	params := this.params(ctx, map[string]any{
		"size":  50,
		"count": 1000,
	})

	// 微秒时间戳
	start := time.Now().UnixNano() / 1e6
	// 通过 channel 通道来控制并发
	channel := make(chan int, cast.ToInt(params["size"]))
	defer close(channel)

	lock := utils.Async[[]any]()

	worker := func() any {
		return time.Now().UnixNano() / 1e6
	}

	for i := 0; i < cast.ToInt(params["count"]); i++ {
		lock.Wait.Add(1)
		channel <- 1
		go func() {
			defer lock.Wait.Done()
			result := worker()
			// 记录结果
			if !utils.Is.Empty(result) {
				lock.Mutex.Lock()
				lock.Data = append(lock.Data, result)
				lock.Mutex.Unlock()
			}
			<-channel
		}()
	}

	// 等待所有的 goroutine 执行完毕
	lock.Wait.Wait()
	end := time.Now().UnixNano() / 1e6

	msg := "播放完成，耗时："
	cost := end - start
	if cost > 1000 {
		msg += fmt.Sprintf("%v s", cost/1000)
	} else {
		msg += fmt.Sprintf("%v ms", cost)
	}

	this.json(ctx, len(lock.Data), msg, 200)
}

func (this *Test) play(ctx *gin.Context) {

	url := "https://aweme.snssdk.com/aweme/v1/aweme/stats/"
	devices := []string{
		"iid=1832202648690573&device_id=3028471298139959&ac=wifi&channel=aweGW&aid=1128&app_name=aweme&version_code=220900&version_name=22.9.0&device_platform=android&os=android&ssmix=a&device_type=LIO-AN00&device_brand=samsung&language=zh&os_api=25&os_version=7.1.2&openudid=44b369da537208bd&manifest_version_code=220901&resolution=540*936&dpi=160&update_version_code=22909900&_rticket=1678867737365&package=com.ss.android.ugc.aweme&mcc_mnc=46007&cpu_support64=false&host_abi=armeabi-v7a&ts=1678867736&is_guest_mode=0&app_type=normal&appTheme=light&need_personal_recommend=1&minor_status=0&is_android_pad=0&cdid=2914cbd6-25eb-44f6-8d6d-0c5f2bf17d43&uuid=863064015149984&md=0",
		"iid=2922916354730683&device_id=2166452352465387&ac=wifi&channel=aweGW&aid=1128&app_name=aweme&version_code=220900&version_name=22.9.0&device_platform=android&os=android&ssmix=a&device_type=TAS-AN00&device_brand=samsung&language=zh&os_api=25&os_version=7.1.2&openudid=236d9d4ce5929fb0&manifest_version_code=220901&resolution=540*960&dpi=160&update_version_code=22909900&_rticket=1679037603483&package=com.ss.android.ugc.aweme&mcc_mnc=46007&cpu_support64=false&host_abi=armeabi-v7a&ts=1679037602&is_guest_mode=0&app_type=normal&appTheme=light&need_personal_recommend=1&minor_status=0&is_android_pad=0&cdid=306a845b-4772-4b37-b67b-99f9678ec8c0&uuid=863064014258836&md=0",
		"iid=2623849633154061&device_id=372049818422232&ac=wifi&channel=aweGW&aid=1128&app_name=aweme&version_code=220900&version_name=22.9.0&device_platform=android&os=android&ssmix=a&device_type=SM-G955N&device_brand=samsung&language=zh&os_api=25&os_version=7.1.2&openudid=236d9d4ce5929fb0&manifest_version_code=220901&resolution=540*960&dpi=160&update_version_code=22909900&_rticket=1679037901408&package=com.ss.android.ugc.aweme&mcc_mnc=46007&cpu_support64=false&host_abi=armeabi-v7a&ts=1679037901&is_guest_mode=0&app_type=normal&appTheme=light&need_personal_recommend=1&minor_status=0&is_android_pad=0&cdid=71f7619a-023d-43fd-be06-606f4388aca4&uuid=863064014745881&md=0",
		"iid=1409991616286551&device_id=4259922509839671&ac=wifi&channel=aweGW&aid=1128&app_name=aweme&version_code=220900&version_name=22.9.0&device_platform=android&os=android&ssmix=a&device_type=SM-G977N&device_brand=samsung&language=zh&os_api=25&os_version=7.1.2&openudid=236d9d4ce5929fb0&manifest_version_code=220901&resolution=540*960&dpi=160&update_version_code=22909900&_rticket=1679037998865&package=com.ss.android.ugc.aweme&mcc_mnc=46007&cpu_support64=false&host_abi=armeabi-v7a&ts=1679037998&is_guest_mode=0&app_type=normal&appTheme=light&need_personal_recommend=1&minor_status=0&is_android_pad=0&cdid=9ccd3005-52ce-43e7-ba22-fc4bc5f4707f&uuid=354730014187623&md=0",
		"iid=2148863429314407&device_id=108169847170183&ac=wifi&channel=aweGW&aid=1128&app_name=aweme&version_code=220900&version_name=22.9.0&device_platform=android&os=android&ssmix=a&device_type=SM-G9810&device_brand=samsung&language=zh&os_api=25&os_version=7.1.2&openudid=236d9d4ce5929fb0&manifest_version_code=220901&resolution=540*960&dpi=160&update_version_code=22909900&_rticket=1679038217838&package=com.ss.android.ugc.aweme&mcc_mnc=46007&cpu_support64=false&host_abi=armeabi-v7a&ts=1679038217&is_guest_mode=0&app_type=normal&appTheme=light&need_personal_recommend=1&minor_status=0&is_android_pad=0&cdid=aeabcca1-b5db-4e3d-85ac-6c71c176471a&uuid=351564014019319&md=0",
	}

	// 随机获取一个设备
	device := devices[rand.Intn(len(devices))]
	headers := map[string]any{
		"Accept-Encoding":           "gzip",
		"x-tt-request-tag":          "t=1;n=0",
		"x-tt-dt":                   "AAA4ZBR44A3DDUP5PLLUI6I5P4GSIOQBVA2BDQ2VVEDAY7GNUQJJU3FTIMYGV7V6K6JZPOHAG2DKOWELVMBOH5GSIC2AVIBHZATHM4KVX3VNO3DUGKQ54DDONTMHRIOIWJBBNGOZ5634EMIOS3DJWZA",
		"activity_now_client":       "1678260482710",
		"X-SS-REQ-TICKET":           "1678260481439",
		"passport-sdk-version":      "20372",
		"sdk-version":               "2",
		"x-vc-bdturing-sdk-version": "2.2.1.cn",
		"User-Agent":                "com.ss.android.ugc.aweme/210701 (Linux; U; Android 7.1.2; zh_CN; OPPO R11 Plus; Build/NMF26X;tt-ok/3.10.0.2)",
		// "X-Ladon":"iqTrDnpY5zWAcmuWrK/H5NNcljc+PRULT950UXYMRe7AFfnt",
		// "X-Gorgon":"040490770000ac07c9e2ed63894f72003c8ead93b8c98d8e443f",
		// "X-Khronos":"1678971075",
		// "X-Argus":"vojawVMPPSCO8yJ2soXaxL+ysvwr0ABXetapm6C0pU2iq2WKLRuOOY/is5M5Dw1SzC7Ps+vxYtgjWAw5GbS3XnSEWJ4o5Q03FzqP+b5MMHrbLQRc+fK7+SEBR8Ipq2YG++Z5PBMet+n/fWlldjt/e9fhNucEIRIX8emY+X3+uHwa9ImDM1WbieazMBpRKQcweKM2mEMWp0t9xOOd22XKn8l/",
		"Content-Type":   "application/x-www-form-urlencoded; charset=UTF-8",
		"Content-Length": "326",
		"Host":           "aweme.snssdk.com",
		"Connection":     "Keep-Alive",
		"Cookie":         "store-region=cn-gd; store-region-src=did; install_id=1181290010790424; ttreq=1$3815304baa56cfd1dfccdee6a0e12a413bf962cb; odin_tt=23295945c67d4529b09c34eebfc553b0bb52ba08761026b93be504243eb2dbe8ed716e925128517330f1834a023dbabb74c1519641d3d2869e32377b6144ab401b6a7ec6dfd18d93c39834a4cef8cdb7",
	}
	payload := map[string]any{
		// [重要] - 用户首次安装该应用的时间戳
		"first_install_time": 1678196308,
		// 是否是商业账号，0表示不是，1表示是
		"is_commerce": 0,
		// 数据来源，0表示来自推荐，1表示来自搜索
		"data_from": 0,
		// 是否在当前用户的可见范围内，false表示不在
		"is_in_scope": false,
		// 关注状态，0为未关注，1为已关注
		"follow_status": 0,
		// [重要] - 页面类型，3表示推荐
		"tab_type": 3,
		// 作品类型，0为普通作品，1为广告
		"aweme_type": 0,
		// [重要] - 视频id
		"item_id": 7105677849029315854,
		// 是否同步到其它平台
		"sync_origin": false,
		// 是否是广告
		"is_from_ad": 0,
		// 上一个视频的播放时间
		"pre_item_playtime": utils.Rand.Int(0, 1000),
		// 粉丝状态，0为未关注，1为已关注
		"follower_status": 0,
		// 是否熟悉该用户，0表示不熟悉
		"is_familiar": 0,
		// 上一个视频的热词
		"pre_hot_sentence": []any{},
		// 附近视频的等级
		"nearby_level": 0,
		// 播放时间单位：1秒
		"play_delta": 1,
		// 上一个视频的id
		"pre_item_id": 7105677849029315854,
		// 用户操作时间
		"action_time": time.Now().Unix(),
	}

	sign := utils.Curl(utils.CurlRequest{
		Method: "POST",
		Url:    "http://192.168.110.69:1010/api/douyin/sign",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: utils.Json.Encode(map[string]any{
			"url":     url + "?" + device,
			"headers": headers,
		}),
	}).Send()
	// // 抖音签名算法
	// sign := utils.Curl.Post("http://192.168.110.69:1010/api/douyin/sign", map[string]any{
	// 	"url": url + "?" + device,
	// 	"headers": headers,
	// }, map[string]any{
	// 	"Content-Type": "application/json",
	// })

	// map[string]any to url.Values
	values := url2.Values{}
	for key, val := range payload {
		values.Add(key, cast.ToString(val))
	}

	signs := cast.ToStringMap(sign.Json["data"])
	headers["X-Ladon"] = signs["X-Ladon"]
	headers["X-Khronos"] = signs["X-Khronos"]
	headers["X-Gorgon"] = signs["X-Gorgon"]
	headers["X-Argus"] = signs["X-Argus"]
	// headers["X-SS-STUB"] = strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(utils.Json.Encode(payload)))))

	client := &http.Client{}
	request, err := http.NewRequest("POST", url+"?"+device, nil)
	request.Body = io.NopCloser(strings.NewReader(values.Encode()))

	if err != nil {
		this.json(ctx, nil, err.Error(), 500)
		return
	}

	for key, val := range headers {
		request.Header.Set(key, cast.ToString(val))
	}

	response, err := client.Do(request)
	if err != nil {
		this.json(ctx, nil, err.Error(), 500)
		return
	}

	if response.StatusCode == http.StatusOK {

		var reader io.Reader
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(response.Body)
		case "deflate":
			reader = flate.NewReader(response.Body)
		default:
			reader = response.Body
		}
		body, _ := io.ReadAll(reader)

		this.json(ctx, utils.Json.Decode(string(body)), "ok", 200)
		return
	}
}

// 实现 TCP ping 且支持带端口ping的功能
func (this *Test) tcping(ctx *gin.Context) {

	params := this.params(ctx)

	start := time.Now()

	host := cast.ToString(params["host"])
	_, item := utils.Net.Tcping(host, map[string]any{
		"count": cast.ToInt(params["count"]),
	})

	this.json(ctx, item, fmt.Sprintf("%v 个PING总耗时：%v", params["count"], time.Since(start).String()), 200)
}

func (this *Test) save(ctx *gin.Context) {

	params := this.params(ctx)

	// model.Db.Model(&model.Users{}).Where("id = ?", 2).Find(&params)

	// [22.797ms] [rows:1] SELECT `id`,`account` FROM `uutx_users` WHERE id = 2 AND account = "97783391@qq.com"

	// .Field("id", "account") .Where("account", "97783391@qq.com")
	// data := tool.DB(&model.Users{}).Where([]any{
	// 	[]any{"id", "=", 2},
	// 	[]any{"account", "=", "97783391@qq.com"},
	// }).Field("id", "account").Select()
	// data := tool.DB(&model.Users{}).Where([]any{
	// 	"id", "=", 2,
	// }).Where("account", "97783391@qq.com").Field("id", "account").Select()
	// data := tool.DB(&model.Users{}).Select()

	// data := tool.DB(&model.Users{}).Select([]any{2, 14, 15})
	// data := tool.DB(&model.Users{}).Count()
	// data := tool.DB(&model.Users{}).Column()
	// data := tool.DB(&model.Users{}).Sum("id")
	// data := tool.DB(&model.Users{}).Max("id")
	// var data any
	// data := tool.DB(&model.Users{}).WithTrashed().WhereNull("avatar","login_time").Select()
	// tool.DB(&model.Users{}).Save()
	// data = tool.DB(&model.Device{}).Create(map[string]any{
	// 	"nickname":   "test",
	// })
	// data = tool.DB(&model.Device{}).Where("id", 1).Update(map[string]any{
	// 	"nickname": "兔子呀",
	// })
	// tool.DB(&model.Device{}).Force().Delete([]any{1, 2})
	// tool.DB(&model.Device{}).Create([]map[string]any{
	// 	{"nickname": "test1"},
	// 	{"nickname": "test2"},
	// 	{"nickname": "test3"},
	// })
	// tool.DB(&model.Device{}).Destroy([]any{4, "5"}, true)
	// tool.DB(&model.Device{}).Destroy(3, true)
	// data = tool.DB(&model.Device{}).Select()

	table := model.Users{}
	item := facade.DB.Model(&table).Or([]any{
		[]any{"account", "=", params["account"]},
		[]any{"email", "=", params["account"]},
	}).Find()

	this.json(ctx, map[string]any{
		"item":  item,
		"table": table,
	}, "数据请求成功！", 200)
}

func (this *Test) chrome(ctx *gin.Context) {

	params := this.params(ctx, map[string]any{
		"ip":   "",
		"port": "",
	})

	ip := params["ip"].(string)
	port := params["port"].(string)
	path := params["path"].(string)

	if utils.Is.Empty(path) {
		this.json(ctx, nil, "path不能为空！", 400)
		return
	}

	// 获取当前所在目录
	root, _ := os.Getwd()

	// 禁用chrome headless
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// 禁用chrome headless
		chromedp.Flag("headless", false),
		// 关闭Chrome正受到自动测试软件的控制。
		chromedp.Flag("enable-automation", false),
		// 指定Chrome的用户数据目录 - 放到D盘
		chromedp.UserDataDir(root+"/public/cache/"+path),
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		// 禁止加载拓展
		chromedp.Flag("disable-extensions", false),
		// 加载拓展，多个拓展用逗号隔开
		// chromedp.Flag("load-extension", "F:\\extension\\WebRTC Leak Shield"),
		// 禁用 WEBRTC
		chromedp.Flag("webrtc-ip-handling-policy", "disable_non_proxied_udp"),
		// 设置窗口最大化
		// chromedp.Flag("start-maximized", true),
		// 置顶
		chromedp.Flag("always-on-top", true),
		// 窗口位置放在右半屏
		chromedp.Flag("window-position", cast.ToString(utils.Get.Resolution(0)/2)+",0"),
		// 窗口大小
		chromedp.Flag("window-size", cast.ToString(utils.Get.Resolution(0)/2)+","+cast.ToString(utils.Get.Resolution(1))),
	)

	if !utils.Is.Empty(ip) && !utils.Is.Empty(port) {
		// 设置代理
		opts = append(opts, chromedp.ProxyServer("http://"+ip+":"+port))
	}

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	// 创建浏览器实例
	ctxChorme, _ := chromedp.NewContext(allocCtx)

	err := chromedp.Run(ctxChorme, []chromedp.Action{
		chromedp.Navigate("https://www.douyin.com/"),
		chromedp.WaitVisible(`#root`, chromedp.ByQuery),
	}...)

	if err != nil {
		fmt.Println(err)
	}

	this.json(ctx, nil, "数据请求成功！", 200)
}

func (this *Test) alipay(ctx *gin.Context) {

	// 初始化 BodyMap
	body := make(gopay.BodyMap)
	body.Set("subject", "统一收单下单并支付页面接口")
	body.Set("out_trade_no", uuid.New().String())
	body.Set("total_amount", "0.01")
	body.Set("product_code", "FAST_INSTANT_TRADE_PAY")

	payUrl, err := facade.Alipay().TradePagePay(context.Background(), body)
	if err != nil {
		if bizErr, ok := alipay.IsBizError(err); ok {
			fmt.Println(bizErr)
			return
		}
		fmt.Println(err)
		return
	}

	fmt.Println(payUrl)

	this.json(ctx, payUrl, "数据请求成功！", 200)
}

func (this *Test) returnUrl(ctx *gin.Context) {

	params := this.params(ctx)

	fmt.Println("==================== returnUrl：", params)
}

func (this *Test) notifyUrl(ctx *gin.Context) {

	params := this.params(ctx)

	fmt.Println("==================== notifyUrl：", params)
}

func (this *Test) other(ctx *gin.Context) {
	// id := "2253bb7562-de53-4962-b36e-be869fa22981"
}

// 测试网络请求
func (this *Test) request(ctx *gin.Context) {

	params := this.params(ctx)
	headers := this.headers(ctx)

	buff, _ := io.ReadAll(ctx.Request.Body)

	this.json(ctx, map[string]any{
		"method":  ctx.Request.Method,
		"parse":   params,
		"body":    string(buff),
		"from":    ctx.Request.Form,
		"query":   ctx.Request.URL.Query(),
		"headers": headers,
	}, facade.Lang(ctx, "数据请求成功！"), 200)
}

func (this *Test) post(ctx *gin.Context) {
	this.json(ctx, map[string]any{
		"params":  this.params(ctx),
		"headers": this.headers(ctx),
	}, "数据请求成功！", 200)
}

func (this *Test) put(ctx *gin.Context) {
	this.json(ctx, map[string]any{
		"params":  this.params(ctx),
		"headers": this.headers(ctx),
	}, "数据请求成功！", 200)
}

func (this *Test) del(ctx *gin.Context) {
	this.json(ctx, map[string]any{
		"params":  this.params(ctx),
		"headers": this.headers(ctx),
	}, "数据请求成功！", 200)
}

func (this *Test) proxy(ctx *gin.Context) {

	params := this.params(ctx)

	// ip := "8.218.145.108"
	// port := "10002"
	//
	// // apiKey := "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"
	//
	// // 创建一个代理服务器
	// proxy := func(_ *http.Request) (*url.URL, error) {
	// 	return url.Parse("http://" + ip + ":" + port)
	// }
	//
	// // 创建一个客户端
	// client := &http.Client{
	// 	Transport: &http.Transport{
	// 		Proxy: proxy,
	// 	},
	// }
	//
	// item := utils.Curl.Get("https://myexternalip.com/raw", nil, nil)

	// // 创建一个请求
	// req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", nil)
	// if err != nil {
	// 	fmt.Println("创建 err:", err)
	// 	return
	// }
	//
	// // // 设置请求头
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", "Bearer " + apiKey)
	//
	// // 设置请求体
	// req.Body = io.NopCloser(strings.NewReader(`{
	// 	"prompt": "This is a test",
	// 	"max_tokens": 5,
	// 	"temperature": 0.9,
	// 	"top_p": 1,
	// 	"n": 1,
	// 	"logprobs": null,
	// 	"stop": "\",
	// 	"model": "text-davinci-003",
	// }`))

	// // 发送请求
	// resp, err := client.Do(item.Request)
	// if err != nil {
	// 	fmt.Println("发送 err:", err)
	// 	return
	// }
	//
	// // 读取响应
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("读取 err:", err)
	// 	return
	// }
	//
	// fmt.Println(string(body))

	this.json(ctx, params, "数据请求成功！", 200)
}

// RandFloat64Slice 生成随机的切片 - 随机出来的值要均匀分布，不能偏差太大
func RandFloat64Slice(total float64, count ...int) []float64 {
	if len(count) == 0 {
		count = append(count, 5)
	}
	var slice []float64
	var sum float64
	for i := 0; i < count[0]; i++ {
		slice = append(slice, 0.0)
	}
	for i := 0; i < count[0]; i++ {
		slice[i] = RandFloat64(0.0, total)
		sum += slice[i]
	}
	if sum > total {
		for i := 0; i < count[0]; i++ {
			slice[i] = slice[i] * total / sum
		}
	}
	return slice
}

// RandFloat64 生成区间随机的浮点数
func RandFloat64(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func RandIntSlice(total int, count ...int) []int {

	if len(count) == 0 {
		count = append(count, 5)
	}
	var slice []int
	var sum int
	for i := 0; i < count[0]; i++ {
		slice = append(slice, 0)
	}
	for i := 0; i < count[0]; i++ {
		slice[i] = RandInt(0, total)
		sum += slice[i]
	}
	if sum > total {
		for i := 0; i < count[0]; i++ {
			slice[i] = slice[i] * total / sum
		}
	}
	if IntSum(slice) < total {
		slice[0] += total - IntSum(slice)
	}
	return slice
}

func RandInt(min, max int) int {
	return min + rand.Intn(max-min)
}

func IntSum(slice []int) int {
	var sum int
	for _, v := range slice {
		sum += v
	}
	return sum
}



type TryCatch struct {
	tryFunc     func()
	catchFunc   func(err error)
	finallyFunc func(res any)
}

func (t *TryCatch) Run() {
	defer func() {
		if r := recover(); r != nil {
			t.catchFunc(r.(error))
		}
		if t.finallyFunc != nil {
			t.finallyFunc(nil)
		}
	}()
	t.tryFunc()
	if t.finallyFunc != nil {
		t.finallyFunc(nil)
	}
}

func Try(tryFunc func()) *TryCatch {
	return &TryCatch{tryFunc, nil, nil}
}

func (t *TryCatch) Catch(catchFunc func(err error)) *TryCatch {
	t.catchFunc = catchFunc
	return t
}

func (t *TryCatch) Finally(finallyFunc func(res any)) {
	t.finallyFunc = finallyFunc
	t.Run()
}