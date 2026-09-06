package subscription

import (
	"context"
	"testing"
	"time"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/LatinaApi/internal/repository"
	"github.com/LalatinaHub/LatinaApi/internal/service/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	user *model.User
	err  error
}

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
	return m.user, m.err
}
func (m *mockUserRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) CreateUserWithDefaults(ctx context.Context, id int64) (*model.User, error) {
	return m.user, m.err
}

type mockServerRepo struct {
	server *model.Server
	err    error
}

func (m *mockServerRepo) GetAll(ctx context.Context) ([]model.Server, error) {
	return nil, nil
}
func (m *mockServerRepo) GetServerByCode(ctx context.Context, code string) (*model.Server, error) {
	return m.server, m.err
}

type mockProxyRepo struct {
	proxies []model.ProxyNode
	err     error
}

func (m *mockProxyRepo) GetRelays(ctx context.Context, excludedCountryCodes []string, maxPerCountry int) ([]model.ProxyNode, error) {
	return nil, nil
}
func (m *mockProxyRepo) GetProxiesByFilter(ctx context.Context, filter repository.ProxyFilter) ([]model.ProxyNode, error) {
	return m.proxies, m.err
}

func TestSubscriptionService_SuccessCases(t *testing.T) {
	now := time.Now()
	validUser := &model.User{
		ID:         1,
		Token:      "tok-valid",
		Password:   "pass123",
		Expired:    now.Add(24 * time.Hour),
		ServerCode: "SG01",
		Quota:      500, // 500 MB
		VPN:        "trojan",
	}

	server := &model.Server{
		ID:         1,
		Code:       "SG01",
		Domain:     "sg01.foolvpn.me",
		IP:         "1.2.3.4",
		Country:    "SIN",
		UsersCount: 10,
		UsersMax:   100,
	}

	userRepo := &mockUserRepo{user: validUser}
	serverRepo := &mockServerRepo{server: server}
	proxyRepo := &mockProxyRepo{}
	convSvc := converter.NewConverterService()

	svc := NewSubscriptionService(userRepo, serverRepo, proxyRepo, convSvc, "LatinaHub")
	ctx := context.Background()

	// 1. Clash User-Agent
	resClash, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:     "tok-valid",
		UserAgent: "Clash.Meta/v1.18.0",
	})
	require.NoError(t, err)
	assert.Equal(t, "application/yaml; charset=utf-8", resClash.ContentType)
	assert.Equal(t, "config.yaml", resClash.Filename)
	assert.Contains(t, resClash.Content, "PROXIES")
	assert.Contains(t, resClash.UserInfo, "total=524288000") // 500 * 1024 * 1024

	// 2. sing-box User-Agent
	resSingbox, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:     "tok-valid",
		UserAgent: "sing-box 1.14.0",
	})
	require.NoError(t, err)
	assert.Equal(t, "application/json; charset=utf-8", resSingbox.ContentType)
	assert.Equal(t, "config.json", resSingbox.Filename)
	assert.Contains(t, resSingbox.Content, "mixed-in")

	// 3. v2rayNG User-Agent (Base64)
	resB64, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:     "tok-valid",
		UserAgent: "v2rayNG/1.8.5",
	})
	require.NoError(t, err)
	assert.Equal(t, "text/plain; charset=utf-8", resB64.ContentType)
	assert.Equal(t, "sub.txt", resB64.Filename)
	assert.NotEmpty(t, resB64.Content)

	// 4. Raw format parameter override
	resRaw, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "raw",
	})
	require.NoError(t, err)
	assert.Contains(t, resRaw.Content, "trojan://")

	// 5. CDN and SNI Domain overrides
	resOverride, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "raw",
		CDN:    "custom-cdn.com",
		SNI:    "custom-sni.com",
	})
	require.NoError(t, err)
	assert.Contains(t, resOverride.Content, "custom-cdn.com")
}

func TestSubscriptionService_ErrorCases(t *testing.T) {
	now := time.Now()
	convSvc := converter.NewConverterService()
	ctx := context.Background()

	// 1. Expired user
	expiredUser := &model.User{
		ID:         2,
		Token:      "tok-expired",
		Expired:    now.Add(-1 * time.Hour),
		Quota:      500,
		ServerCode: "SG01",
		VPN:        "trojan",
	}
	svcExpired := NewSubscriptionService(&mockUserRepo{user: expiredUser}, &mockServerRepo{}, &mockProxyRepo{}, convSvc, "")
	_, err := svcExpired.GetSubscription(ctx, SubscriptionRequest{Token: "tok-expired"})
	assert.ErrorIs(t, err, model.ErrSubscriptionExpired)

	// 2. Quota exceeded
	depletedUser := &model.User{
		ID:         3,
		Token:      "tok-depleted",
		Expired:    now.Add(24 * time.Hour),
		Quota:      0,
		ServerCode: "SG01",
		VPN:        "trojan",
	}
	svcDepleted := NewSubscriptionService(&mockUserRepo{user: depletedUser}, &mockServerRepo{}, &mockProxyRepo{}, convSvc, "")
	_, err = svcDepleted.GetSubscription(ctx, SubscriptionRequest{Token: "tok-depleted"})
	assert.ErrorIs(t, err, model.ErrQuotaExceeded)

	// 3. Invalid token
	svcInvalid := NewSubscriptionService(&mockUserRepo{err: model.ErrUserNotFound}, &mockServerRepo{}, &mockProxyRepo{}, convSvc, "")
	_, err = svcInvalid.GetSubscription(ctx, SubscriptionRequest{Token: "not-found"})
	assert.ErrorIs(t, err, model.ErrUserNotFound)
}

func TestDetectFormat(t *testing.T) {
	assert.Equal(t, "clash", DetectFormat("Clash.Meta/v1.18.0"))
	assert.Equal(t, "clash", DetectFormat("mihomo/v1.18.0"))
	assert.Equal(t, "clash", DetectFormat("ClashVerge/1.0.0"))
	assert.Equal(t, "singbox", DetectFormat("sing-box/1.14.0"))
	assert.Equal(t, "sfa", DetectFormat("SFA/1.10.0"))
	assert.Equal(t, "bfr", DetectFormat("BFR/1.10.0"))
	assert.Equal(t, "base64", DetectFormat("v2rayNG/1.8.5"))
	assert.Equal(t, "base64", DetectFormat("Shadowrocket/1990"))
	assert.Equal(t, "base64", DetectFormat("Mozilla/5.0"))
}
