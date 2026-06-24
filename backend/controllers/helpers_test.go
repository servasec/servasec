package controllers

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/middleware"
	"github.com/servasec/servasec/backend/testutil"
	"github.com/servasec/servasec/backend/utils"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-jwt-secret-for-testing-only")
	os.Setenv("REFRESH_SECRET", "test-refresh-secret-for-testing-only")
	utils.AccessSecret = []byte(os.Getenv("JWT_SECRET"))
	utils.RefreshSecret = []byte(os.Getenv("REFRESH_SECRET"))

	gin.SetMode(gin.TestMode)

	testutil.SetupTestDB()
	testutil.SeedTestUser(config.DB, "testuser", "test@example.com", "Test1234!", "member")
	testutil.SeedTestUser(config.DB, "adminuser", "admin@example.com", "Admin1234!", "admin")

	code := m.Run()
	os.Exit(code)
}

func setupTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	auth := r.Group("/")
	auth.Use(middleware.AuthRequired())
	{
		auth.GET("/me", GetCurrentUser)
		auth.POST("/logout", Logout)
		auth.PATCH("/me", UpdateCurrentUser)
		auth.PUT("/me/password", UpdateCurrentUserPassword)
	}

	r.POST("/register", Register)
	r.POST("/login", Login)
	r.POST("/refresh", Refresh)

	return r
}

func parseCookies(headers http.Header) map[string]string {
	cookies := make(map[string]string)
	for _, line := range headers.Values("Set-Cookie") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			valueParts := strings.SplitN(parts[1], ";", 2)
			cookies[parts[0]] = strings.TrimSpace(valueParts[0])
		}
	}
	return cookies
}

func setCookieHeader(req *http.Request, name, value string) {
	req.Header.Add("Cookie", name+"="+value)
}
