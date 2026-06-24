package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type registerRequest struct {
	RedirectURIs            []string `json:"redirect_uris"`
	ClientName              string   `json:"client_name"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

func HandleRegister(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	if len(req.RedirectURIs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uris is required"})
		return
	}

	if req.ClientName == "" {
		req.ClientName = "unknown"
	}

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	clientID := "ssc_" + hex.EncodeToString(idBytes)

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	clientSecret := hex.EncodeToString(secretBytes)

	client := &Client{
		ID:           clientID,
		Secret:       clientSecret,
		Name:         req.ClientName,
		RedirectURIs: req.RedirectURIs,
		CreatedAt:    time.Now(),
	}
	DefaultStore.AddClient(client)

	c.JSON(http.StatusCreated, gin.H{
		"client_id":                  clientID,
		"client_secret":              clientSecret,
		"client_name":                client.Name,
		"redirect_uris":              client.RedirectURIs,
		"token_endpoint_auth_method": "none",
	})
}
