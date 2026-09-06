# LatinaApi — Engineering Roadmap & Implementation Plan (TODO)

Dokumen ini memuat rencana kerja sistematis untuk mengimplementasikan dan memodernisasi layanan API `github.com/FoolVPN-ID/megalodon-api` ke dalam repositori `github.com/LalatinaHub/LatinaApi`. Proyek ini mengintegrasikan spesifikasi **Fase 1 (Universal Smart Subscription Engine)** dari `LatinaServer/TODO.md` serta memanfaatkan pustaka domain sentral `github.com/LalatinaHub/common` sebagai fondasi utama (*Leaf Module & Single Source of Truth*).

---

## 🧭 Arsitektur Sistem & Integrasi Modul

```
                     ┌─────────────────────────────────────────────────────────┐
                     │                 Client / User Application               │
                     │  (Clash.Meta, sing-box, v2rayNG, Shadowrocket, Browser) │
                     └────────────────────────────┬────────────────────────────┘
                                                  │
                                       GET /sub?token={TOKEN}
                                                  ▼
                     ┌─────────────────────────────────────────────────────────┐
                     │              LatinaApi Gateway (Gin Engine)             │
                     │  - Request Logger (zerolog with X-Request-ID / Trace-ID)│
                     │  - Rate Limiter, Error & Gzip Middlewares               │
                     │  - User-Agent Detection & Header Parser                 │
                     └────────────────────────────┬────────────────────────────┘
                                                  │
                   ┌──────────────────────────────┴──────────────────────────────┐
                   ▼                                                             ▼
    ┌─────────────────────────────┐                               ┌─────────────────────────────┐
    │  Subscription Orchestrator  │                               │    User & DB Admin Service  │
    │  - Token Authentication     │                               │  - GET /user/:apiToken/:id  │
    │  - Quota & Expiry Check     │                               │  - POST /db/:apiToken/exec  │
    │    (via common model/repo)  │                               │  - POST /api/v1/convert     │
    └──────────────┬──────────────┘                               └──────────────┬──────────────┘
                   │                                                             │
         ┌─────────┴─────────┐                                                   │
         ▼                   ▼                                                   │
┌─────────────────┐ ┌─────────────────┐                                          │
│ Premium Builder │ │  Free DB Proxy  │                                          │
│ Node Generator  │ │  Filter Query   │                                          │
│ (Edge /info +   │ │  (proxies table │                                          │
│  common/region) │ │   model.Proxy)  │                                          │
└────────┬────────┘ └────────┬────────┘                                          │
         └─────────┬─────────┘                                                   │
                   ▼                                                             │
    ┌──────────────────────────────────────────────────────────┐                 │
    │        Central Subconverter Engine (LalatinaHub/common)  │                 │
    │  - conv.ToClash(template)       -> Clash Meta YAML       │                 │
    │  - conv.ToSingbox(prof, tmpl)   -> sing-box v1.14+ JSON  │                 │
    │  - conv.ToBase64()              -> Base64 Encoded Lines  │                 │
    │  - conv.ToRawString()           -> Raw Proxy URLs Lines  │                 │
    │  - conv.ToSFA() / ToBFR()       -> Mobile Profile JSON   │                 │
    └──────────────────────────────┬───────────────────────────┘                 │
                                   │                                             │
                                   └──────────────────────┬──────────────────────┘
                                                          │
                                                          ▼
                     ┌─────────────────────────────────────────────────────────┐
                     │       LalatinaHub/common Data Layer & Turso LibSQL      │
                     │  - database: Singleton Pool (GetDB, InitIndexes, Close) │
                     │  - model: User (IsActive), Server, ProxyNode, KeyValue  │
                     │  - repository: Base UserRepository, ServerRepo, etc.    │
                     │  - proxy: Bi-directional Parser & Formatter (VMess/etc) │
                     │  - region: High-speed O(1) IATA Geolocation Lookup      │
                     └─────────────────────────────────────────────────────────┘
```

---

## 📋 Rencana Kerja Bertahap (Phased Implementation Plan)

