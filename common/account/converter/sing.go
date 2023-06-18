package converter

import (
	"github.com/LalatinaHub/LatinaSub-go/db"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func ToSing(accounts []db.DBScheme, args ...string) option.Options {
	options := option.Options{
		Outbounds: func() []option.Outbound {
			var (
				proxyTags = []string{}
				outbounds = []option.Outbound{}
			)

			for _, outbound := range ToBfa(accounts, args...).Outbounds {
				switch outbound.Type {
				case C.TypeSelector, C.TypeURLTest:
					continue
				case C.TypeDirect, C.TypeBlock, C.TypeDNS:
				default:
					proxyTags = append(proxyTags, outbound.Tag)
				}
				outbounds = append(outbounds, outbound)
			}

			controller := []option.Outbound{
				{
					Type: C.TypeSelector,
					Tag:  "Internet",
					SelectorOptions: option.SelectorOutboundOptions{
						Outbounds: proxyTags,
					},
				},
				{
					Type: C.TypeURLTest,
					Tag:  "Internet - UrlTest",
					URLTestOptions: option.URLTestOutboundOptions{
						Outbounds: proxyTags,
					},
				},
			}

			outbounds = append(controller, outbounds...)
			return outbounds
		}(),
	}

	return options
}
