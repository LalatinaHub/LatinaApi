# LatinaApi

High-performance Universal Subscription & Proxy Engine for the LatinaHub ecosystem, built with Go and Gin adhering to Clean Architecture principles.

## Features

- **Universal Client Negotiation**: Automatic User-Agent detection (Clash Meta/Mihomo, sing-box v1.14+, SFA, BFR, v2rayNG, Shadowrocket, NekoBox, curl/browser).
- **Central Subconverter Engine**: Direct integration with `github.com/LalatinaHub/common/subconverter` for zero-allocation YAML/JSON formatting.
- **Dynamic Node Matrix**: Combines edge server premium nodes (CDN & SNI variations) with curated free proxies.
- **Geolocation Enrichment**: Instant O(1) airport IATA code lookup and flag enrichment via `common/region`.
- **Database & Pooling**: Built-in LibSQL Turso connection pool and auto-indexing via `common/database`.
- **Production-Ready**: Zero-allocation logging with `zerolog` (`X-Request-ID`/`X-Trace-ID`), sliding window rate limiting, CORS, error middleware, and pprof profiling.

## Getting Started

### Prerequisites

- Go 1.24+ (tested on Go 1.25+ and Go 1.27+)
- Turso / LibSQL Database

### Installation & Run

```bash
# Clone the repository
git clone https://github.com/LalatinaHub/LatinaApi.git
cd LatinaApi

# Copy environment template
cp .env.example .env

# Run locally
go run ./cmd/api
```

### Environment Variables

| Variable                     | Default                  | Description                                      |
| ---------------------------- | ------------------------ | ------------------------------------------------ |
| `PORT`                       | `8080`                   | HTTP port to listen on                           |
| `APP_ENV`                    | `development`            | Environment mode (`development` or `production`) |
| `TURSO_DATABASE_URL`         | `file:local.db`          | Turso / LibSQL connection URL                    |
| `TURSO_AUTH_TOKEN`           | -                        | Turso database authentication token              |
| `DEFAULT_SUBSCRIPTION_TITLE` | `LatinaHub Subscription` | Title for generated subscription files           |
| `CACHE_TTL_MINUTES`          | `5`                      | In-memory cache TTL for edge server info         |
| `RATE_LIMIT_RPS`             | `20`                     | Rate limiter requests per second per IP          |

## Endpoints

- `GET /sub?token={token}` - Universal subscription endpoint
- `GET /api/v1/sub?token={token}` - Subscription API alias
- `GET /user/:apiToken/:id` - User management and auto-provisioning
- `POST /db/:apiToken/exec` - Protected atomic SQL transaction execution
- `POST /api/v1/convert` - Raw proxy link converter
- `GET /health` & `GET /api/v1/ping` - Health check endpoints
- `GET /api/v1/info` - Client IP and GeoIP information

## Testing

```bash
go test -v ./...
```

## License

MIT
