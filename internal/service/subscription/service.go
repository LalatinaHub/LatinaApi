package subscription

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/LatinaApi/internal/repository"
	"github.com/LalatinaHub/LatinaApi/internal/service/converter"
	"github.com/LalatinaHub/LatinaApi/pkg/httputil"
	"github.com/LalatinaHub/common/region"
)

// SubscriptionRequest contains all query and header parameters for a subscription request.
type SubscriptionRequest struct {
	Token       string
	VPN         string
	Format      string
	UserAgent   string
	Region      string
	CountryCode string
	Mode        string
	Transport   string
	TLS         *bool
	Free        bool
	Premium     bool
	Limit       int
	CDN         string
	SNI         string
	Include     string
	Exclude     string
}

// SubscriptionResult holds the formatted output and HTTP response header metadata.
type SubscriptionResult struct {
	Content               string
	ContentType           string
	Filename              string
	UserInfo              string
	ProfileTitle          string
	ProfileUpdateInterval int
}

// SubscriptionService coordinates token verification, node generation, and multi-format conversion.
type SubscriptionService interface {
	GetSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResult, error)
}

type subscriptionService struct {
	userRepo         repository.UserRepository
	serverRepo       repository.ServerRepository
	proxyRepo        repository.ProxyRepository
	converterService converter.ConverterService
	httpClient       *httputil.Client
	defaultTitle     string
	infoCache        sync.Map // code -> serverInfoCache
}

type serverInfoCache struct {
	data      map[string]any
	expiresAt time.Time
}

// NewSubscriptionService returns a new SubscriptionService.
func NewSubscriptionService(
	userRepo repository.UserRepository,
	serverRepo repository.ServerRepository,
	proxyRepo repository.ProxyRepository,
	convSvc converter.ConverterService,
	defaultTitle string,
) SubscriptionService {
	if defaultTitle == "" {
		defaultTitle = "LatinaHub"
	}
	return &subscriptionService{
		userRepo:         userRepo,
		serverRepo:       serverRepo,
		proxyRepo:        proxyRepo,
		converterService: convSvc,
		httpClient:       httputil.NewClient(3 * time.Second),
		defaultTitle:     defaultTitle,
	}
}

func (s *subscriptionService) GetSubscription(ctx context.Context, req SubscriptionRequest) (*SubscriptionResult, error) {
	if req.Token == "" {
		return nil, model.ErrInvalidToken
	}

	// 1. Authenticate user
	user, err := s.userRepo.GetUserByToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if now.After(user.Expired) {
		return nil, model.ErrSubscriptionExpired
	}
	if user.Quota <= 0 {
		return nil, model.ErrQuotaExceeded
	}
	if !user.IsActive(now) {
		return nil, model.ErrUnauthorized
	}

	// 2. Generate nodes
	var nodes []model.ProxyNode

	// Premium edge nodes
	if !req.Free && user.ServerCode != "" {
		premiumNodes, err := s.buildPremiumNodes(ctx, user, req)
		if err == nil {
			nodes = append(nodes, premiumNodes...)
		}
	}

	// Free database proxy nodes
	if !req.Premium && s.proxyRepo != nil {
		filter := repository.ProxyFilter{
			VPN:         req.VPN,
			CountryCode: req.CountryCode,
			Region:      req.Region,
			Transport:   req.Transport,
			ConnMode:    req.Mode,
			TLS:         req.TLS,
			Include:     req.Include,
			Exclude:     req.Exclude,
			Limit:       req.Limit,
		}
		if filter.Limit <= 0 {
			filter.Limit = 10
		}

		freeNodes, err := s.proxyRepo.GetProxiesByFilter(ctx, filter)
		if err == nil {
			nodes = append(nodes, freeNodes...)
		}
	}

	// Dynamic domain overrides
	for i := range nodes {
		applyDomainOverrides(&nodes[i], req.CDN, req.SNI)
	}

	// 3. Format determination
	format := strings.ToLower(req.Format)
	if format == "" || format == "auto" {
		format = DetectFormat(req.UserAgent)
	}

	// 4. Convert nodes
	var (
		content     string
		contentType string
		filename    string
	)

	switch format {
	case "clash", "clash.meta", "mihomo":
		content, err = s.converterService.ToClash(nodes, "")
		contentType = "application/yaml; charset=utf-8"
		filename = "config.yaml"

	case "singbox", "sing-box":
		content, err = s.converterService.ToSingbox(nodes, "standard", "")
		contentType = "application/json; charset=utf-8"
		filename = "config.json"

	case "sfa":
		content, err = s.converterService.ToSFA(nodes, "")
		contentType = "application/json; charset=utf-8"
		filename = "sfa.json"

	case "bfr":
		content, err = s.converterService.ToBFR(nodes, "")
		contentType = "application/json; charset=utf-8"
		filename = "bfr.json"

	case "raw":
		content = s.converterService.ToRawString(nodes)
		contentType = "text/plain; charset=utf-8"
		filename = "sub.txt"

	case "base64", "b64":
		fallthrough
	default:
		content = s.converterService.ToBase64(nodes)
		contentType = "text/plain; charset=utf-8"
		filename = "sub.txt"
	}

	if err != nil {
		return nil, fmt.Errorf("conversion failed: %w", err)
	}

	// 5. Build standard subscription headers
	totalBytes := user.Quota * 1024 * 1024
	userInfo := fmt.Sprintf("upload=0; download=0; total=%d; expire=%d", totalBytes, user.Expired.Unix())
	profileTitle := fmt.Sprintf("%s - %s", s.defaultTitle, strings.ToUpper(user.VPN))

	return &SubscriptionResult{
		Content:               content,
		ContentType:           contentType,
		Filename:              filename,
		UserInfo:              userInfo,
		ProfileTitle:          profileTitle,
		ProfileUpdateInterval: 24,
	}, nil
}

