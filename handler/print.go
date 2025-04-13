package handler

import (
	"os"
	"path/filepath"
	"printer/services"

	"github.com/gin-gonic/gin"
)

// HandlePrint 处理打印请求
func HandlePrint(c *gin.Context) {
	// 从JSON body中获取filename
	var reqBody struct {
		Filename string `json:"filename"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(400, gin.H{"error": "无效的请求格式"})
		return
	}

	if reqBody.Filename == "" {
		c.JSON(400, gin.H{"error": "文件名不能为空"})
		return
	}

	// 构建完整的文件路径
	filePath := filepath.Join("uploads", reqBody.Filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "文件不存在"})
		return
	}

	// 创建打印服务实例
	printService := &services.PrintService{}

	// 执行打印
	err := printService.PrintFile(filePath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "打印成功"})
}

func HandlePreOpenFile(c *gin.Context) {
	// 从JSON body中获取filename
	var reqBody struct {
		Filename string `json:"filename"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(400, gin.H{"error": "无效的请求格式"})
		return
	}

	if reqBody.Filename == "" {
		c.JSON(400, gin.H{"error": "文件名不能为空"})
		return
	}

	// 构建完整的文件路径
	filePath := filepath.Join("uploads", reqBody.Filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "文件不存在"})
		return
	}

	// 创建打印服务实例
	printService := &services.PrintService{}

	// 执行预打开
	err := printService.OpenFile(filePath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "成功"})
}
