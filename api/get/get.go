package getRoute

import (
	"net/http"
	"strings"

	apiHelper "github.com/LalatinaHub/LatinaApi/api/helper"
	"github.com/LalatinaHub/LatinaApi/common/account"
	"github.com/LalatinaHub/LatinaApi/common/account/converter"
	"github.com/LalatinaHub/LatinaApi/common/helper"
	"github.com/LalatinaHub/LatinaApi/common/member"

	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/gin-gonic/gin"
)

func GetHandler(c *gin.Context) {
	var (
		proxies  []db.DBScheme
		format   = c.Query("format")
		password = c.Query("pass")
		cdn      = strings.Split(c.DefaultQuery("cdn", ""), ",")
		sni      = strings.Split(c.DefaultQuery("sni", ""), ",")
		args     = strings.Split(c.DefaultQuery("arg", ""), ",")
	)

	// Build headers and filters
	disposition := "filename=FUCKMETILLTHEDAYLIGHT"
	filter := helper.BuildFilter(c)

	// Authenticate user
	if filter == "" || password == "" {
		c.String(http.StatusUnauthorized, "Password invalid / not provided, get one from https://foolvpn.t.me !")
		return
	}

	// Get proxies account
	switch c.Query("premium") {
	case "1":
		proxies = append(proxies, account.Get(filter)...)
	case "2":
		proxies = append(proxies, member.GenerateDBSchemes(member.GetPremiumAccount(password), c)...)
	default:
		proxies = append(proxies, member.GenerateDBSchemes(member.GetPremiumAccount(password), c)...)
		proxies = append(proxies, account.Get(filter)...)
	}

	if c.Query("ip") == "1" {
		for i := 0; i < len(proxies); i++ {
			ip := proxies[i].Ip

			if ip != "" {
				proxies[i].Server = ip
				proxies[i].Host = ip
				proxies[i].SNI = ip
			}
		}
	}

	// Populate bugs
	proxies = account.PopulateBugs(proxies, cdn, sni)

	// Set headers and filters
	c.Header("Content-Disposition", disposition)

	switch format {
	case "clash":
		c.String(http.StatusOK, converter.ToClash(proxies, args...))
	case "surfboard":
		c.String(http.StatusOK, strings.Replace(converter.ToSurfboard(proxies, args...), "URL_PLACEHOLDER", apiHelper.GetRequestedURL(c), 1))
	case "raw":
		c.String(http.StatusOK, converter.ToRaw(proxies, args...))
	case "bfa", "sfa":
		c.JSON(http.StatusOK, converter.ToSfa(proxies, args...))
	case "sing":
		c.JSON(http.StatusOK, converter.ToSing(proxies, args...))
	default:
		c.JSON(http.StatusOK, proxies)
	}
}
