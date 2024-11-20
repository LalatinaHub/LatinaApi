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

			for _, outbound := range ToSfa(accounts, args...).Outbounds {
				switch outbound.Type {
				case C.TypeTrojan, C.TypeVMess, C.TypeShadowsocks, C.TypeVLESS, C.TypeHysteria2:
					proxyTags = append(proxyTags, outbound.Tag)
					outbounds = append(outbounds, outbound)
				}
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
					Tag:  "Internet - URLTest",
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
