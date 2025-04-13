package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type VncConnection struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Password string `json:"password,omitempty"`
}

const vncConfigFile = "vnc_connections.json"

func getVncConfigPath() string {
	return filepath.Join("config", vncConfigFile)
}

func loadVncConnections() ([]VncConnection, error) {
	configPath := getVncConfigPath()

	// 确保配置目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, err
	}

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 获取当前主机的本地IP
		localIP := "localhost"
		if addrs, err := net.InterfaceAddrs(); err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
					localIP = ipNet.IP.String()
					break
				}
			}
		}

		defaultConnections := []VncConnection{{
			Name: "本地服务器",
			Url:  fmt.Sprintf("%s:5900", localIP),
		}}
		data, err := json.Marshal(defaultConnections)
		if err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
			return nil, err
		}
		return defaultConnections, nil
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var connections []VncConnection
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil, err
	}

	return connections, nil
}

func saveVncConnections(connections []VncConnection) error {
	data, err := json.Marshal(connections)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(getVncConfigPath(), data, 0644)
}

// ListVncConnections 获取VNC连接列表
func ListVncConnections(c *gin.Context) {
	connections, err := loadVncConnections()
	if err != nil {
		c.JSON(500, gin.H{"error": "加载VNC连接失败"})
		return
	}

	c.JSON(200, connections)
}

// AddVncConnection 添加新的VNC连接
func AddVncConnection(c *gin.Context) {
	var newConnection VncConnection
	if err := c.BindJSON(&newConnection); err != nil {
		c.JSON(400, gin.H{"error": "无效的请求数据"})
		return
	}

	connections, err := loadVncConnections()
	if err != nil {
		c.JSON(500, gin.H{"error": "加载VNC连接失败"})
		return
	}

	connections = append(connections, newConnection)

	if err := saveVncConnections(connections); err != nil {
		c.JSON(500, gin.H{"error": "保存VNC连接失败"})
		return
	}

	c.JSON(200, newConnection)
}

// UpdateVncConnection 更新VNC连接
func UpdateVncConnection(c *gin.Context) {
	index := c.Param("index")
	var updatedConnection VncConnection
	if err := c.BindJSON(&updatedConnection); err != nil {
		c.JSON(400, gin.H{"error": "无效的请求数据"})
		return
	}

	connections, err := loadVncConnections()
	if err != nil {
		c.JSON(500, gin.H{"error": "加载VNC连接失败"})
		return
	}

	// 检查索引是否有效
	idx := 0
	if _, err := fmt.Sscanf(index, "%d", &idx); err != nil || idx < 0 || idx >= len(connections) {
		c.JSON(400, gin.H{"error": "无效的连接索引"})
		return
	}

	connections[idx] = updatedConnection

	if err := saveVncConnections(connections); err != nil {
		c.JSON(500, gin.H{"error": "保存VNC连接失败"})
		return
	}

	c.JSON(200, updatedConnection)
}

// DeleteVncConnection 删除VNC连接
func DeleteVncConnection(c *gin.Context) {
	index := c.Param("index")

	connections, err := loadVncConnections()
	if err != nil {
		c.JSON(500, gin.H{"error": "加载VNC连接失败"})
		return
	}

	// 检查索引是否有效
	idx := 0
	if _, err := fmt.Sscanf(index, "%d", &idx); err != nil || idx < 0 || idx >= len(connections) {
		c.JSON(400, gin.H{"error": "无效的连接索引"})
		return
	}

	// 删除指定索引的连接
	connections = append(connections[:idx], connections[idx+1:]...)

	if err := saveVncConnections(connections); err != nil {
		c.JSON(500, gin.H{"error": "保存VNC连接失败"})
		return
	}

	c.JSON(200, gin.H{"message": "连接已删除"})
}
