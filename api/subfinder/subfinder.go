package subfinder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
)

func SubfinderHandler(c *gin.Context) {
	var (
		domain = c.Query("domain")
	)

	subfinderOpts := &runner.Options{
		Threads:            10,
		Timeout:            30,
		MaxEnumerationTime: 10,
	}

	subfinder, err := runner.NewRunner(subfinderOpts)
	if err != nil {
		c.String(http.StatusServiceUnavailable, fmt.Sprintf("Failed to create subfinder runner: %v", err))
		return
	}

	output := &bytes.Buffer{}
	if err = subfinder.EnumerateSingleDomainWithCtx(context.Background(), domain, []io.Writer{output}); err != nil {
		c.String(http.StatusServiceUnavailable, fmt.Sprintf("Failed to enumerate domain: %v", err))
		return
	}

	c.String(http.StatusOK, output.String())
}
