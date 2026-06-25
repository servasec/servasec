package sso

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const stateCookieName = "oauth_state"
const stateCookieMaxAge = 600
const stateLen = 32

func generateState() (string, error) {
	b := make([]byte, stateLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func setStateCookie(c *gin.Context, state string) {
	c.SetCookie(stateCookieName, state, stateCookieMaxAge, "/api/auth/", "", true, true)
}

func getStateCookie(c *gin.Context) (string, error) {
	state, err := c.Cookie(stateCookieName)
	return state, err
}

func clearStateCookie(c *gin.Context) {
	c.SetCookie(stateCookieName, "", -1, "/api/auth/", "", true, true)
}

func redirectError(c *gin.Context, msg string) {
	clearStateCookie(c)
	c.Redirect(http.StatusFound, PublicURL()+"/login?error="+msg)
}

func redirectSuccess(c *gin.Context) {
	clearStateCookie(c)
	c.Redirect(http.StatusFound, PublicURL()+"/")
}