### 🏗️ Fase 1: Inisialisasi Modul & Pondasi Arsitektur Bersih (Clean Architecture)
> **Target**: Menyiapkan modul Go 1.24+ (kompatibel Go 1.27+), konfigurasi dependency terkurasi sesuai ekosistem `LatinaHub`, struktur direktori Clean Architecture yang selaras dengan `LatinaServer`, dan utilitas inti.

- [x] **1.1. Inisialisasi Go Module & Audit Dependency**:
  - Inisialisasi modul `github.com/LalatinaHub/LatinaApi` (target `go 1.24.5` / `go 1.25.5` selaras dengan `common` dan `LatinaServer`).
  - Pasang dependensi langsung (*direct dependencies*):
    - `github.com/LalatinaHub/common v0.1.0` (Leaf Module: model domain, database pool, repository base, proxy parser, dan subconverter)
    - `github.com/gin-gonic/gin v1.10.0` (HTTP web framework)
    - `github.com/gin-contrib/cors v1.7.3` (CORS middleware)
    - `github.com/rs/zerolog v1.35.1` (Zero-allocation structured logger)
    - `github.com/google/uuid v1.6.0` (UUID generator untuk token pengguna & request tracing)
    - `github.com/joho/godotenv v1.5.1` (Loader variabel lingkungan lokal `.env`)
    - `github.com/stretchr/testify v1.12.1` (Framework unit & mock testing)
  - Dependensi tak-langsung (*indirect / abstracted by common*):
    - `github.com/tursodatabase/libsql-client-go` (Telah dibungkus aman oleh `common/database`)
    - `gopkg.in/yaml.v3` (Telah dibungkus oleh `common/subconverter`)
- [x] **1.2. Struktur Direktori Standar (Konsisten dengan Ekosistem LatinaServer)**:
  ```
  LatinaApi/
  ├── cmd/
  │   └── api/
  │       └── main.go                 # Entrypoint server, DI wire-up & graceful shutdown
  ├── config/
  │   └── config.go                # Validasi env (PORT, TURSO_URL, TURSO_TOKEN, CACHE_TTL, dll.)
  ├── internal/
  │   ├── domain/
  │   │   └── model/                  # Type-alias re-export dari common/model (User, Server, ProxyNode, KV)
  │   ├── repository/                 # Data access layer (mengomposisi common/repository + query khusus)
  │   │   ├── interfaces.go           # Re-export & perluasan interface repository
  │   │   ├── user_repository.go      # Query GetUserByToken, GetUserByID, CreateUserWithDefaults
  │   │   ├── server_repository.go    # Query GetServerByCode & filter
  │   │   ├── proxy_repository.go     # Query GetProxiesByFilter (dinamis: vpn, cc, tls, transport)
  │   │   └── kv_repository.go        # Wrapper GetValueByKey berbasis common/repository.KVRepository
  │   ├── service/
  │   │   ├── subscription/           # Core orchestrator langganan & builder node edge/free
  │   │   ├── converter/              # Adapter service yang memanggil common/subconverter
  │   │   ├── user/                   # Manajemen akun user & provisioning otomatis
  │   │   └── dbadmin/                # Eksekusi SQL transaction via API token
  │   └── api/
  │       ├── handler/                # HTTP handlers Gin (SubHandler, UserHandler, AdminHandler, Ping)
  │       ├── middleware/             # RequestLogger, Recovery, RateLimiter, CORS, ErrorMiddleware
  │       └── router.go               # Definisi routing Gin & grouping rute
  ├── pkg/
  │   ├── logger/                     # Structured zerolog wrapper selaras dengan LatinaServer/pkg/logger
  │   └── httputil/                   # Client HTTP dengan pooling, context timeout & exponential backoff
  ├── .air.toml                       # Live reload konfigurasi lokal
  ├── Dockerfile                      # Multi-stage distroless build
  ├── go.mod
  ├── go.sum
  └── TODO.md
  ```
- [x] **1.3. Konfigurasi Lingkungan (`config`)**:
  - Validasi variabel lingkungan: `PORT` (default 8080), `APP_ENV` (development/production), `TURSO_DATABASE_URL`, `TURSO_AUTH_TOKEN`, `DEFAULT_SUBSCRIPTION_TITLE`, `CACHE_TTL_MINUTES`, `RATE_LIMIT_RPS`.
