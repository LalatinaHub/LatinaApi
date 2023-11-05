package converter

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/LalatinaHub/LatinaApi/common/account/protocol"
	"github.com/LalatinaHub/LatinaApi/common/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
	C "github.com/sagernet/sing-box/constant"
)

func ToRaw(accounts []db.DBScheme, args ...string) string {
	var result []string

	for _, account := range accounts {
		tls := ""
		if account.TLS {
			tls = "tls"
		}

		switch account.VPN {
		case C.TypeVMess:
			tls := ""
			if account.TLS {
				tls = "tls"
			}

			vmess := &protocol.VmessStruct{
				Add:            account.Server,
				Port:           uint16(account.ServerPort),
				Aid:            account.AlterId,
				Id:             account.UUID,
				Net:            account.Transport,
				Path:           account.Path,
				Ps:             account.Remark,
				Tls:            tls,
				Security:       account.Security,
				SkipCretVerify: account.Insecure,
				Sni:            account.SNI,
				Host:           account.Host,
			}

			switch account.Transport {
			case "grpc":
				vmess.Path = account.ServiceName
			}

			j, _ := json.Marshal(vmess)
			result = append(result, "vmess://"+helper.EncodeToBase64(string(j)))
		case C.TypeVLESS, C.TypeTrojan:
			u := url.URL{
				Scheme:   account.VPN,
				Host:     account.Server + ":" + strconv.Itoa(account.ServerPort),
				Fragment: account.Remark,
			}

			// Set userinfo
			switch account.VPN {
			case C.TypeVLESS:
				u.User = url.User(account.UUID)
			default:
				u.User = url.User(account.Password)
			}

			// Set queries
			q := u.Query()
			if tls == "" {
				tls = "none"
			}

			q.Set("security", tls)
			q.Set("type", account.Transport)
			q.Set("sni", account.SNI)
			q.Set("allowInsecure", "true")
			switch account.Transport {
			case C.V2RayTransportTypeWebsocket:
				q.Set("host", account.Host)
				q.Set("path", account.Path)
			case C.V2RayTransportTypeGRPC:
				q.Set("serviceName", account.ServiceName)
			}

			u.RawQuery, _ = url.PathUnescape(q.Encode())
			result = append(result, u.String())
		case C.TypeShadowsocks:
			u := url.URL{
				Scheme:   "ss",
				Host:     fmt.Sprintf("%s:%d", account.Server, account.ServerPort),
				User:     url.User(helper.EncodeToBase64(fmt.Sprintf("%s:%s", account.Method, account.Password))),
				Fragment: account.Remark,
			}

			result = append(result, u.String())
		}
	}

	return strings.Join(result, "\n")
}
