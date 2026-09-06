package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyRepository_GetProxiesByFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewProxyRepository(db)
	ctx := context.Background()

	cols := []string{
		"id", "server", "ip", "server_port", "uuid", "password", "security", "alter_id",
		"method", "plugin", "plugin_opts", "host", "tls", "transport", "path", "service_name",
		"insecure", "sni", "remark", "conn_mode", "country_code", "region", "org", "vpn", "raw",
	}

	rows := sqlmock.NewRows(cols).
		AddRow(
			1, "proxy.example.com", "1.1.1.1", 443, "uuid-1", "pass", "auto", 0,
			"", "", "", "example.com", true, "ws", "/ws", "",
			true, "example.com", "SG-Node-1", "cdn", "SG", "SIN", "Cloudflare", "vmess", "vmess://raw",
		)

	tlsTrue := true
	mock.ExpectQuery("SELECT (.+) FROM proxies WHERE (.+) ORDER BY RANDOM\\(\\) LIMIT \\?").
		WithArgs("vmess", "SG", "SIN", "ws", "cdn", 1, "%Node%", "%Bad%", 5).
		WillReturnRows(rows)

	filter := ProxyFilter{
		VPN:         "vmess",
		CountryCode: "sg",
		Region:      "sin",
		Transport:   "WS",
		ConnMode:    "CDN",
		TLS:         &tlsTrue,
		Include:     "Node",
		Exclude:     "Bad",
		Limit:       5,
	}

	proxies, err := repo.GetProxiesByFilter(ctx, filter)
	require.NoError(t, err)
	require.Len(t, proxies, 1)
	assert.Equal(t, "SG-Node-1", proxies[0].Remark)
	assert.Equal(t, "vmess", proxies[0].VPN)
	assert.Equal(t, "SG", proxies[0].CountryCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}
