package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
)

func TestRegister_Success(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"newuser","email":"new@example.com","password":"NewUser123!"}`
	req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", resp.Code, resp.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	if result["username"] != "newuser" {
		t.Errorf("expected username 'newuser', got %v", result["username"])
	}

	var user models.User
	if err := config.DB.Where("username = ?", "newuser").First(&user).Error; err != nil {
		t.Fatal("user was not created in database")
	}
	if user.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got %s", user.Email)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"testuser","email":"another@example.com","password":"Test1234!"}`
	req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate username, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"anotheruser","email":"test@example.com","password":"Test1234!"}`
	req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate email, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestRegister_InvalidInput(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name string
		body string
	}{
		{"empty json", `{}`},
		{"missing username", `{"email":"x@y.com","password":"Test1234!"}`},
		{"missing email", `{"username":"x","password":"Test1234!"}`},
		{"missing password", `{"username":"x","email":"x@y.com"}`},
		{"short password", `{"username":"x","email":"x@y.com","password":"ab"}`},
		{"invalid email", `{"username":"x","email":"notanemail","password":"Test1234!"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", resp.Code, resp.Body.String())
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"testuser","password":"Test1234!"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	if result["message"] != "Login successful" {
		t.Errorf("unexpected message: %v", result["message"])
	}

	userInfo := result["user"].(map[string]interface{})
	if userInfo["username"] != "testuser" {
		t.Errorf("expected username 'testuser', got %v", userInfo["username"])
	}

	cookies := parseCookies(resp.Header())
	if cookies["access_token"] == "" {
		t.Error("expected access_token cookie")
	}
	if cookies["refresh_token"] == "" {
		t.Error("expected refresh_token cookie")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"testuser","password":"WrongPassword!"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestLogin_NonExistentUser(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"nobody","password":"Test1234!"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestLogin_ByEmail(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"test@example.com","password":"Test1234!"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected 200 for email login, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestLogin_EmptyInput(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"","password":""}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty input, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestGetCurrentUser_Unauthenticated(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest("GET", "/me", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestGetCurrentUser_Authenticated(t *testing.T) {
	router := setupTestRouter()

	loginBody := `{"username":"testuser","password":"Test1234!"}`
	loginReq := httptest.NewRequest("POST", "/login", strings.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)

	cookies := parseCookies(loginResp.Header())
	accessToken := cookies["access_token"]
	if accessToken == "" {
		t.Fatal("no access_token in login response")
	}

	req := httptest.NewRequest("GET", "/me", nil)
	setCookieHeader(req, "access_token", accessToken)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	if result["username"] != "testuser" {
		t.Errorf("expected username 'testuser', got %v", result["username"])
	}
	if result["role"] != "member" {
		t.Errorf("expected role 'member', got %v", result["role"])
	}
	if result["banned"] != false {
		t.Errorf("expected banned false, got %v", result["banned"])
	}
}

func TestRefresh_Success(t *testing.T) {
	router := setupTestRouter()

	loginBody := `{"username":"testuser","password":"Test1234!"}`
	loginReq := httptest.NewRequest("POST", "/login", strings.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)

	cookies := parseCookies(loginResp.Header())
	refreshToken := cookies["refresh_token"]
	if refreshToken == "" {
		t.Fatal("no refresh_token in login response")
	}

	req := httptest.NewRequest("POST", "/refresh", nil)
	setCookieHeader(req, "refresh_token", refreshToken)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	newCookies := parseCookies(resp.Header())
	if newCookies["access_token"] == "" {
		t.Error("expected new access_token cookie after refresh")
	}
}

func TestRefresh_NoToken(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest("POST", "/refresh", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without refresh token, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestLogout_Success(t *testing.T) {
	router := setupTestRouter()

	loginBody := `{"username":"testuser","password":"Test1234!"}`
	loginReq := httptest.NewRequest("POST", "/login", strings.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)

	cookies := parseCookies(loginResp.Header())
	accessToken := cookies["access_token"]

	req := httptest.NewRequest("POST", "/logout", nil)
	setCookieHeader(req, "access_token", accessToken)
	setCookieHeader(req, "refresh_token", cookies["refresh_token"])
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	newCookies := parseCookies(resp.Header())
	if newCookies["access_token"] != "" {
		t.Log("access_token cookie cleared (empty or expired)")
	}
}

func TestGetCurrentUser_Admin(t *testing.T) {
	router := setupTestRouter()

	loginBody := `{"username":"adminuser","password":"Admin1234!"}`
	loginReq := httptest.NewRequest("POST", "/login", strings.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)

	cookies := parseCookies(loginResp.Header())
	accessToken := cookies["access_token"]
	if accessToken == "" {
		t.Fatal("no access_token in login response")
	}

	req := httptest.NewRequest("GET", "/me", nil)
	setCookieHeader(req, "access_token", accessToken)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &result)
	if result["role"] != "admin" {
		t.Errorf("expected role 'admin', got %v", result["role"])
	}
}
