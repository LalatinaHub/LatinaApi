package parseRoute

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	apiHelper "github.com/LalatinaHub/LatinaApi/api/helper"
	"github.com/LalatinaHub/LatinaSub-go/account"
	"github.com/LalatinaHub/LatinaSub-go/provider"
	"github.com/gin-gonic/gin"
)

type PostData struct {
	Urls string `form:"urls"`
}

func ParseHandler(c *gin.Context) {
	var (
		data     PostData
		accounts []account.Account
		buf      *strings.Builder = new(strings.Builder)
	)

	if err := c.ShouldBind(&data); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	urls := strings.Join(strings.Split(data.Urls, ","), "|")
	resp, err := apiHelper.Fetch("https://sub.bonds.id/sub2?target=clash&insert=false&url=" + url.QueryEscape(urls))
	if err == nil || resp.StatusCode == 200 {
		io.Copy(buf, resp.Body)

		outbounds, err := provider.Parse(buf.String())
		if err == nil && len(outbounds) > 0 {
			for _, outbound := range outbounds {
				accounts = append(accounts, *account.New(outbound))
			}
		}
	}

	if _, err := json.Marshal(accounts); err != nil {
		c.String(http.StatusBadRequest, err.Error())
	} else {
		c.JSON(http.StatusOK, accounts)
	}
}
