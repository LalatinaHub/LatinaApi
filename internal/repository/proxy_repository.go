package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/common/repository"
)

type proxyRepo struct {
	repository.ProxyRepository
	db *sql.DB
}

// NewProxyRepository returns an extended ProxyRepository backed by *sql.DB.
func NewProxyRepository(db *sql.DB) ProxyRepository {
	return &proxyRepo{
		ProxyRepository: repository.NewProxyRepository(db),
		db:              db,
	}
}

func (r *proxyRepo) GetProxiesByFilter(ctx context.Context, filter ProxyFilter) ([]model.ProxyNode, error) {
	var (
		conditions []string
		args       []any
	)

	conditions = append(conditions, "1=1")

	if filter.VPN != "" {
		conditions = append(conditions, "vpn = ?")
		args = append(args, strings.ToLower(filter.VPN))
	}
	if filter.CountryCode != "" {
		conditions = append(conditions, "country_code = ?")
		args = append(args, strings.ToUpper(filter.CountryCode))
	}
	if filter.Region != "" {
		conditions = append(conditions, "region = ?")
		args = append(args, strings.ToUpper(filter.Region))
	}
	if filter.Transport != "" {
		conditions = append(conditions, "transport = ?")
		args = append(args, strings.ToLower(filter.Transport))
	}
	if filter.ConnMode != "" {
		conditions = append(conditions, "conn_mode = ?")
		args = append(args, strings.ToLower(filter.ConnMode))
	}
	if filter.TLS != nil {
		tlsVal := 0
		if *filter.TLS {
			tlsVal = 1
		}
		conditions = append(conditions, "tls = ?")
		args = append(args, tlsVal)
	}
	if filter.Include != "" {
		conditions = append(conditions, "remark LIKE ?")
		args = append(args, "%"+filter.Include+"%")
	}
	if filter.Exclude != "" {
		conditions = append(conditions, "remark NOT LIKE ?")
		args = append(args, "%"+filter.Exclude+"%")
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	query := fmt.Sprintf(
		"SELECT id, server, ip, server_port, uuid, password, security, alter_id, method, plugin, plugin_opts, host, tls, transport, path, service_name, insecure, sni, remark, conn_mode, country_code, region, org, vpn, raw FROM proxies WHERE %s ORDER BY RANDOM() LIMIT ?;",
		strings.Join(conditions, " AND "),
	)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query proxies with filter: %w", err)
	}
	defer rows.Close()

	var proxies []model.ProxyNode
	for rows.Next() {
		var p model.ProxyNode
		err := rows.Scan(
			&p.ID,
			&p.Server,
			&p.IP,
			&p.ServerPort,
			&p.UUID,
			&p.Password,
			&p.Security,
			&p.AlterID,
			&p.Method,
			&p.Plugin,
			&p.PluginOpts,
			&p.Host,
			&p.TLS,
			&p.Transport,
			&p.Path,
			&p.ServiceName,
			&p.Insecure,
			&p.SNI,
			&p.Remark,
			&p.ConnMode,
			&p.CountryCode,
			&p.Region,
			&p.Org,
			&p.VPN,
			&p.Raw,
		)
		if err != nil {
			continue
		}
		proxies = append(proxies, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating filtered proxies: %w", err)
	}

	return proxies, nil
}
