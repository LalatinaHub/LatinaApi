package getRoute

import (
	"net/http"
	"strings"

	apiHelper "github.com/LalatinaHub/LatinaApi/api/helper"
	"github.com/LalatinaHub/LatinaApi/common/account"
	"github.com/LalatinaHub/LatinaApi/common/account/converter"
	"github.com/LalatinaHub/LatinaApi/common/helper"
	"github.com/gin-gonic/gin"
)

func GetHandler(c *gin.Context) {
	format := c.Query("format")
	cdn := strings.Split(c.DefaultQuery("cdn", ""), ",")
	sni := strings.Split(c.DefaultQuery("sni", ""), ",")

	// Build headers and filters
	disposition := "filename=FUCKMETILLDAYLIGHT"
	filter := helper.BuildFilter(c)
	proxies := account.Get(filter)

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
		c.String(http.StatusOK, converter.ToClash(proxies))
	case "surfboard":
		c.String(http.StatusOK, strings.Replace(converter.ToSurfboard(proxies), "URL_PLACEHOLDER", apiHelper.GetRequestedURL(c), 1))
	case "raw":
		c.String(http.StatusOK, converter.ToRaw(proxies))
	case "bfa", "sfa":
		c.JSON(http.StatusOK, converter.ToBfa(proxies))
	default:
		c.JSON(http.StatusOK, proxies)
	}
}
