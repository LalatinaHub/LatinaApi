package converter

import (
	"net/netip"
	"strings"

	"github.com/LalatinaHub/LatinaSub-go/account"
	"github.com/LalatinaHub/LatinaSub-go/db"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	dns "github.com/sagernet/sing-dns"
)

func ToBfa(accounts []db.DBScheme) option.Options {
	var (
		tags      []string
		outbounds []option.Outbound
	)
	for _, proxy := range strings.Split(ToRaw(accounts), "\n") {
		outbound := account.New(proxy).Outbound
		outbounds = append(outbounds, outbound)
		tags = append(tags, outbound.Tag)
	}

	return option.Options{
		Log: &option.LogOptions{
			Disabled:  false,
			Level:     "error",
			Timestamp: false,
		},
		DNS: &option.DNSOptions{
			Servers: []option.DNSServerOptions{
				{
					Tag:     "dns_remote",
					Address: "8.8.8.8",
				},
			},
			DNSClientOptions: option.DNSClientOptions{
				Strategy: option.DomainStrategy(dns.DomainStrategyPreferIPv4),
			},
			Final: "dns_remote",
		},
		Inbounds: []option.Inbound{
			{
				Type: "tun",
				Tag:  "tun-in",
				TunOptions: option.TunInboundOptions{
					Inet4Address: option.Listable[option.ListenPrefix]{option.ListenPrefix(netip.MustParsePrefix("172.19.0.1/28"))},
					AutoRoute:    true,
					Stack:        "system",
					InboundOptions: option.InboundOptions{
						SniffEnabled: true,
					},
				},
			},
		},
		Outbounds: append([]option.Outbound{
			{
				Type: "direct",
				Tag:  "direct",
			},
			{
				Type: "block",
				Tag:  "block",
			},
			{
				Type: "dns",
				Tag:  "dns-out",
			},
			{
				Type: "selector",
				Tag:  "tunnel",
				SelectorOptions: option.SelectorOutboundOptions{
					Outbounds: []string{"urltest", "selector"},
				},
			},
			{
				Type: "urltest",
				Tag:  "urltest",
				URLTestOptions: option.URLTestOutboundOptions{
					Outbounds: tags,
				},
			},
			{
				Type: "selector",
				Tag:  "selector",
				SelectorOptions: option.SelectorOutboundOptions{
					Outbounds: tags,
				},
			},
			{
				Type: "selector",
				Tag:  "ads",
				SelectorOptions: option.SelectorOutboundOptions{
					Outbounds: []string{
						"block",
						"direct",
						"tunnel",
					},
				},
			},
		}, outbounds...),
		Route: &option.RouteOptions{
			Rules: []option.Rule{
				{
					Type: C.RuleTypeDefault,
					DefaultOptions: option.DefaultRule{
						Geosite:  option.Listable[string]{"category-ads-all"},
						Outbound: "ads",
					},
				},
				{
					Type: C.RuleTypeDefault,
					DefaultOptions: option.DefaultRule{
						Protocol: option.Listable[string]{"dns"},
						Port:     option.Listable[uint16]{53},
						Outbound: "dns-out",
					},
				},
			},
			Final:               "tunnel",
			FindProcess:         true,
			AutoDetectInterface: true,
			OverrideAndroidVPN:  true,
		},
		Experimental: &option.ExperimentalOptions{
			ClashAPI: &option.ClashAPIOptions{
				ExternalController: "0.0.0.0:9090",
				ExternalUI:         "yacd",
				StoreSelected:      true,
			},
		},
	}
}
