package mcp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func HandleStreamableHTTP(c *gin.Context) {
	var req RPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON-RPC request"})
		return
	}

	userID := c.GetUint("user_id")
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	ctx := HandlerContext{
		DB:            config.DB,
		UserID:        userID,
		User:          user.(*models.User),
		AccessibleIDs: utils.GetAccessibleAppIDs(c),
	}

	c.JSON(http.StatusOK, dispatchRequest(ctx, &req))
}
