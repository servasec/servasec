package pro

import "github.com/gin-gonic/gin"

type AuditLogger interface {
	Middleware() gin.HandlerFunc
	RegisterRoutes(router *gin.Engine)
}

type noopAuditLogger struct{}

func (n *noopAuditLogger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func (n *noopAuditLogger) RegisterRoutes(router *gin.Engine) {}

var Audit AuditLogger = &noopAuditLogger{}
