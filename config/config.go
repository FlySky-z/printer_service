package config

// 设置gin为release模式
import "github.com/gin-gonic/gin"

// SetGinMode 设置gin为release模式
func SetGinMode(mode string) {
	switch mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}