- [x] **1.4. Logging & Graceful Shutdown**:
  - Logger terstruktur `zerolog` selaras dengan `LatinaServer/pkg/logger`: integrasi `X-Request-ID` dan `X-Trace-ID`, mode human-readable console writer saat dev, dan JSON saat production.
  - Shutdown bersih menggunakan `os.Signal` (`SIGINT`, `SIGTERM`) dengan `context.WithTimeout(ctx, 5*time.Second)` dan pemanggilan `database.Close()`.

---

### 💾 Fase 2: Integrasi Lapisan Basis Data & `LalatinaHub/common`
> **Target**: Mengadopsi connection pool LibSQL dari `common`, re-export domain model, dan memperluas repository query khusus API subscription.

- [x] **2.1. Integrasi Pool Koneksi Turso dari `common/database`**:
  - Gunakan `database.GetDB()` dari `github.com/LalatinaHub/common/database` (otomatis mengonfigurasi pool: `MaxOpenConns: 25`, `MaxIdleConns: 5`, lifetime 5m, idle 1m).
  - Panggil `database.InitIndexes(ctx)` saat server startup untuk memastikan indeks tabel `users` (`expired`, `quota`, `server_code`, `vpn`) terpasang.
  - Pastikan penutupan pool koneksi via `database.Close()` saat shutdown server.
- [x] **2.2. Re-Export Model Domain (`internal/domain/model`)**:
  - Mengikuti konvensi `LatinaServer`, lakukan re-export type-alias dari `common/model` untuk menjaga portabilitas internal tanpa merusak kompatibilitas:
    - `type User = model.User` (memiliki method `IsActive(now time.Time) bool`)
    - `type Server = model.Server` (memiliki method `IsFull() bool`)
    - `type ProxyNode = model.ProxyNode`
    - `type KeyValue = model.KeyValue`
- [x] **2.3. Extended Repository Queries (`internal/repository`)**:
  - Memperluas interface repository dasar dari `common/repository` untuk kebutuhan unik subscription engine:
    - **`UserRepository`**:
      - Menyematkan (*embed*) `common/repository.UserRepository` (mewarisi `GetActiveUsersGroupedByVPN`, `DeductQuota`, `DeductQuotaBatch`, `CreateUser`).
      - Menambahkan query spesifik:
        - `GetUserByToken(ctx context.Context, token string) (*model.User, error)`: `SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users WHERE token = ? LIMIT 1;`
        - `GetUserByID(ctx context.Context, id int64) (*model.User, error)`: Pencarian user berdasarkan ID.
        - `CreateUserWithDefaults(ctx context.Context, id int64) (*model.User, error)`: Auto-provisioning user dengan default password random, token UUID, kuota 1000MB, expired hari ini.
    - **`ServerRepository`**:
      - Menyematkan `common/repository.ServerRepository` (mewarisi `GetAll`).
      - Menambahkan query spesifik:
        - `GetServerByCode(ctx context.Context, code string) (*model.Server, error)`: Mengambil server berdasarkan kode unik (`SELECT ... WHERE code = ? LIMIT 1`).
    - **`ProxyRepository`**:
      - Menyematkan `common/repository.ProxyRepository` (mewarisi `GetRelays`).
      - Menambahkan query spesifik:
        - `GetProxiesByFilter(ctx context.Context, filter ProxyFilter) ([]model.ProxyNode, error)`: Query dinamis terindeks ke tabel `proxies` dengan filter opsional (`vpn`, `country_code`, `region`, `tls`, `transport`, `conn_mode`, `limit`, random order).
    - **`KVRepository`**:
      - Menyematkan `common/repository.KVRepository` (mewarisi `GetAll`).
      - Menambahkan query spesifik:
        - `GetValueByKey(ctx context.Context, key string) (string, error)`: Mengambil nilai konfigurasi berdasarkan key (misal `apiToken`).
