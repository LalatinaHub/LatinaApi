package apiHelper

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func GetRequestedURL(c *gin.Context) string {
	var queries string
	schema := "http://"
	if c.Request.TLS != nil {
		schema = "https://"
	}

	result := schema + c.Request.Host + c.Request.URL.Path
	for key, value := range c.Request.URL.Query() {
		if queries == "" {
			queries = fmt.Sprintf("?%s=%s", key, url.QueryEscape(value[0]))
		} else {
			queries = fmt.Sprintf("%s&%s=%s", queries, key, url.QueryEscape(value[0]))
		}
	}

	return result + queries
}

func Fetch(link string) (*http.Response, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return httpClient.Get(link)
}
