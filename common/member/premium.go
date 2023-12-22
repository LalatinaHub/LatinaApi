package member

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	apiHelper "github.com/LalatinaHub/LatinaApi/api/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/LalatinaHub/LatinaSub-go/geoip"
	H "github.com/LalatinaHub/LatinaSub-go/helper"
	"github.com/gin-gonic/gin"
)

type PremiumData struct {
	Id       string
	Password string
	VPN      string
	Domain   string
	Quota    int
	CC       string
	Adblock  bool
}

func GetPremiumAccount(cred any) PremiumData {
	var (
		id, pass, vpn, domain, cc sql.NullString
		quota                     sql.NullInt64
		query                     string
		adblock                   sql.NullBool
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
		err := rows.Scan(&id, &pass, &vpn, &domain, &quota, &cc, &adblock)
		if err != nil {
			fmt.Println(err)
		}
	}

	return PremiumData{
		Id:       id.String,
		Password: pass.String,
		VPN:      vpn.String,
		Domain:   domain.String,
		Quota:    int(quota.Int64),
		CC:       cc.String,
		Adblock:  adblock.Bool,
	}
}

func GenerateDBSchemes(accountData PremiumData, c *gin.Context) []db.DBScheme {
	var (
		result  []db.DBScheme
		buf     = new(strings.Builder)
		vpsInfo geoip.GeoIpJson
	)

	resp, err := apiHelper.Fetch("https://" + accountData.Domain + "/info")
	if err != nil {
		return []db.DBScheme{}
	}
	defer resp.Body.Close()

	io.Copy(buf, resp.Body)
	if resp.StatusCode == 200 {
		json.Unmarshal([]byte(buf.String()), &vpsInfo)
	}

	var (
		countries  = strings.Split(c.DefaultQuery("cc", vpsInfo.CountryCode), ",")
		ports      = strings.Split(c.DefaultQuery("port", "80,443"), ",")
		networks   = strings.Split(c.DefaultQuery("network", "ws,tcp,grpc"), ",")
		securities = strings.Split(c.DefaultQuery("tls", "1,0"), ",")
		modes      = strings.Split(c.DefaultQuery("mode", "cdn,sni"), ",")
		vpns       = strings.Split(c.DefaultQuery("vpn", "trojan,vmess,vless"), ",")
	)

	for _, cc := range countries {
		if cc != vpsInfo.CountryCode {
			continue
		}

		for _, p := range ports {
			port, _ := strconv.Atoi(p)
			for _, tls := range securities {
				for _, network := range networks {
					for _, mode := range modes {
						for _, vpn := range vpns {
							if (network == "tcp" || mode == "sni") && (port == 80) {
								continue
							} else if (tls == "0" && port == 443) || (tls == "1" && port == 80) {
								continue
							} else if (network == "tcp" || network == "grpc") && mode == "cdn" {
								continue
							} else if network == "grpc" && port == 80 {
								continue
							} else if vpn == "vless" && network == "tcp" {
								continue
							} else if accountData.VPN != vpn {
								continue
							} else if accountData.Domain == "" {
								continue
							}

							var (
								domain = accountData.Domain
								tlsstr = "TLS"
							)

							if port == 80 || tls == "0" {
								tlsstr = "NTLS"
							}

							var relayString string
							if accountData.CC != vpsInfo.CountryCode && accountData.CC != "" {
								relayString = fmt.Sprintf("%s <- ", H.CCToEmoji(accountData.CC))
							}
							remark := fmt.Sprintf("%d %s%s ✨ %s %s %s %s", len(result)+1, relayString, H.CCToEmoji(vpsInfo.CountryCode), vpsInfo.Org, strings.ToUpper(network), strings.ToUpper(mode), tlsstr)

							d := db.DBScheme{
								Server:        domain,
								Ip:            vpsInfo.Ip,
								ServerPort:    port,
								Password:      accountData.Password,
								UUID:          accountData.Password,
								Security:      "auto",
								AlterId:       0,
								Method:        "",
								Plugin:        "",
								Protocol:      "",
								ProtocolParam: "",
								OBFS:          "",
								OBFSParam:     "",
								Host:          domain,
								TLS:           tls == "1",
								Transport:     network,
								Path:          "/" + vpn,
								ServiceName:   vpn,
								Insecure:      true,
								SNI:           domain,
								Remark:        remark,
								ConnMode:      mode,
								CountryCode:   vpsInfo.CountryCode,
								Region:        vpsInfo.Region,
								Org:           vpsInfo.Org,
								VPN:           vpn,
							}

							result = append(result, d)
						}
					}
				}
			}
		}
	}

	return result
}
