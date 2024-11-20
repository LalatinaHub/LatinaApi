package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	apiHelper "github.com/LalatinaHub/LatinaApi/api/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/LalatinaHub/LatinaSub-go/provider"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func ToBfa(accounts []db.DBScheme, args ...string) option.Options {
	var (
		baseConfig = "https://raw.githubusercontent.com/malikshi/sing-box-examples/refs/heads/main/BoxForMagisk/config.json"
		tags       []string
		outbounds  []option.Outbound
	)

	if proxies, err := provider.Parse(ToClash(accounts)); err == nil {
		for _, outbound := range proxies {
			if _, err := json.Marshal(outbound); err != nil {
				fmt.Println("Error Provider:", err.Error())
				fmt.Println("Error Parsing:", outbound.Tag)
			} else {
				outbounds = append(outbounds, outbound)
				tags = append(tags, outbound.Tag)
			}
		}
	}

	var (
		options option.Options
		buf     = new(strings.Builder)
	)

	resp, err := apiHelper.Fetch(baseConfig)
	if err != nil {
		fmt.Println(err)
		return option.Options{}
	}
	defer resp.Body.Close()

	io.Copy(buf, resp.Body)
	if resp.StatusCode == 200 {
		options.UnmarshalJSON([]byte(buf.String()))
	}

	for i, dnsRule := range options.DNS.Rules {
		rule := dnsRule.DefaultOptions

		if rule.Server == "direct-dns" {
			if rule.Outbound != nil || rule.Geosite != nil || rule.DomainSuffix != nil || rule.Network != nil {
				continue
			}
		} else {
			continue
		}

		if i+1 == len(options.DNS.Rules) {
			options.DNS.Rules = options.DNS.Rules[:i]
		} else {
			options.DNS.Rules = append(options.DNS.Rules[:i], options.DNS.Rules[i+1:]...)
		}
	}

	filteredOutbounds := []option.Outbound{}
	for _, outbound := range options.Outbounds {
		switch outbound.Tag {
		case C.TypeBlock, C.TypeDirect, C.TypeDNS, C.TypeSelector, C.TypeURLTest:
			filteredOutbounds = append(filteredOutbounds, outbound)
		}
	}

	options.Outbounds = append(filteredOutbounds, outbounds...)
	for i, outbound := range options.Outbounds {
		switch outbound.Tag {
		case "Internet", "Lock Region ID":
			options.Outbounds[i].SelectorOptions.Outbounds = append(options.Outbounds[i].SelectorOptions.Outbounds, tags...)
		case "Best Latency":
			options.Outbounds[i].URLTestOptions.Outbounds = append(options.Outbounds[i].URLTestOptions.Outbounds, tags...)
		}
	}

	// options.Route.GeoIP.DownloadURL = ""
	// options.Route.Geosite.DownloadURL = ""
	options.Experimental.ClashAPI.Secret = ""

	return options
}
