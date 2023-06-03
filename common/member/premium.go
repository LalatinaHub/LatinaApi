package member

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/LalatinaHub/LatinaSub-go/ipapi"
	"github.com/gin-gonic/gin"
	C "github.com/sagernet/sing-box/constant"
)

type PremiumData struct {
	Id       string
	Password string
	VPN      string
	Domain   string
}

func GetPremiumAccount(cred any) PremiumData {
	var (
		id, pass, vpn, domain sql.NullString
		query                 string
	)

	if reflect.TypeOf(cred).Kind() == reflect.String {
		query = fmt.Sprintf("SELECT premium.* FROM premium JOIN public.users ON premium.id = public.users.id WHERE public.users.password = '%s' AND (SELECT EXTRACT(DAY FROM NOW() - (SELECT EXPIRED FROM public.users WHERE public.users.password = '%s'))) < 1", cred, cred)
	} else {
		query = fmt.Sprintf("SELECT premium.* FROM premium JOIN public.users ON premium.id = public.users.id WHERE public.users.id = %d AND (SELECT EXTRACT(DAY FROM NOW() - (SELECT EXPIRED FROM public.users WHERE public.users.id = %d))) < 1", cred, cred)
	}

	rows, err := db.New().Conn().Query(query)
	if err != nil {
		return PremiumData{}
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&id, &pass, &vpn, &domain)
	}

	return PremiumData{
		Id:       id.String,
		Password: pass.String,
		VPN:      vpn.String,
		Domain:   domain.String,
	}
}

func GenerateDBSchemes(accountData PremiumData, c *gin.Context) []db.DBScheme {
	var (
		result  []db.DBScheme
		buf     = new(strings.Builder)
		vpsInfo ipapi.Ipapi
	)
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := httpClient.Get("https://" + accountData.Domain + "/info")
	if err != nil {
		return []db.DBScheme{}
	}
	defer resp.Body.Close()

	io.Copy(buf, resp.Body)
	if resp.StatusCode == 200 {
		json.Unmarshal([]byte(buf.String()), &vpsInfo)
	}

	var (
		ports    = strings.Split(c.DefaultQuery("port", "80,443"), ",")
		networks = strings.Split(c.DefaultQuery("network", "ws,tcp"), ",")
		tls      = c.DefaultQuery("tls", "1")
		modes    = strings.Split(c.DefaultQuery("mode", "cdn,sni"), ",")
		vpns     = strings.Split(c.DefaultQuery("vpn", "trojan"), ",")
	)

	for _, p := range ports {
		port, _ := strconv.Atoi(p)
		for _, network := range networks {
			for _, mode := range modes {
				for _, vpn := range vpns {
					if (network == "tcp" || mode == "sni") && port == 80 {
						continue
					} else if tls == "0" && port == 443 {
						continue
					} else if tls == "1" && port == 80 {
						continue
					} else if network == "ws" && mode == "sni" {
						continue
					} else if network == "tcp" && mode == "cdn" {
						continue
					} else if accountData.VPN != vpn {
						continue
					} else if accountData.Domain == "" {
						continue
					}

					d := db.DBScheme{
						Server:        accountData.Domain,
						Ip:            vpsInfo.Ip,
						ServerPort:    port,
						Security:      "auto",
						AlterId:       0,
						Method:        "",
						Plugin:        "",
						Protocol:      "",
						ProtocolParam: "",
						OBFS:          "",
						OBFSParam:     "",
						Host:          accountData.Domain,
						TLS:           tls == "1",
						Transport:     network,
						Path:          "/" + vpn,
						ServiceName:   vpn,
						Insecure:      true,
						SNI:           accountData.Domain,
						Remark:        vpsInfo.Org + " " + network + mode,
						ConnMode:      mode,
						CountryCode:   vpsInfo.CountryCode,
						Region:        vpsInfo.Region,
						Org:           vpsInfo.Org,
						VPN:           vpn,
					}

					switch vpn {
					case C.TypeTrojan, C.TypeVLESS:
						d.Password = accountData.Password
					default:
						d.UUID = accountData.Password
					}

					result = append(result, d)
				}
			}
		}
	}

	return result
}
