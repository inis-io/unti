package controller

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var socket = websocket.Upgrader{
	// 读取存储空间大小
	ReadBufferSize: 1024,
	// 写入存储空间大小
	WriteBufferSize: 1024,
	// 允许跨域
	CheckOrigin: func(request *http.Request) bool {
		return true
	},
}

// 客户端是 websocket 连接和集线器之间的中间人。
type client struct {
	hub *hub
	// 客户端信息
	info *info
	// 出站消息的缓冲通道。
	send chan []byte
	// websocket 连接。
	conn *websocket.Conn
}

// 客户端信息结构体
type info struct {
	ID      string `json:"id"`
	To      string `json:"to"`
	Type    string `json:"type"`
	Content any    `json:"data"`
}

// Hub 维护活跃客户端的集合，并将消息广播
type hub struct {
	// 注册的客户。
	clients map[string]*client
	// 来自客户端的入站消息。
	notice chan []byte
	// 注册来自客户端的请求。
	connect chan *client
	// 取消注册来自客户端的请求。
	close chan *client
}

var Hub = func() *hub {
	return &hub{
		notice:  make(chan []byte),
		connect: make(chan *client),
		close:   make(chan *client),
		clients: make(map[string]*client),
	}
}()

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	line  = []byte{'\n'}
	space = []byte{' '}
)
