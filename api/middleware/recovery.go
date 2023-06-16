package middleware

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/xuanxuan000/sipserver/m"
)

func Recovery(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, 4<<10)
			length := runtime.Stack(stack, true)
			logrus.Error(string(stack[:length]))
			m.JsonResponse(c, m.StatusSysERR, "服务器错误，请联系管理员")
		}
	}()
	c.Next()
}
