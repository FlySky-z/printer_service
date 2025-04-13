package services

import (
	"log"
	"net"

	"github.com/gorilla/websocket"
)

type ProxyServer struct {
	buffer  int
	wsConn  *websocket.Conn
	tcpAddr *net.TCPAddr
	tcpConn *net.TCPConn
}

// 新建一个代理服务器实例, 默认缓冲区大小为65536
func NewProxyServer(wsCoon *websocket.Conn, tcpAddr *net.TCPAddr) *ProxyServer {
	proxyServer := &ProxyServer{
		wsConn:  wsCoon,
		tcpAddr: tcpAddr,
		buffer:  65536,
	}
	return proxyServer
}

// SetBuffer 设置缓冲区大小
func (p *ProxyServer) SetBuffer(buffer int) {
	p.buffer = buffer
}

// GetBuffer 获取缓冲区大小
func (p *ProxyServer) GetBuffer() int {
	return p.buffer
}

// Start 启动代理服务器
func (p *ProxyServer) Start() {
	go p.handleWebSocket()
	go p.handleTCP()
}

// Dial 建立TCP连接
func (p *ProxyServer) Dial() error {
	tcpConn, err := net.DialTCP(p.tcpAddr.Network(), nil, p.tcpAddr)
	if err != nil {
		message := "dialing fail: " + err.Error()
		log.Println(message)

		p.wsConn.WriteMessage(websocket.TextMessage, []byte(message))

		return err
	}

	p.tcpConn = tcpConn
	return nil
}

// handleWebSocket 处理WebSocket连接
func (p *ProxyServer) handleWebSocket() {
	for {
		// 读取WebSocket消息
		_, message, err := p.wsConn.ReadMessage()
		if err != nil {
			p.TearDown()
			break
		}

		// 将消息转发到TCP连接
		if _, err := p.tcpConn.Write(message); err != nil {
			log.Println("webSocketToTCP error:", err.Error())
			// 可能没有初始化tcpConn
			p.Dial()
			p.tcpConn.Write(message)
		}
	}
}

// handleTCP 处理TCP连接
func (p *ProxyServer) handleTCP() {
	for {
		// 读取TCP消息
		message := make([]byte, p.buffer)
		n, err := p.tcpConn.Read(message)
		if err != nil {
			p.TearDown()
			break
		}

		// 将消息转发到WebSocket连接
		if err := p.wsConn.WriteMessage(websocket.BinaryMessage, message[:n]); err != nil {
			log.Println("tcpToWebSocket error:", err.Error())
			break
		}
	}
}

func (p *ProxyServer) TearDown() {
	p.tcpConn.Close()
	p.wsConn.Close()
}
