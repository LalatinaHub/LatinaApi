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
	options := option.Options{
		Log: &option.LogOptions{
			Disabled:  true,
			Level:     "error",
			Timestamp: true,
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
		Outbounds: []option.Outbound{
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
				Tag:  "proxies",
				SelectorOptions: option.SelectorOutboundOptions{
					Outbounds: []string{},
				},
			},
			{
				Type: "selector",
				Tag:  "ADS",
				SelectorOptions: option.SelectorOutboundOptions{
					Outbounds: []string{
						"block",
						"direct",
						"proxies",
					},
				},
			},
		},
		Route: &option.RouteOptions{
			Rules: []option.Rule{
				{
					Type: C.RuleTypeDefault,
					DefaultOptions: option.DefaultRule{
						Geosite:  option.Listable[string]{"category-ads-all"},
						Outbound: "ADS",
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
			Final:               "proxies",
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

	for _, proxy := range strings.Split(ToRaw(accounts), "\n") {
		account := account.New(proxy)
		options.Outbounds = append(options.Outbounds, account.Outbound)
		options.Outbounds[3].SelectorOptions.Outbounds = append(options.Outbounds[3].SelectorOptions.Outbounds, account.Outbound.Tag)
	}

	return options
}