- [x] **2.4. Unit Testing & Mocking Repository**:
  - Menguji fungsionalitas repository menggunakan mock interfaces dari `github.com/LalatinaHub/common/repository/mocks` dan `sqlmock`.

---

### 🚀 Fase 3: Smart Subscription Engine & Node Generation (`/sub`)
> **Target**: Mengorkestrasi pembuatan proxy node premium & free, validasi akun pengguna, dan resolusi domain dinamis.

- [x] **3.1. Parsing Parameter Query `/sub` & `/api/v1/sub`**:
  - Autentikasi: `token` (dan backward-compatible alias `pass`).
  - Parameter filter:
    - `vpn`: protokol VPN (`trojan`, `vless`, `vmess`, `shadowsocks` / `ss`).
    - `format`: target format output (`auto`, `clash`, `singbox` / `sing-box`, `sfa`, `bfr`, `raw`, `base64`).
    - `region`, `cc`: filter berdasarkan region IATA / kode negara 2 huruf.
    - `mode`: mode koneksi (`cdn`, `sni`).
    - `transport`: jenis transport (`ws`, `grpc`, `tcp`).
    - `tls`: status TLS (`0` = non-TLS port 80, `1` = TLS port 443).
    - `free`: filter node free (`1` = hanya free proxy, `0` = gabungan).
    - `premium`: filter node premium (`1` = hanya premium edge, `0` = gabungan).
    - `limit`: batas maksimum node yang dikembalikan (default 10).
    - `cdn`, `sni`: custom domain/host overrides.
    - `include`, `exclude`: regex / string matching pada remark node.
- [x] **3.2. Validasi Akun Menggunakan Model Domain `common`**:
  - Ambil data akun via `userRepo.GetUserByToken(ctx, token)`.
  - Validasi status akun menggunakan method bawaan domain `user.IsActive(time.Now())`:
    - Jika akun tidak ditemukan: HTTP 403 Forbidden / HTTP 404 Not Found.
    - Jika `now.After(user.Expired)`: HTTP 403 Forbidden dengan payload `"Subscription expired"`.
    - Jika `user.Quota <= 0`: HTTP 403 Forbidden dengan payload `"Subscription quota exceeded"`.
- [x] **3.3. Premium Node Builder (Edge Node Generation & Geolocation Enrichment)**:
  - Dapatkan konfigurasi server dari `ServerRepository.GetServerByCode(ctx, user.ServerCode)`.
  - Fetch info edge server dinamis (`/api/v1/info`) dengan caching in-memory (5 menit) dan HTTP timeout 3 detik.
  - Manfaatkan `github.com/LalatinaHub/common/region` (`region.Lookup(iataCode)`) untuk memperkaya remark node dengan nama kota dan emoji bendera negara secara instan (O(1) lookup table).
  - Bangun matriks kombinasi node sesuai spesifikasi `megalodon-api`:
    - CDN mode: WS/gRPC port 80/443 (lewati TCP dan gRPC port 80).
    - SNI mode: TLS port 443 only.
  - Bentuk objek `model.ProxyNode` untuk masing-masing varian.
- [x] **3.4. Penggabungan Node Free & Dynamic Domain Overrides**:
  - Jika `premium != 1`, ambil node dari `ProxyRepository.GetProxiesByFilter(...)`.
  - Lakukan override field `Server`, `Host`, dan `SNI` jika parameter query `cdn` atau `sni` disediakan oleh client.

---

### 📱 Fase 4: Integrasi Pustaka Sentral `common/subconverter` & Converter Service
> **Target**: Menghilangkan duplikasi kode formatting konfigurasi dengan memanfaatkan langsung engine murni Go `github.com/LalatinaHub/common/subconverter`.

- [x] **4.1. Integrasi Converter Clash Meta (Mihomo)**:
  - Gunakan `subconverter.New(nodes).ToClash(template)` atau `ToClashMap(template)` dari pustaka `common`.
  - Otomatis menghasilkan struktur standar:
    - Injeksi definisi `proxies` (Shadowsocks, VMess, VLESS, Trojan).
    - Proxy Groups lengkap: `PROXIES` (select), `AUTO-FALLBACK` (url-test 300s), `LOAD-BALANCE` (round-robin).
    - Dukungan template `"cf"` untuk injeksi Trojan UDP relay detour otomatis.
    - Rules standar: LAN bypass, MATCH rules.
