package converter

import (
	"github.com/LalatinaHub/LatinaSub-go/db"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func ToSing(accounts []db.DBScheme, args ...string) option.Options {
	options := option.Options{
		Outbounds: func() []option.Outbound {
			outbounds := []option.Outbound{}
			selector := option.Outbound{
				Type: C.TypeSelector,
				Tag:  "Internet",
				SelectorOptions: option.SelectorOutboundOptions{
					Outbounds: []string{},
				},
			}

			for _, outbound := range ToBfa(accounts, args...).Outbounds {
				switch outbound.Type {
				case C.TypeSelector, C.TypeURLTest:
				default:
					selector.SelectorOptions.Outbounds = append(selector.SelectorOptions.Outbounds, outbound.Tag)
					outbounds = append(outbounds, outbound)
				}
			}

			outbounds = append([]option.Outbound{selector}, outbounds...)
			return outbounds
		}(),
	}

	return options
}