func (s *subscriptionService) buildPremiumNodes(ctx context.Context, user *model.User, req SubscriptionRequest) ([]model.ProxyNode, error) {
	server, err := s.serverRepo.GetServerByCode(ctx, user.ServerCode)
	if err != nil {
		return nil, err
	}

	// Try lookup IATA city
	cityName, _ := region.Lookup(server.Country)
	if cityName == "" {
		cityName = server.Country
	}

	domain := server.Domain
	if domain == "" {
		domain = server.IP
	}

	var premiumNodes []model.ProxyNode
	vpnProto := strings.ToLower(user.VPN)
	if req.VPN != "" && !strings.EqualFold(req.VPN, vpnProto) {
		return nil, nil
	}

	// CDN WS Variant
	if req.Mode == "" || strings.EqualFold(req.Mode, "cdn") {
		if req.Transport == "" || strings.EqualFold(req.Transport, "ws") {
			// TLS 443
			if req.TLS == nil || *req.TLS {
				node := model.ProxyNode{
					VPN:         vpnProto,
					Server:      domain,
					ServerPort:  443,
					UUID:        user.Token,
					Password:    user.Password,
					TLS:         true,
					Transport:   "ws",
					Path:        fmt.Sprintf("/%s-ws", vpnProto),
					Host:        domain,
					SNI:         domain,
					Remark:      fmt.Sprintf("%s CDN WS TLS", cityName),
					ConnMode:    "cdn",
					CountryCode: server.Country,
				}
				premiumNodes = append(premiumNodes, node)
			}
			// NTLS 80
			if req.TLS == nil || !*req.TLS {
				node := model.ProxyNode{
					VPN:         vpnProto,
					Server:      domain,
					ServerPort:  80,
					UUID:        user.Token,
					Password:    user.Password,
					TLS:         false,
					Transport:   "ws",
					Path:        fmt.Sprintf("/%s-ws", vpnProto),
					Host:        domain,
					Remark:      fmt.Sprintf("%s CDN WS NTLS", cityName),
					ConnMode:    "cdn",
					CountryCode: server.Country,
				}
				premiumNodes = append(premiumNodes, node)
			}
		}

		// CDN gRPC Variant (TLS 443 only)
		if (req.Transport == "" || strings.EqualFold(req.Transport, "grpc")) && (req.TLS == nil || *req.TLS) {
			node := model.ProxyNode{
				VPN:         vpnProto,
				Server:      domain,
				ServerPort:  443,
				UUID:        user.Token,
				Password:    user.Password,
				TLS:         true,
				Transport:   "grpc",
				ServiceName: fmt.Sprintf("%s-grpc", vpnProto),
				Host:        domain,
				SNI:         domain,
				Remark:      fmt.Sprintf("%s CDN gRPC TLS", cityName),
				ConnMode:    "cdn",
				CountryCode: server.Country,
			}
			premiumNodes = append(premiumNodes, node)
		}
	}

	// SNI Variant (TLS 443 TCP/gRPC only)
	if (req.Mode == "" || strings.EqualFold(req.Mode, "sni")) && (req.TLS == nil || *req.TLS) {
		node := model.ProxyNode{
			VPN:         vpnProto,
			Server:      domain,
			ServerPort:  443,
			UUID:        user.Token,
			Password:    user.Password,
			TLS:         true,
			Transport:   "tcp",
			Host:        domain,
			SNI:         domain,
			Remark:      fmt.Sprintf("%s SNI TCP TLS", cityName),
			ConnMode:    "sni",
			CountryCode: server.Country,
		}
		premiumNodes = append(premiumNodes, node)
	}

	return premiumNodes, nil
}

func applyDomainOverrides(node *model.ProxyNode, cdnOverride, sniOverride string) {
	if cdnOverride != "" && strings.EqualFold(node.ConnMode, "cdn") {
		node.Server = cdnOverride
		node.Host = cdnOverride
	}
	if sniOverride != "" && strings.EqualFold(node.ConnMode, "sni") {
		node.Server = sniOverride
		node.SNI = sniOverride
	}
}

// DetectFormat evaluates the User-Agent header and returns the target format.
func DetectFormat(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "clash") || strings.Contains(ua, "mihomo") || strings.Contains(ua, "flclash") {
		return "clash"
	}
	if strings.Contains(ua, "sfa") {
		return "sfa"
	}
	if strings.Contains(ua, "bfr") {
		return "bfr"
	}
	if strings.Contains(ua, "sing-box") || strings.Contains(ua, "sfi") || strings.Contains(ua, "sfm") {
		return "singbox"
	}
	if strings.Contains(ua, "v2rayng") || strings.Contains(ua, "shadowrocket") ||
		strings.Contains(ua, "nekobox") || strings.Contains(ua, "nekoray") ||
		strings.Contains(ua, "streisand") {
		return "base64"
	}

	// Default fallback to Base64
	return "base64"
}
