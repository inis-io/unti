package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
	"github.com/unti-io/go-utils/utils"
	"net/http"
	"strings"
	"time"
)

func init() {
	go Hub.run()
}

type Index struct {
	// 继承
	base
}

// Read - GET请求本体
func (this Index) Read(ctx *gin.Context) {
	// 转小写
	method := strings.ToLower(ctx.Param("method"))
	allow := map[string]any{
		//"connect": this.connect,
	}
	_, err := this.call(allow, method, ctx)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"data": nil,
			"msg":  "方法调用错误：" + err.Error(),
			"code": 500,
		})
		return
	}
}

// Connect - socket 连接
func (this Index) Connect(ctx *gin.Context) {

	// 生成客户端ID
	id := guid()

	conn, err := socket.Upgrade(ctx.Writer, ctx.Request, map[string][]string{
		// 客户端ID
		"X-Client-Id": {id},
		// 客户端信息
		"X-Client-info": {"Welcome to inis pro socket service！"},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &client{
		hub:  Hub,
		conn: conn,
		send: make(chan []byte, 256),
		info: &info{
			ID:      id,
			Type:    "connect",
			Content: "连接成功",
		},
	}
	client.hub.connect <- client

	go client.write()
	go client.read()
}

// 客户端消息读取
func (this *client) read() {

	fmt.Println("============= read - socket =============")

	defer func() {
		this.hub.close <- this
		this.conn.Close()
	}()
	this.conn.SetReadLimit(maxMessageSize)
	this.conn.SetReadDeadline(time.Now().Add(pongWait))
	this.conn.SetPongHandler(func(string) error {
		this.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, msg, err := this.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("error: ", err)
			}
			break
		}
		// 有效的JSON数据
		if json.Valid(msg) {
			info := &info{
				ID: this.info.ID,
			}
			item := Json(msg)
			if empty := utils.Is.Empty(item["to"]); empty {
				info.Type = "broadcast"
			} else {
				info.To = cast.ToString(item["to"])
				info.Type = "single"
			}

			// 删掉这个字段
			delete(item, "to")
			info.Content = item

			msg, _ = json.Marshal(info)
			this.hub.notice <- msg
		} else {
			fmt.Println("无效的JSON数据", string(msg))
		}
	}
}

// 客户端消息写入
func (this *client) write() {

	fmt.Println("============= write - socket =============")

	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		this.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-this.send:
			this.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				this.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			next, err := this.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			next.Write(msg)

			// Add queued chat messages to the current websocket message.
			len := len(this.send)
			for i := 0; i < len; i++ {
				next.Write(line)
				next.Write(<-this.send)
			}

			if err := next.Close(); err != nil {
				return
			}
		case <-ticker.C:
			this.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := this.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (hub *hub) run() {
	for {
		select {
		// 注册
		case client := <-hub.connect:
			fmt.Println("开始连接：" + client.info.ID)
			hub.clients[client.info.ID] = client
			// 发送连接成功消息
			client.send <- []byte(`{"type":"connect","content":"连接成功","id":"` + client.info.ID + `"}`)
		// 退出连接
		case client := <-hub.close:
			fmt.Println("客户端退出连接")
			if _, ok := hub.clients[client.info.ID]; ok {
				delete(hub.clients, client.info.ID)
				close(client.send)
			}
		// 通知通信
		case message := <-hub.notice:
			content := Json(message)
			fmt.Println("进入通知通道：", content)
			if empty := utils.Is.Empty(content["type"]); empty || content["type"] == "broadcast" {
				hub.broadcast(message)
			} else if content["type"] == "single" {
				hub.singlecast(message)
			}
		}
		for key, val := range hub.clients {
			fmt.Println("连接：", key, val.conn.RemoteAddr())
		}
		fmt.Println("================== 在线人数", len(hub.clients), " ==================")
	}
}

// 广播消息
func (hub *hub) broadcast(message []byte) {
	fmt.Println("进入广播通道")
	for _, client := range hub.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(hub.clients, client.info.ID)
		}
	}
}

// 单播消息
func (hub *hub) singlecast(message []byte) {
	content := Json(message)
	to := content["to"]
	fmt.Println("进入单播通道:", content)
	if empty := utils.Is.Empty(to); empty {
		// 没有指定发送对象
		to = content["id"]
		fmt.Println("没有指定发送对象，发送给自己")
	}
	for _, client := range hub.clients {
		if client.info.ID == cast.ToString(to) {
			select {
			case client.send <- message:
				fmt.Println("发送给：" + cast.ToString(to))
			default:
				close(client.send)
				delete(hub.clients, client.info.ID)
			}
		}
	}
}
