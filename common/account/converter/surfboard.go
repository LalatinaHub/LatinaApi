package converter

import (
	"fmt"
	"strings"

	"github.com/LalatinaHub/LatinaSub-go/db"
	C "github.com/sagernet/sing-box/constant"
)

var baseConfig = `#!MANAGED-CONFIG URL_PLACEHOLDER interval=21600 strict=true

[General]
dns-server = system, 1.1.1.1, 1.0.0.1
skip-proxy = 127.0.0.1, 192.168.0.0/16, 10.0.0.0/8, 172.16.0.0/12, 100.64.0.0/10, localhost, *.local
proxy-test-url = http://cp.cloudflare.com
internet-test-url = http://cp.cloudflare.com
test-timeout = 10
http-listen = 0.0.0.0:1080
socks5-listen = 0.0.0.0:1081
udp-policy-not-supported-behaviour = DIRECT

[Proxy]
On = direct
Off = reject

PROXIES_PLACEHOLDER

[Rule]
PROCESS-NAME,com.whatsapp,WA
PROCESS-NAME,com.termux,Termux

FINAL,Spider

[Panel]
PanelA = title="Touch Me", content="Provided with ❤ by LalatinaHub", style=info

[Proxy Group]
Spider = select, SELECT, LOAD-BALANCE, URL-TEST, FALLBACK
WA = select, Spider, DIRECT, REJECT
Termux = select, Spider, DIRECT, REJECT
`

func ToSurfboard(accounts []db.DBScheme, args ...string) string {
	var (
		result           string
		remarks, proxies []string
		modes            []string = []string{"SELECT", "URL-TEST", "FALLBACK", "LOAD-BALANCE"}
	)

	for _, account := range accounts {
		var proxy string
		account.Remark = strings.ReplaceAll(account.Remark, ",", "")
		remarks = append(remarks, account.Remark)

		switch account.VPN {
		case C.TypeVMess:
			proxy = fmt.Sprintf("%s=%s,%s,%d,username=%s,udp-relay=true,tls=%t,skip-cert-verify=%t,sni=%s", account.Remark, account.VPN, account.Server, account.ServerPort, account.UUID, account.TLS, true, account.SNI)
		case C.TypeTrojan:
			proxy = fmt.Sprintf("%s=%s,%s,%d,password=%s,udp-relay=true,tls=%t,skip-cert-verify=%t,sni=%s", account.Remark, account.VPN, account.Server, account.ServerPort, account.Password, account.TLS, true, account.SNI)
		case C.TypeShadowsocks:
			obfsMode := "http"

			if account.TLS {
				obfsMode = "tls"
			}

			proxy = fmt.Sprintf("%s=%s,%s,%d,encrypt-method=%s,password=%s,udp-relay=true,obfs=%s,obfs-host=%s,obfs-uri=/", account.Remark, account.VPN, account.Server, account.ServerPort, account.Method, account.Password, obfsMode, account.Host)
		}

		switch account.Transport {
		case C.V2RayTransportTypeWebsocket:
			proxy = fmt.Sprintf("%s,ws=%t,ws-path=%s,ws-headers=Host:%s", proxy, true, account.Path, account.Host)
		}

		proxies = append(proxies, proxy)
	}

	result = strings.Replace(baseConfig, "PROXIES_PLACEHOLDER", strings.Join(proxies, "\n"), 1)
	for _, mode := range modes {
		result = fmt.Sprintf("%s\n%s", result, fmt.Sprintf("%s=%s,%s", mode, strings.ToLower(mode), strings.Join(remarks, ",")))
	}

	return result
}
