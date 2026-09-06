package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LalatinaHub/LatinaApi/internal/api/handler"
	"github.com/LalatinaHub/LatinaApi/internal/api/middleware"
	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/LatinaApi/internal/repository"
	"github.com/LalatinaHub/LatinaApi/internal/service/converter"
	"github.com/LalatinaHub/LatinaApi/internal/service/subscription"
	"github.com/LalatinaHub/LatinaApi/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockUserRepo struct{}

func (m *mockUserRepo) GetActiveUsersGroupedByVPN(ctx context.Context) (map[string][]model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) DeductQuota(ctx context.Context, userID int64, usedBytes int64) (int64, bool, error) {
	return 0, false, nil
}
func (m *mockUserRepo) DeductQuotaBatch(ctx context.Context, usages map[int64]int64) ([]int64, error) {
	return nil, nil
}
func (m *mockUserRepo) CreateUser(ctx context.Context, u *model.User) (int64, error) {
	return 1, nil
}
func (m *mockUserRepo) GetUserByToken(ctx context.Context, token string) (*model.User, error) {
	if token == "tok-123" {
		return &model.User{
			ID:         1,
			Token:      "tok-123",
			Password:   "pass123",
			Expired:    time.Now().Add(24 * time.Hour),
			ServerCode: "SG01",
			Quota:      1000,
			VPN:        "trojan",
		}, nil
	}
	return nil, model.ErrUserNotFound
}
func (m *mockUserRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	if id == 1 {
		return &model.User{
			ID:         1,
			Token:      "tok-123",
			Quota:      1000,
			Expired:    time.Now().Add(24 * time.Hour),
			ServerCode: "SG01",
			VPN:        "trojan",
		}, nil
	}
	return nil, model.ErrUserNotFound
}
func (m *mockUserRepo) CreateUserWithDefaults(ctx context.Context, id int64) (*model.User, error) {
	return &model.User{ID: id, Token: "tok-new"}, nil
}

type mockServerRepo struct{}

func (m *mockServerRepo) GetAll(ctx context.Context) ([]model.Server, error) {
	return nil, nil
}
func (m *mockServerRepo) GetServerByCode(ctx context.Context, code string) (*model.Server, error) {
	return &model.Server{Code: code, Domain: "sg01.example.com", Country: "SG"}, nil
}

type mockProxyRepo struct{}

func (m *mockProxyRepo) GetRelays(ctx context.Context, excludedCountryCodes []string, maxPerCountry int) ([]model.ProxyNode, error) {
	return nil, nil
}
func (m *mockProxyRepo) GetProxiesByFilter(ctx context.Context, filter repository.ProxyFilter) ([]model.ProxyNode, error) {
	return nil, nil
}

type mockKVRepo struct{}

func (m *mockKVRepo) GetAll(ctx context.Context) (map[string]any, error) {
	return nil, nil
}
func (m *mockKVRepo) GetValueByKey(ctx context.Context, key string) (string, error) {
	if key == "apiToken" {
		return "admin-secret", nil
	}
	return "", nil
}

type mockDBAdminService struct{}

func (m *mockDBAdminService) ExecSQL(ctx context.Context, apiToken string, queries []string) ([]int64, error) {
	if apiToken != "admin-secret" {
		return nil, model.ErrUnauthorized
	}
	return []int64{int64(len(queries))}, nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	userRepo := &mockUserRepo{}
	serverRepo := &mockServerRepo{}
	proxyRepo := &mockProxyRepo{}
	kvRepo := &mockKVRepo{}
	convSvc := converter.NewConverterService()

	subService := subscription.NewSubscriptionService(userRepo, serverRepo, proxyRepo, convSvc, "LatinaHub")
	userService := user.NewUserService(userRepo, kvRepo)
	adminService := &mockDBAdminService{}

	return SetupRouter(RouterConfig{
		SubHandler:     handler.NewSubHandler(subService),
		UserHandler:    handler.NewUserHandler(userService),
		AdminHandler:   handler.NewAdminHandler(adminService),
		HealthHandler:  handler.NewHealthHandler(),
		InfoHandler:    handler.NewInfoHandler(),
		ConvertHandler: handler.NewConvertHandler(convSvc),
		RateLimiter:    middleware.NewRateLimiter(100, 100),
		IsProduction:   false,
	})
}

func TestRouter_Endpoints(t *testing.T) {
	router := setupTestRouter()

	// 1. Root
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "LatinaApi")

	// 2. Ping
	req, _ = http.NewRequest(http.MethodGet, "/ping", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pong")

	// 3. Info
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/info", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ip")

	// 4. Sub - Missing token
	req, _ = http.NewRequest(http.MethodGet, "/sub", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 5. Sub - Valid token + Clash User-Agent
	req, _ = http.NewRequest(http.MethodGet, "/sub?token=tok-123", nil)
	req.Header.Set("User-Agent", "Clash.Meta/1.18")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/yaml; charset=utf-8", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Header().Get("Subscription-Userinfo"))
	assert.NotEmpty(t, w.Header().Get("Profile-Title"))
	assert.Contains(t, w.Body.String(), "PROXIES")

	// 6. User - Authorized
	req, _ = http.NewRequest(http.MethodGet, "/user/admin-secret/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "tok-123")

	// 7. User - Unauthorized
	req, _ = http.NewRequest(http.MethodGet, "/user/wrong-secret/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 8. DB Exec - Success
	payload := `{"queries": ["SELECT 1", "SELECT 2"]}`
	req, _ = http.NewRequest(http.MethodPost, "/db/admin-secret/exec", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")

	// 9. Convert - Success
	rawProxy := "trojan://pass@tr.example.com:443?security=tls#Example"
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/convert?format=clash", strings.NewReader(rawProxy))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Example")
}
