package apiHelper

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func DownloadFile(fullURLFile string) string {
	fileURL, err := url.Parse(fullURLFile)
	if err != nil {
		log.Fatal(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(fullURLFile)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	size, _ := io.Copy(file, resp.Body)

	defer file.Close()

	fmt.Printf("Downloaded a file %s with size %d\n", fileName, size)
	return fileName
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
