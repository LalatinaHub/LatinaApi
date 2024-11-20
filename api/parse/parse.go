package parseRoute

import (
	"encoding/json"
	"net/http"
	"strings"

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
	)

	if err := c.ShouldBind(&data); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	outbounds, err := provider.Parse(strings.Join(strings.Split(data.Urls, ","), "\n"))
	if err == nil && len(outbounds) > 0 {
		for _, outbound := range outbounds {
			accounts = append(accounts, *account.New(outbound))
		}
	}

	if _, err := json.Marshal(accounts); err != nil {
		c.String(http.StatusBadRequest, err.Error())
	} else {
		c.JSON(http.StatusOK, accounts)
	}
}
