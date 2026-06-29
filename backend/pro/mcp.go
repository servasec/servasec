package pro

import "github.com/gin-gonic/gin"

type MCPServer interface {
	RegisterRoutes(router *gin.Engine)
}

type noopMCPServer struct{}

func (n *noopMCPServer) RegisterRoutes(router *gin.Engine) {}

var MCP MCPServer = &noopMCPServer{}
