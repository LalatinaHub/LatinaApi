package subfinder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
)

type ScanResult struct {
	Domains []string `json:"domains"`
}

func SubfinderHandler(c *gin.Context) {
	var (
		domain = c.Query("domain")
		result = ScanResult{}
	)

	subfinderOpts := &runner.Options{
		Threads:            10,
		Timeout:            30,
		MaxEnumerationTime: 10,
	}

	subfinder, err := runner.NewRunner(subfinderOpts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": fmt.Sprintf("Failed to create subfinder runner: %v", err),
		})
		return
	}

	output := &bytes.Buffer{}
	if err = subfinder.EnumerateSingleDomainWithCtx(context.Background(), domain, []io.Writer{output}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": fmt.Sprintf("Failed to enumerate domain: %v", err),
		})
		return
	}

	for _, sub := range strings.Split(output.String(), "\n") {
		if sub != "" {
			result.Domains = append(result.Domains, sub)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"error":  false,
		"result": result,
	})
}
