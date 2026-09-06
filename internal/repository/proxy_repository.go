package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/common/proxy"
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
		vpns := splitAndTrim(filter.VPN)
		if len(vpns) == 1 {
			conditions = append(conditions, "vpn = ?")
			args = append(args, strings.ToLower(vpns[0]))
		} else if len(vpns) > 1 {
			placeholders := make([]string, len(vpns))
			for i, v := range vpns {
				placeholders[i] = "?"
				args = append(args, strings.ToLower(v))
			}
			conditions = append(conditions, fmt.Sprintf("vpn IN (%s)", strings.Join(placeholders, ",")))
		}
	}
	if filter.CountryCode != "" {
		ccs := splitAndTrim(filter.CountryCode)
		if len(ccs) == 1 {
			conditions = append(conditions, "country_code = ?")
			args = append(args, strings.ToUpper(ccs[0]))
		} else if len(ccs) > 1 {
			placeholders := make([]string, len(ccs))
			for i, c := range ccs {
				placeholders[i] = "?"
				args = append(args, strings.ToUpper(c))
			}
			conditions = append(conditions, fmt.Sprintf("country_code IN (%s)", strings.Join(placeholders, ",")))
		}
	}
	if filter.Region != "" {
		regions := splitAndTrim(filter.Region)
		if len(regions) == 1 {
			conditions = append(conditions, "LOWER(region) = ?")
			args = append(args, strings.ToLower(regions[0]))
		} else if len(regions) > 1 {
			placeholders := make([]string, len(regions))
			for i, rg := range regions {
				placeholders[i] = "?"
				args = append(args, strings.ToLower(rg))
			}
			conditions = append(conditions, fmt.Sprintf("LOWER(region) IN (%s)", strings.Join(placeholders, ",")))
		}
	}
	if filter.Transport != "" {
		transports := splitAndTrim(filter.Transport)
		if len(transports) == 1 {
			conditions = append(conditions, "transport = ?")
			args = append(args, strings.ToLower(transports[0]))
		} else if len(transports) > 1 {
			placeholders := make([]string, len(transports))
			for i, t := range transports {
				placeholders[i] = "?"
				args = append(args, strings.ToLower(t))
			}
			conditions = append(conditions, fmt.Sprintf("transport IN (%s)", strings.Join(placeholders, ",")))
		}
	}
	if filter.ConnMode != "" {
		modes := splitAndTrim(filter.ConnMode)
		if len(modes) == 1 {
			conditions = append(conditions, "conn_mode = ?")
			args = append(args, strings.ToLower(modes[0]))
		} else if len(modes) > 1 {
			placeholders := make([]string, len(modes))
			for i, m := range modes {
				placeholders[i] = "?"
				args = append(args, strings.ToLower(m))
			}
			conditions = append(conditions, fmt.Sprintf("conn_mode IN (%s)", strings.Join(placeholders, ",")))
		}
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
		p.Raw = proxy.DecodeIfBase64(p.Raw)
		proxies = append(proxies, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating filtered proxies: %w", err)
	}

	return proxies, nil
}

func splitAndTrim(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
