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
		Password:   "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d",
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
	b64FreeVless := "dmxlc3M6Ly8zMGE4OGM0MC01ODFlLTQ5NTItYWE0Yy04YWY5NTY4NmVkZTBAd3d3Lmdvdi51YTo4ODgwP2VuY3J5cHRpb249bm9uZSZzZWN1cml0eT1ub25lJnR5cGU9d3MmaG9zdD1yYXBpZC1sYWItOTVlZi4xNzMtNzRjLndvcmtlcnMuZGV2JnBhdGg9L3B5aXA9cHJveHlpcC5rci5jbWxpdXNzc3MubmV0I0BEZWx0YUtyb25lY2tlckdpdGh1Yg=="
	proxyRepo := &mockProxyRepo{
		proxies: []model.ProxyNode{
			{
				VPN:      "vless",
				ConnMode: "cdn",
				Raw:      b64FreeVless,
			},
		},
	}
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

	// 3. v2rayNG User-Agent (Raw links, not base64)
	resV2Ray, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:     "tok-valid",
		UserAgent: "v2rayNG/1.8.5",
	})
	require.NoError(t, err)
	assert.Equal(t, "text/plain; charset=utf-8", resV2Ray.ContentType)
	assert.Equal(t, "sub.txt", resV2Ray.Filename)
	assert.NotEmpty(t, resV2Ray.Content)
	assert.Contains(t, resV2Ray.Content, "trojan://")

	// 4. Raw format parameter override
	resRaw, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "raw",
	})
	require.NoError(t, err)
	assert.Contains(t, resRaw.Content, "trojan://")
	assert.Contains(t, resRaw.Content, "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d")
	assert.NotContains(t, resRaw.Content, "tok-valid@")

	// 5. CDN and SNI Domain overrides
	resOverride, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "raw",
		CDN:    "custom-cdn.com",
		SNI:    "custom-sni.com",
	})
	require.NoError(t, err)
	assert.Contains(t, resOverride.Content, "custom-cdn.com")

	// 6. Comma-separated VPN and Mode filters
	resMulti, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "raw",
		VPN:    "vmess,vless,trojan",
		Mode:   "cdn,sni",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resMulti.Content)
	assert.Contains(t, resMulti.Content, "trojan://")

	// 7. format=base64 and default format return raw unencoded URIs
	resExplicitB64, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "base64",
	})
	require.NoError(t, err)
	assert.Contains(t, resExplicitB64.Content, "trojan://")
	assert.NotContains(t, resExplicitB64.Content, "ey")

	// 8. Empty format with generic browser User-Agent returns raw
	resBrowser, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:     "tok-valid",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	})
	require.NoError(t, err)
	assert.Contains(t, resBrowser.Content, "trojan://")

	// 9. Free node from DB with base64 encoded raw URL is decoded into plaintext URI
	resWithFree, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "raw",
	})
	require.NoError(t, err)
	assert.Contains(t, resWithFree.Content, "vless://30a88c40-581e-4952-aa4c-8af95686ede0@www.gov.ua:8880")
	assert.NotContains(t, resWithFree.Content, "dmxlc3M6")

	// 10. mode=cdn with cdn parameter: server and sni set to cdn, host untouched
	resCDNOnly, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "raw",
		Mode:   "cdn",
		CDN:    "104.18.2.2",
	})
	require.NoError(t, err)
	assert.Contains(t, resCDNOnly.Content, "@104.18.2.2:443")
	assert.Contains(t, resCDNOnly.Content, "sni=104.18.2.2")
	assert.Contains(t, resCDNOnly.Content, "host=sg01.foolvpn.me")
	assert.NotContains(t, resCDNOnly.Content, "SNI TCP TLS")

	// 11. mode=sni with sni parameter: sni set to sni, server and host untouched
	resSNIOnly, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "raw",
		Mode:   "sni",
		SNI:    "google.com",
	})
	require.NoError(t, err)
	assert.Contains(t, resSNIOnly.Content, "@sg01.foolvpn.me:443")
	assert.Contains(t, resSNIOnly.Content, "sni=google.com")
	assert.Contains(t, resSNIOnly.Content, "host=sg01.foolvpn.me")
	assert.NotContains(t, resSNIOnly.Content, "@google.com")
	assert.NotContains(t, resSNIOnly.Content, "CDN WS")

	// 12. mode=cdn,sni with both cdn and sni parameters
	resBoth, err := svc.GetSubscription(ctx, SubscriptionRequest{
		Token:  "tok-valid",
		Format: "raw",
		Mode:   "cdn,sni",
		CDN:    "104.18.2.2",
		SNI:    "google.com",
	})
	require.NoError(t, err)
	// CDN node has server 104.18.2.2 and sni 104.18.2.2
	assert.Contains(t, resBoth.Content, "@104.18.2.2:443")
	assert.Contains(t, resBoth.Content, "sni=104.18.2.2")
	// SNI node has server sg01.foolvpn.me and sni google.com (not server google.com)
	assert.Contains(t, resBoth.Content, "@sg01.foolvpn.me:443")
	assert.Contains(t, resBoth.Content, "sni=google.com")
	assert.NotContains(t, resBoth.Content, "@google.com:443")
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

	// 4. Nil user repository safety
	svcNil := NewSubscriptionService(nil, nil, nil, convSvc, "")
	_, err = svcNil.GetSubscription(ctx, SubscriptionRequest{Token: "tok-test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection unavailable")
}

func TestDetectFormat(t *testing.T) {
	assert.Equal(t, "clash", DetectFormat("Clash.Meta/v1.18.0"))
	assert.Equal(t, "clash", DetectFormat("mihomo/v1.18.0"))
	assert.Equal(t, "clash", DetectFormat("ClashVerge/1.0.0"))
	assert.Equal(t, "singbox", DetectFormat("sing-box/1.14.0"))
	assert.Equal(t, "sfa", DetectFormat("SFA/1.10.0"))
	assert.Equal(t, "bfr", DetectFormat("BFR/1.10.0"))
	assert.Equal(t, "raw", DetectFormat("v2rayNG/1.8.5"))
	assert.Equal(t, "raw", DetectFormat("Shadowrocket/1990"))
	assert.Equal(t, "raw", DetectFormat("Mozilla/5.0"))
}
