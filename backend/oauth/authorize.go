package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
	"golang.org/x/crypto/bcrypt"
)

func codeVerifierToChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateCode() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var loginTmpl = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html>
<head><title>Servasec - Authorize</title></head>
<body style="font-family:sans-serif;max-width:400px;margin:40px auto">
  <h2>Servasec Authorization</h2>
  {{if .Error}}<p style="color:red">{{.Error}}</p>{{end}}
  <form method="post" action="{{.Action}}">
    <p><input name="username" placeholder="Username or email" required style="width:100%;padding:8px"></p>
    <p><input name="password" type="password" placeholder="Password" required style="width:100%;padding:8px"></p>
    <p><button type="submit" style="padding:8px 24px">Log in & authorize</button></p>
  </form>
</body>
</html>`))

func HandleAuthorize(c *gin.Context) {
	q := c.Request.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	responseType := q.Get("response_type")
	challenge := q.Get("code_challenge")
	challengeMethod := q.Get("code_challenge_method")
	state := q.Get("state")

	if clientID == "" || redirectURI == "" || responseType == "" || challenge == "" {
		c.String(http.StatusBadRequest, "missing required OAuth parameters")
		return
	}

	if responseType != "code" {
		c.String(http.StatusBadRequest, "unsupported response_type")
		return
	}

	if challengeMethod != "" && challengeMethod != "S256" {
		c.String(http.StatusBadRequest, "unsupported code_challenge_method")
		return
	}

	client := DefaultStore.GetClient(clientID)
	if client == nil {
		c.String(http.StatusBadRequest, "invalid client_id")
		return
	}

	validRedirect := false
	for _, u := range client.RedirectURIs {
		if u == redirectURI {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		c.String(http.StatusBadRequest, "invalid redirect_uri")
		return
	}

	callbackURL := redirectURI + "?code=%s&state=%s"

	if c.Request.Method == "POST" {
		username := strings.TrimSpace(c.PostForm("username"))
		password := c.PostForm("password")

		var user models.User
		if err := config.DB.Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil {
			renderLogin(c, q, "Invalid credentials")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			renderLogin(c, q, "Invalid credentials")
			return
		}
		if user.Banned {
			renderLogin(c, q, "Account is banned")
			return
		}

		code := generateCode()
		DefaultStore.AddCode(&AuthCode{
			Code:            code,
			ClientID:        clientID,
			UserID:          user.ID,
			RedirectURI:     redirectURI,
			ExpiresAt:       time.Now().Add(5 * time.Minute),
			Challenge:       challenge,
			ChallengeMethod: challengeMethod,
		})

		loc := fmt.Sprintf(callbackURL, url.QueryEscape(code), url.QueryEscape(state))
		c.Redirect(http.StatusFound, loc)
		return
	}

	claims, _ := utils.GetClaimsFromCookie(c)
	if claims != nil {
		var user models.User
		if err := config.DB.First(&user, claims.UserID).Error; err == nil && !user.Banned {
			code := generateCode()
			DefaultStore.AddCode(&AuthCode{
				Code:            code,
				ClientID:        clientID,
				UserID:          user.ID,
				RedirectURI:     redirectURI,
				ExpiresAt:       time.Now().Add(5 * time.Minute),
				Challenge:       challenge,
				ChallengeMethod: challengeMethod,
			})
			loc := fmt.Sprintf(callbackURL, url.QueryEscape(code), url.QueryEscape(state))
			c.Redirect(http.StatusFound, loc)
			return
		}
	}

	renderLogin(c, q, "")
}

func renderLogin(c *gin.Context, q url.Values, errMsg string) {
	action := "/oauth/authorize?" + q.Encode()
	c.Header("Content-Type", "text/html; charset=utf-8")
	loginTmpl.Execute(c.Writer, map[string]any{
		"Action": action,
		"Error":  errMsg,
	})
}
