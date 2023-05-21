package converter

import (
	"fmt"
	"os"
	"strings"

	"github.com/LalatinaHub/LatinaSub-go/db"
	C "github.com/sagernet/sing-box/constant"
)

func ToSurfboard(accounts []db.DBScheme, args ...string) string {
	var (
		result           string
		remarks, proxies []string
		modes            []string = []string{"SELECT", "URL-TEST", "FALLBACK", "LOAD-BALANCE"}
	)

	baseConfig, _ := os.ReadFile("resources/config/surfboard.conf")

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
			// WIP
			// proxy = fmt.Sprintf("%s=%s,%s,%d,encrypt-method=%s,password=%s,udp-relay=true,obfs=%s,obfs-host=%s,obfs-uri=%s", account.Remark, account.VPN, account.Server, account.ServerPort, account.Method, account.Password, obfs, account.Host, account.Path)
		}

		switch account.Transport {
		case C.V2RayTransportTypeWebsocket:
			proxy = fmt.Sprintf("%s,ws=%t,ws-path=%s,ws-headers:Host=%s", proxy, true, account.Path, account.Host)
		}

		proxies = append(proxies, proxy)
	}

	result = strings.Replace(string(baseConfig), "PROXIES_PLACEHOLDER", strings.Join(proxies, "\n"), 1)
	for _, mode := range modes {
		result = fmt.Sprintf("%s\n%s", result, fmt.Sprintf("%s=%s,%s", mode, strings.ToLower(mode), strings.Join(remarks, ",")))
	}

	return result
}
