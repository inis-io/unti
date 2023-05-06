package controller

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"inis/app/facade"
	"mime/multipart"
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

}

// 测试网络请求
func (this *Test) request(ctx *gin.Context) {

	this.json(ctx, map[string]any{
		"method" : ctx.Request.Method,
		"params" : this.params(ctx),
		"headers": this.headers(ctx),
	}, facade.Lang(ctx, "数据请求成功！"), 200)
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

	var item *facade.StorageResponse

	item = facade.Storage.Upload(facade.Storage.Path() + suffix, bytes)
	if item.Error != nil {
		this.json(ctx, nil, err.Error(), 400)
		return
	}
	fmt.Println("文件上传成功：", item.Domain + item.Path)

	item = facade.NewStorage("oss").Upload(facade.NewStorage("oss").Path() + suffix, bytes)
	fmt.Println("文件上传成功：", item.Domain + item.Path)

	item = facade.NewStorage(facade.StorageModeCOS).Upload(facade.NewStorage(facade.StorageModeCOS).Path() + suffix, bytes)
	fmt.Println("文件上传成功：", item.Domain + item.Path)

	this.json(ctx, item.Domain + item.Path, "ok", 200)
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