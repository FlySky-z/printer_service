package handler

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const uploadDir = "uploads"

// FileInfo 文件信息结构
type FileInfo struct {
	Filename   string `json:"filename"`
	Size       int64  `json:"size"`
	UploadTime string `json:"upload_time"`
}

// UploadFile 处理文件上传
func UploadFile(c *gin.Context) {
	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create upload directory"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// 构建文件保存路径
	filePath := filepath.Join(uploadDir, header.Filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create file"})
		return
	}
	defer dst.Close()

	// 保存文件
	if _, err := dst.ReadFrom(file); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(200, gin.H{
		"message":  "File uploaded successfully",
		"filename": header.Filename,
	})
}

// DownloadFile 处理文件下载
func DownloadFile(c *gin.Context) {
	filename := c.Param("filename")
	filePath := filepath.Join(uploadDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "File not found"})
		return
	}

	c.File(filePath)
}

// ListFiles 获取已上传的文件列表
func ListFiles(c *gin.Context) {
	files := make([]FileInfo, 0)

	// 确保上传目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": "Failed to access upload directory"})
		return
	}

	// 读取目录内容
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read directory"})
		return
	}

	// 获取文件信息
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, FileInfo{
				Filename:   info.Name(),
				Size:       info.Size(),
				UploadTime: info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	c.JSON(200, gin.H{"files": files})
}

// DeleteFile 处理文件删除
func DeleteFile(c *gin.Context) {
	filename := c.Param("filename")
	filePath := filepath.Join(uploadDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "File not found"})
		return
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete file"})
		return
	}

	c.JSON(200, gin.H{
		"message":  "File deleted successfully",
		"filename": filename,
	})
}
