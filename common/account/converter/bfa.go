package converter

import (
	"net/netip"
	"strconv"
	"strings"

	"github.com/LalatinaHub/LatinaSub-go/account"
	"github.com/LalatinaHub/LatinaSub-go/db"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	dns "github.com/sagernet/sing-dns"
)

func ToBfa(accounts []db.DBScheme, args ...string) option.Options {
	var (
		tfo, xudp bool   = false, false
		direct    string = ""
		tags      []string
		outbounds []option.Outbound
		routes    []option.Rule
	)

	for _, arg := range args {
		switch arg {
		case "tfo":
			tfo = true
		case "xudp":
			xudp = true
		}

		if strings.HasPrefix(arg, "direct:") {
			direct = strings.TrimPrefix(arg, "direct:")
		}
	}

	for _, proxy := range strings.Split(ToRaw(accounts, args...), "\n") {
		outbound := account.New(proxy).Outbound

		switch outbound.Type {
		case C.TypeVMess:
			outbound.VMessOptions.TCPFastOpen = tfo

			if xudp {
				outbound.VMessOptions.PacketEncoding = "xudp"
			}
		case C.TypeVLESS:
			outbound.VLESSOptions.TCPFastOpen = tfo
		case C.TypeShadowsocks:
			outbound.ShadowsocksOptions.TCPFastOpen = tfo
		}

		outbounds = append(outbounds, outbound)
		tags = append(tags, outbound.Tag)
	}

	if direct != "" {
		rule := option.Rule{
			Type: C.RuleTypeDefault,
			DefaultOptions: option.DefaultRule{
				PackageName: option.Listable[string]{},
				UserID:      option.Listable[int32]{},
				Outbound:    "direct",
			},
		}

		for _, d := range strings.Split(direct, "-") {
			if uid, _ := strconv.Atoi(d); uid > 0 {
				rule.DefaultOptions.UserID = append(rule.DefaultOptions.UserID, int32(uid))
			} else {
				rule.DefaultOptions.PackageName = append(rule.DefaultOptions.PackageName, d)
			}
		}

		routes = append(routes, rule)
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
					Detour:  "direct",
				},
			},
			DNSClientOptions: option.DNSClientOptions{
				Strategy: option.DomainStrategy(dns.DomainStrategyPreferIPv4),
			},
			Final: "dns_remote",
		},
		Inbounds: []option.Inbound{
			{
				Type: C.TypeMixed,
				Tag:  "mixed-in",
				MixedOptions: option.HTTPMixedInboundOptions{
					ListenOptions: option.ListenOptions{
						Listen:     option.NewListenAddress(netip.IPv4Unspecified()),
						ListenPort: 2080,
					},
				},
			},
			{
				Type: C.TypeTun,
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
			Rules: append([]option.Rule{
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
			}, routes...),
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
