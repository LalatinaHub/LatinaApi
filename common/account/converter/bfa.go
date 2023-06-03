package converter

import (
	"fmt"
	"io"
	"strings"

	apiHelper "github.com/LalatinaHub/LatinaApi/api/helper"
	"github.com/LalatinaHub/LatinaSub-go/account"
	"github.com/LalatinaHub/LatinaSub-go/db"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func ToBfa(accounts []db.DBScheme, args ...string) option.Options {
	var (
		baseConfig      = "https://raw.githubusercontent.com/iyarivky/sing-ribet/main/config/config.json"
		tfo, xudp  bool = false, false
		tags       []string
		outbounds  []option.Outbound
		mux        *option.MultiplexOptions
	)

	for _, arg := range args {
		switch arg {
		case "tfo":
			tfo = true
		case "xudp":
			xudp = true
		case "mux":
			mux = &option.MultiplexOptions{
				Enabled:    true,
				Protocol:   "smux",
				MaxStreams: 32,
			}
		}
	}

	for _, proxy := range strings.Split(ToRaw(accounts, args...), "\n") {
		outbound := account.New(proxy).Outbound

		switch outbound.Type {
		case C.TypeVMess:
			outbound.VMessOptions.TCPFastOpen = tfo
			outbound.VMessOptions.Multiplex = mux
			if xudp {
				outbound.VMessOptions.PacketEncoding = "xudp"
			}
		case C.TypeTrojan:
			outbound.TrojanOptions.Multiplex = mux
		case C.TypeVLESS:
			outbound.VLESSOptions.TCPFastOpen = tfo
		case C.TypeShadowsocks:
			outbound.ShadowsocksOptions.TCPFastOpen = tfo
			outbound.ShadowsocksOptions.MultiplexOptions = mux
		}

		outbounds = append(outbounds, outbound)
		tags = append(tags, outbound.Tag)
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

	options.Outbounds = append(options.Outbounds, outbounds...)
	for i, outbound := range options.Outbounds {
		switch outbound.Tag {
		case "Internet", "Lock Region ID":
			options.Outbounds[i].SelectorOptions.Outbounds = append(options.Outbounds[i].SelectorOptions.Outbounds, tags...)
		case "Best Latency":
			options.Outbounds[i].URLTestOptions.Outbounds = append(options.Outbounds[i].URLTestOptions.Outbounds, tags...)
		}
	}

	options.Experimental.ClashAPI.Secret = ""

	return options
}