- [x] **4.2. Integrasi Converter sing-box v1.14+**:
  - Gunakan `subconverter.New(nodes).ToSingbox("standard", template)` dari `common/subconverter`.
  - Menghasilkan dokumen JSON sing-box lengkap:
    - DNS resolver (`remote-dns` 1.1.1.1 dan `direct-dns` 223.5.5.5).
    - Inbound `mixed-in` (SOCKS5/HTTP port 2080).
    - Outbounds (`selector`, `urltest`, direct, block, dns-out).
    - Route rules: DNS hijack, private IP bypass, auto-detect-interface.
    - Dukungan template `"cf"` untuk detour Trojan UDP.
- [x] **4.3. Integrasi Converter Format Mobile sing-box (SFA / BFR)**:
  - Gunakan `conv.ToSFA(template)` (Sing-Box For Android) dan `conv.ToBFR(template)` (Box For Root) untuk profil kompatibilitas mobile client lama (grup `Internet`, `Lock Region ID`, `Best Latency`).
- [x] **4.4. Integrasi Converter Base64 & Raw Lines**:
  - Gunakan `subconverter.New(nodes).ToBase64()` untuk aplikasi mobile legacy (v2rayNG, Shadowrocket, NekoBox) yang membutuhkan string Base64 dari URI `vmess://`, `vless://`, `trojan://`, dan `ss://`.
  - Gunakan `subconverter.New(nodes).ToRawString()` untuk format raw plain-text baris per baris.
- [x] **4.5. Adapter Service `internal/service/converter`**:
  - Membungkus pemanggilan `common/subconverter` dengan layer caching (sync.Map / LRU) jika daftar node identik, dan menyetel MIME `Content-Type` yang sesuai (`application/yaml`, `application/json`, `text/plain`).

---

### 🧠 Fase 5: Sistem Deteksi User-Agent & Header Langganan Standar
> **Target**: Menghilangkan setup manual bagi user melalui negosiasi format otomatis dan metadata profil.

- [x] **5.1. Dynamic User-Agent Detector**:
  - Evaluasi header `User-Agent` dari request client jika parameter query `format` tidak ditentukan:
    | Pola User-Agent | Target Panggilan Converter | MIME Content-Type | Kategori Client |
    |---|---|---|---|
    | `Clash.Meta`, `mihomo`, `ClashVerge`, `Flclash`, `ClashX`, `ClashForAndroid` | `conv.ToClash("")` | `application/yaml; charset=utf-8` | Clash-family |
    | `sing-box`, `SFI`, `SFM`, `sing-box-tools` | `conv.ToSingbox("standard", "")` | `application/json; charset=utf-8` | sing-box modern |
    | `SFA`, `BFR` | `conv.ToSFA("")` / `conv.ToBFR("")` | `application/json; charset=utf-8` | sing-box mobile legacy |
    | `v2rayNG`, `Shadowrocket`, `v2rayN`, `NekoBox`, `NekoRay`, `Streisand` | `conv.ToBase64()` | `text/plain; charset=utf-8` | Legacy URI clients |
    | Browser (`Mozilla`, `Chrome`, `Safari`), `curl`, `Wget` | `conv.ToRawString()` / Web Redirect | `text/plain; charset=utf-8` | Terminal / Browser |
- [x] **5.2. Standard Subscription Headers**:
  - Set header HTTP sesuai standar internasional subscription VPN:
    - `Subscription-Userinfo: upload=0; download={used_bytes}; total={quota_bytes}; expire={unix_timestamp}`
      *(Kalkulasi `total = user.Quota * 1024 * 1024` karena quota disimpan dalam MB)*.
    - `Profile-Update-Interval: 24` (interval auto-update 24 jam).
    - `Profile-Title: LatinaHub - {UserTag}`.
    - `Content-Disposition: attachment; filename="{filename}"` (`.yaml`, `.json`, atau `.txt` sesuai format).
    - `Cache-Control: private, no-cache, no-transform`.

