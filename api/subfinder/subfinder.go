package subfinder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
)

type ResultObject struct {
	Host   string `json:"host"`
	Ip     string `json:"ip"`
	Source string `json:"source"`
}

func SubfinderHandler(c *gin.Context) {
	var (
		domain    = c.Query("domain")
		all       = c.Query("all")
		ip        = c.Query("ip")
		recursive = c.Query("recursive")
		result    = []ResultObject{}
	)

	subfinderOpts := &runner.Options{
		Threads:            10,
		Timeout:            30,
		MaxEnumerationTime: 10,
		All:                false,
		HostIP:             false,
		RemoveWildcard:     false,
		OnlyRecursive:      false,
		JSON:               true,
	}

	// Options by Query
	if all == "1" {
		subfinderOpts.All = true
	}
	if ip == "1" {
		subfinderOpts.HostIP = true
		subfinderOpts.RemoveWildcard = true
	}
	if recursive == "1" {
		subfinderOpts.OnlyRecursive = true
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

	objs := output.String()
	for _, obj := range strings.Split(objs, "\n") {
		if obj == "" {
			continue
		}

		var m ResultObject
		json.Unmarshal([]byte(obj), &m)
		result = append(result, m)
	}

	c.JSON(http.StatusOK, gin.H{
		"error":  false,
		"result": result,
	})
}
