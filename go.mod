module github.com/LalatinaHub/LatinaApi

go 1.20

require (
	github.com/LalatinaHub/LatinaBot v0.0.0-20230615000929-9775f021724e
	github.com/LalatinaHub/LatinaSub-go v0.0.0-20230612055535-8af29b3ddfc6
	github.com/gin-gonic/gin v1.9.0
	github.com/go-co-op/gocron v1.18.0
	github.com/sagernet/sing-box v1.2.7-0.20230520041538-9ee0f0e74e50
	github.com/sethvargo/go-password v0.2.0
	github.com/ulule/limiter/v3 v3.11.2
)

replace (
	github.com/LalatinaHub/LatinaBot => ../LatinaBot
	github.com/LalatinaHub/LatinaSub-go => ../LatinaSub-go
	github.com/sagernet/sing-box => github.com/LalatinaHub/sing-box v0.0.0-20230610102329-e349d097c3ec
)

require (
	berty.tech/go-libtor v1.0.385 // indirect
	github.com/Dreamacro/clash v1.16.0 // indirect
	github.com/Dreamacro/protobytes v0.0.0-20230324064118-87bc784139cd // indirect
	github.com/NicoNex/echotron/v3 v3.25.0 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/bytedance/sonic v1.8.5 // indirect
	github.com/caddyserver/certmagic v0.18.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/cloudflare/circl v1.3.2 // indirect
	github.com/cretz/bine v0.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-chi/chi/v5 v5.0.8 // indirect
	github.com/go-chi/cors v1.2.1 // indirect
	github.com/go-chi/render v1.0.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/gofrs/uuid/v5 v5.0.0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/pprof v0.0.0-20230309165930-d61513b1440d // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/insomniacslk/dhcp v0.0.0-20230516061539-49801966e6cb // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/libdns/libdns v0.2.1 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/mholt/acmez v1.1.1 // indirect
	github.com/miekg/dns v1.1.54 // indirect
	github.com/onsi/ginkgo/v2 v2.9.1 // indirect
	github.com/ooni/go-libtor v1.1.8 // indirect
	github.com/oschwald/maxminddb-golang v1.10.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.14 // indirect
	github.com/pires/go-proxyproto v0.7.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/qtls-go1-18 v0.2.0 // indirect
	github.com/quic-go/qtls-go1-19 v0.3.0 // indirect
	github.com/quic-go/qtls-go1-20 v0.2.0 // indirect
	github.com/sagernet/cloudflare-tls v0.0.0-20221031050923-d70792f4c3a0 // indirect
	github.com/sagernet/go-tun2socks v1.16.12-0.20220818015926-16cb67876a61 // indirect
	github.com/sagernet/netlink v0.0.0-20220905062125-8043b4a9aa97 // indirect
	github.com/sagernet/quic-go v0.0.0-20230202071646-a8c8afb18b32 // indirect
	github.com/sagernet/reality v0.0.0-20230406110435-ee17307e7691 // indirect
	github.com/sagernet/sing v0.2.5-0.20230530114415-221f066dba7c // indirect
	github.com/sagernet/sing-dns v0.1.5-0.20230426113254-25d948c44223 // indirect
	github.com/sagernet/sing-mux v0.0.0-20230517134606-1ebe6bb26646 // indirect
	github.com/sagernet/sing-shadowsocks v0.2.2-0.20230509053848-d83f8fe1194c // indirect
	github.com/sagernet/sing-shadowsocks2 v0.0.0-20230608022204-942fbd98f846 // indirect
	github.com/sagernet/sing-shadowtls v0.1.2-0.20230531025805-ebadc7615da3 // indirect
	github.com/sagernet/sing-tun v0.1.5-0.20230610004555-e881f2101310 // indirect
	github.com/sagernet/sing-vmess v0.1.5-0.20230417103030-8c3070ae3fb3 // indirect
	github.com/sagernet/smux v0.0.0-20230312102458-337ec2a5af37 // indirect
	github.com/sagernet/tfo-go v0.0.0-20230303015439-ffcfd8c41cf9 // indirect
	github.com/sagernet/utls v0.0.0-20230309024959-6732c2ab36f2 // indirect
	github.com/sagernet/websocket v0.0.0-20220913015213-615516348b4e // indirect
	github.com/sagernet/wireguard-go v0.0.0-20230420044414-a7bac1754e77 // indirect
	github.com/scjalliance/comshim v0.0.0-20230315213746-5e51f40bd3b9 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/u-root/uio v0.0.0-20230220225925-ffce2a382923 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	go4.org/netipx v0.0.0-20230303233057-f1b76eb4bb35 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20230320184635-7606e756e683 // indirect
	google.golang.org/grpc v1.55.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gvisor.dev/gvisor v0.0.0-20230609002524-f143e1baf0bb // indirect
	lukechampine.com/blake3 v1.1.7 // indirect
)

require (
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-contrib/static v0.0.1
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.12.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.2.2 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
