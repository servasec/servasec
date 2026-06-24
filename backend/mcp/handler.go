package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func writeSSE(w io.Writer, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func HandleSSE(c *gin.Context) {
	userID := c.GetUint("user_id")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	session := NewSession(userID)
	defer DeleteSession(session.ID)

	writeSSE(c.Writer, "endpoint", "/mcp/message")
	writeSSE(c.Writer, "session_id", session.ID)

	ctx := c.Request.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-session.Ch:
			if !ok {
				return
			}
			writeSSE(c.Writer, "message", string(data))
		}
	}
}

func HandleMessage(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId query parameter required"})
		return
	}

	session := GetSession(sessionID)
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	userID := c.GetUint("user_id")
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}
	accessibleIDs := utils.GetAccessibleAppIDs(c)

	var req RPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON-RPC request"})
		return
	}

	go func() {
		ctx := HandlerContext{
			DB:            config.DB,
			UserID:        userID,
			User:          user.(*models.User),
			AccessibleIDs: accessibleIDs,
		}
		resp := dispatchRequest(ctx, &req)
		data, _ := json.Marshal(resp)
		select {
		case session.Ch <- data:
		default:
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}
