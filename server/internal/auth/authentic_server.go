package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/plutodemon/llog"
	"p2p_remote_desk/server/config"
	"p2p_remote_desk/server/lkit"
	"syscall"
)

func Start() {
	router := gin.Default()
	router.Use(customMiddleware())

	userGroup := router.Group("/user")
	userGroup.POST("/login", Login)

	serverConfig := config.GetConfig().Server
	err := router.Run(lkit.GetAddr(serverConfig.Host, serverConfig.AuthPort))
	if err != nil {
		lkit.SigChan <- syscall.SIGTERM
	}
}

// 自定义中间件
func customMiddleware() gin.HandlerFunc {
	handFunc := func(c *gin.Context) {
		llog.Info("发起请求 ip: ", c.ClientIP(), " 请求路径: ", c.Request.URL.Path)
		c.Next()
		// c.Abort()
	}
	return handFunc
}
