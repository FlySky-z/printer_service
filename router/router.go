package router

import (
	"io/fs"
	"net/http"
	"printer/frontend"
	"printer/handler"

	"github.com/gin-gonic/gin"
)

// 抽出一个函数统一处理 index.html 返回
func serveIndexHTML(c *gin.Context, distFS fs.FS) {
	content, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "index.html not found")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// embed.FS
	distFS, _ := fs.Sub(frontend.Assets(), "dist")
	// 静态文件路由
	assets, _ := fs.Sub(distFS, "assets")
	r.StaticFS("/assets", http.FS(assets))
	// 显式设置 `/` 路由
	r.GET("/", func(c *gin.Context) { serveIndexHTML(c, distFS) })
	r.GET("/vnc", func(c *gin.Context) { serveIndexHTML(c, distFS) })

	// 打印路由
	r.POST("/print", handler.HandlePrint)
	// 预打印路由
	r.POST("/preopen", handler.HandlePreOpenFile)

	// 文件相关路由
	files := r.Group("/files")
	{
		files.GET("", handler.ListFiles)               // 获取文件列表
		files.POST("", handler.UploadFile)             // 上传文件
		files.GET("/:filename", handler.DownloadFile)  // 下载文件
		files.DELETE("/:filename", handler.DeleteFile) // 删除文件
	}

	// WebSocket路由
	r.GET("/websockify", func(c *gin.Context) {
		handler.HandleWebsockifyHTTP(c.Writer, c.Request)
	})

	// VNC连接相关路由
	vnc := r.Group("/api/vnc")
	{
		vnc.GET("/connections", handler.ListVncConnections)            // 获取连接列表
		vnc.POST("/connections", handler.AddVncConnection)             // 添加新连接
		vnc.PUT("/connections/:index", handler.UpdateVncConnection)    // 更新连接
		vnc.DELETE("/connections/:index", handler.DeleteVncConnection) // 删除连接
	}

	return r
}