---

### ⚙️ Fase 6: Endpoint Manajemen User, DB Admin & Converter Tool
> **Target**: Mempertahankan fungsionalitas pendukung sistem dari `megalodon-api` dengan arsitektur bersih dan pustaka sentral `common`.

- [x] **6.1. Endpoint `GET /user/:apiToken/:id`**:
  - Verifikasi `apiToken` terhadap database menggunakan `kvRepo.GetValueByKey(ctx, "apiToken")`.
  - Cari user berdasarkan `id` via `userRepo.GetUserByID(ctx, id)`.
  - Jika belum ada, lakukan auto-provisioning akun baru via `userRepo.CreateUserWithDefaults(ctx, id)` (memanfaatkan `common/repository.UserRepository.CreateUser`).
  - Kembalikan data user dalam format JSON.
- [x] **6.2. Endpoint `POST /db/:apiToken/exec`**:
  - Verifikasi `apiToken` terhadap tabel `kv`.
  - Ambil instance DB via `database.GetDB()`, jalankan transaksi terisolasi (`tx, err := db.BeginTx(ctx, nil)`).
  - Eksekusi rangkaian query dengan rollback otomatis pada error dan commit saat sukses.
  - Sanitasi pesan error untuk mencegah database credential leakage.
- [x] **6.3. Utility & Parsing Endpoints**:
  - `GET /health` & `GET /api/v1/ping`: Health check endpoint menggunakan `database.Ping(ctx)`.
  - `GET /api/v1/info`: Informasi IP & GeoIP edge/cluster.
  - `POST /api/v1/convert`: Konversi URL proxy sembarang ke format yang diinginkan menggunakan `common/subconverter.NewFromRaw(rawConfig)` dan `common/proxy.Parser`.

---

### 🛡️ Fase 7: Standar Keandalan, Keamanan & Konkurensi (Golang Pro)
> **Target**: Memenuhi standar kinerja tinggi Go 1.21+ dan keandalan sistem microservice.

- [x] **7.1. Concurrency & Context Safety**:
  - Semua operasi I/O (database query, HTTP call ke edge server) wajib menyertakan `context.Context` dengan batas timeout terukur.
  - Gunakan bounded worker pools (`sync.WaitGroup` + channel semaphore) untuk request concurrent.
- [x] **7.2. In-Memory Caching**:
  - Thread-safe memory cache (menggunakan `sync.Map` atau `sync.RWMutex`) untuk data edge server info (`/api/v1/info`) dan query konfigurasi `kv`.
- [x] **7.3. Rate Limiting & Proteksi DoS**:
  - Middleware Token Bucket / Sliding Window rate limiter (konsisten dengan rate limiter di `LatinaServer/internal/service/web/ratelimit.go`) untuk melindungi endpoint `/sub`.
- [x] **7.4. Profiling & Observability**:
  - Pasang endpoint debug `net/http/pprof` pada `/debug/pprof` (hanya aktif jika `APP_ENV != production`).
  - Correlation logging: teruskan `X-Request-ID` dan `X-Trace-ID` pada semua log entry zerolog.

---

### 🧪 Fase 8: Pengujian Menyeluruh (Testing & Quality Assurance)
> **Target**: Memastikan keandalan fungsional, tidak ada regresi, dan bebas race condition.

- [x] **8.1. Table-Driven Unit Tests**:
  - Test parser query & detektor User-Agent (`detector_test.go`).
  - Test adapter converter Clash, sing-box, dan Base64 (`converter_test.go`).
  - Test validasi masa aktif dan kuota akun (`subscription_test.go`).
- [x] **8.2. HTTP Handler Testing (httptest)**:
  - Pengujian endpoint Gin dengan `httptest.NewRecorder()`:
    - Token tidak valid / kosong (HTTP 403 / 400).
    - Akun kadaluwarsa atau kuota habis (HTTP 403).
    - Negosiasi User-Agent Clash, sing-box, v2rayNG (format body & header `Content-Disposition`/`Subscription-Userinfo`).
    - Overriding domain via query `cdn` dan `sni`.
