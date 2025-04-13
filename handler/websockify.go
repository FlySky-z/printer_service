package handler

import (
	"log"
	"net"
	"net/http"
	"printer/services"
	"time"

	"github.com/gorilla/websocket"
)

// WebsockifyConfig 存储websockify代理的配置
type WebsockifyConfig struct {
	// 默认目标TCP服务器地址，如果WebSocket请求未指定，将使用此地址
	DefaultTarget string

	// 最大连接缓冲区大小
	BufferSize int

	// TCP连接超时设置
	DialTimeout time.Duration

	// 读写超时设置
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// 心跳间隔
	HeartbeatInterval time.Duration

	// 是否允许自定义目标地址（通过URL参数）
	AllowCustomTarget bool
}

// DefaultWebsockifyConfig 返回默认配置
func DefaultWebsockifyConfig() WebsockifyConfig {
	return WebsockifyConfig{
		DefaultTarget: "localhost:5900",
		BufferSize:    65536,
	}
}

// Websockify 是websockify代理的主结构体
type Websockify struct {
	config   WebsockifyConfig
	upgrader websocket.Upgrader
}

// AuthenticateOrigin 用于验证WebSocket连接的来源
func authenticateOrigin(r *http.Request) bool {
	return true
}

// NewWebsockify 创建一个新的Websockify实例
func NewWebsockify(config WebsockifyConfig) *Websockify {
	return &Websockify{
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.BufferSize,
			WriteBufferSize: config.BufferSize,
			CheckOrigin:     authenticateOrigin,
			Subprotocols:    []string{"binary"},
		},
	}
}

// HandleWebsockify 处理WebSocket连接请求并建立到TCP服务器的代理
func (ws *Websockify) HandleWebsockify(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("host")
	if addr == "" {
		addr = ws.config.DefaultTarget
		log.Println("未提供主机参数，使用默认TCP服务器地址:", addr)
	}

	log.Printf("开始Websockify连接，目标TCP服务器: %s", addr)
	clientConn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		http.Error(w, "WebSocket连接失败", http.StatusInternalServerError)
		return
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Printf("解析TCP地址失败: %v", err)
		http.Error(w, "无效的TCP地址", http.StatusBadRequest)
		return
	}

	proxyServer := services.NewProxyServer(clientConn, tcpAddr)
	proxyServer.SetBuffer(ws.config.BufferSize)

	if err := proxyServer.Dial(); err != nil {
		log.Printf("TCP连接失败: %v", err)
		clientConn.Close()
		return
	}

	go proxyServer.Start()
}

// HandleWebsockifyHTTP 提供一个HTTP处理函数，使用默认配置
func HandleWebsockifyHTTP(w http.ResponseWriter, r *http.Request) {
	websockify := NewWebsockify(DefaultWebsockifyConfig())
	websockify.HandleWebsockify(w, r)
}
