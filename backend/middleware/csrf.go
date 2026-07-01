package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/utils"
)

var csrfSecret string

func InitCSRFProtection() {
	csrfSecret = os.Getenv("CSRF_SECRET")
	if csrfSecret == "" {
		csrfSecret, _ = utils.GenerateRandomString(32)
		debug.Log("CSRF_SECRET not set, using random secret for development")
	}
	debug.Log("CSRF protection initialized")
}

func CSRFProtection() gin.HandlerFunc {
	csrfMiddleware := csrf.Protect(
		[]byte(csrfSecret),
		csrf.Secure(true),
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.Path("/"),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reason := csrf.FailureReason(r)
			debug.Log("CSRF validation failed for %s %s: %v", r.Method, r.URL.Path, reason)

			var errorCode string
			switch reason {
			case csrf.ErrNoToken:
				errorCode = "csrf_token_missing"
			case csrf.ErrBadToken:
				errorCode = "csrf_token_invalid"
			default:
				errorCode = "csrf_validation_failed"
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"` + errorCode + `"}`))
		})),
	)

	return func(c *gin.Context) {
		authMethod := c.GetString("auth_method")
		if authMethod == "app_token" || authMethod == "api_key" || authMethod == "bearer" {
			c.Next()
			return
		}

		responseWritten := false
		wrappedWriter := &csrfResponseWriter{
			ResponseWriter: c.Writer,
			onWrite: func() {
				responseWritten = true
			},
		}

		originalWriter := c.Writer
		c.Writer = wrappedWriter

		csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Set("csrf_token", csrf.Token(r))
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)

		c.Writer = originalWriter

		if responseWritten && c.Writer.Status() == http.StatusForbidden {
			c.Abort()
		}
	}
}

type csrfResponseWriter struct {
	gin.ResponseWriter
	onWrite     func()
	writeCalled bool
}

func (w *csrfResponseWriter) WriteHeader(code int) {
	if !w.writeCalled {
		w.writeCalled = true
		if w.onWrite != nil {
			w.onWrite()
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *csrfResponseWriter) Write(data []byte) (int, error) {
	if !w.writeCalled {
		w.writeCalled = true
		if w.onWrite != nil {
			w.onWrite()
		}
	}
	return w.ResponseWriter.Write(data)
}

func GetCSRFToken(c *gin.Context) string {
	if token, exists := c.Get("csrf_token"); exists {
		if tokenStr, ok := token.(string); ok {
			return tokenStr
		}
	}
	return ""
}