- [x] **8.3. Concurrency Race Testing**:
  - Eksekusi `go test -race -v ./...` untuk memastikan tidak adanya kondisi data race.
- [x] **8.4. Benchmark Tests**:
  - Benchmark pengolahan request `/sub` dan pipeline formatting di bawah beban konkurensi tinggi.

---

## 📊 Matriks Fitur: Porting & Integrasi Ekosistem

| Fitur / Komponen | `megalodon-api` Asli | Status di `LatinaHub/common` | Implementasi di `LatinaApi` |
|---|---|---|---|
| **Pustaka Domain** | Dependensi lokal internal terpisah | Tersedia (`github.com/LalatinaHub/common`) | Dijadikan dependensi sentral (*Single Source of Truth*) |
| **Koneksi Database** | LibSQL ad-hoc via custom struct | `common/database.GetDB()` pool | Singleton pool dengan auto-index & graceful close |
| **Domain Model** | Struct User/Server terisolasi | `common/model` (`User`, `Server`, `ProxyNode`, `KV`) | Re-export type alias di `internal/domain/model` |
| **URL Proxy Parser** | Regex parser ad-hoc | `common/proxy.NewParser()` | Digunakan langsung untuk parsing & formatting proxy |
| **Clash Meta Output** | External subconverter server | `common/subconverter.ToClash()` | Native Go generator via `common/subconverter` |
| **sing-box Output** | Outbound list JSON via subconverter | `common/subconverter.ToSingbox()` | Native JSON generator (standard, SFA, BFR profiles) |
| **Base64 / Raw Output**| String concatenation manual | `common/subconverter.ToBase64()` | Native Base64 & raw lines generator |
| **IATA Geolocation** | Regex / map statis parsial | `common/region.Lookup()` (9,200+ airport) | O(1) airport lookup untuk remark node |
| **Endpoint Sub** | `GET /sub?pass=...` | - | `GET /sub` & `GET /api/v1/sub` (`?token=...` & `?pass=...`) |
| **Deteksi Klien** | Mengandalkan parameter `format` | - | Auto-detect via header `User-Agent` + fallback manual |
| **Standar Header** | Belum standar | - | `Subscription-Userinfo`, `Profile-Update-Interval`, dll. |
| **Validasi Akun** | Cek expired & kuota dasar | `user.IsActive(now)` method | Granular HTTP 403 responses dengan pesan jelas |
| **Manajemen User** | `GET /user/:apiToken/:id` | `UserRepository.CreateUser()` | Porting dengan auto-provisioning & validasi aman |
| **Eksekusi DB Admin** | `POST /db/:apiToken/exec` | Transaksional via `GetDB()` | Transaksi atomic `BeginTx()` dengan error sanitization |
| **Logging** | Gin default log | - | Structured Zero-Allocation `zerolog` with trace-id |

---

## 🚦 Verifikasi & Acceptance Criteria (Checklist Selesai)

1. [x] Repositori dapat dibangun bersih dengan `go build ./...` tanpa error atau peringatan linter.
2. [x] `GET /sub?token={valid_token}` dengan User-Agent `Clash.Meta` mengembalikan file YAML valid dari `common/subconverter` yang dapat diimpor langsung ke Clash Verge / Mihomo.
3. [x] `GET /sub?token={valid_token}` dengan User-Agent `sing-box` mengembalikan file JSON valid sesuai skema sing-box v1.14+.
4. [x] `GET /sub?token={valid_token}` dengan User-Agent `v2rayNG` mengembalikan string Base64 yang berisi URI vmess/vless/trojan yang valid.
5. [x] Header `Subscription-Userinfo` mencerminkan sisa kuota dan tanggal kadaluwarsa yang akurat dari basis data.
6. [x] Token yang sudah kedaluwarsa atau kuotanya habis mendapatkan penolakan HTTP 403 yang aman (`user.IsActive(now)`).
7. [x] Endpoint `POST /api/v1/convert` dapat mengonversi raw string URL proxy menggunakan `common/subconverter.NewFromRaw()`.
8. [x] Semua unit test lolos dengan `go test -v ./...`.


