package converter

import (
	"testing"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleNodes() []model.ProxyNode {
	return []model.ProxyNode{
		{
			VPN:        "trojan",
			Server:     "tr.example.com",
			ServerPort: 443,
			Password:   "secret",
			TLS:        true,
			Remark:     "Trojan-SG",
		},
		{
			VPN:        "vmess",
			Server:     "vm.example.com",
			ServerPort: 443,
			UUID:       "11111111-2222-3333-4444-555555555555",
			TLS:        true,
			Transport:  "ws",
			Path:       "/ws",
			Host:       "vm.example.com",
			Remark:     "VMess-WS",
		},
	}
}

func TestConverterService_AllFormats(t *testing.T) {
	svc := NewConverterService()
	nodes := sampleNodes()

	// 1. Clash
	clashYAML, err := svc.ToClash(nodes, "")
	require.NoError(t, err)
	assert.Contains(t, clashYAML, "Trojan-SG")
	assert.Contains(t, clashYAML, "VMess-WS")
	assert.Contains(t, clashYAML, "PROXIES")

	// 2. sing-box
	singboxJSON, err := svc.ToSingbox(nodes, "standard", "")
	require.NoError(t, err)
	assert.Contains(t, singboxJSON, "Trojan-SG")
	assert.Contains(t, singboxJSON, "mixed-in")

	// 3. SFA
	sfaJSON, err := svc.ToSFA(nodes, "")
	require.NoError(t, err)
	assert.Contains(t, sfaJSON, "Lock Region ID")

	// 4. BFR
	bfrJSON, err := svc.ToBFR(nodes, "")
	require.NoError(t, err)
	assert.Contains(t, bfrJSON, "Best Latency")

	// 5. Raw
	rawStr := svc.ToRawString(nodes)
	assert.Contains(t, rawStr, "trojan://")
	assert.Contains(t, rawStr, "vmess://")

	// 6. Base64
	b64 := svc.ToBase64(nodes)
	assert.NotEmpty(t, b64)
}

func TestConverterService_ConvertRaw(t *testing.T) {
	svc := NewConverterService()
	rawInput := "trojan://secret@tr.example.com:443?security=tls#Trojan-SG\nss://YWVzLTEyOC1nY206cGFzc3dvcmRAZXhhbXBsZS5jb206ODM4OA==#Example-SS"

	// Convert to Clash
	clash, err := svc.ConvertRaw(rawInput, "clash")
	require.NoError(t, err)
	assert.Contains(t, clash, "Trojan-SG")
	assert.Contains(t, clash, "Example-SS")

	// Convert to Singbox
	sb, err := svc.ConvertRaw(rawInput, "singbox")
	require.NoError(t, err)
	assert.Contains(t, sb, "Trojan-SG")

	// Convert to Base64
	b64, err := svc.ConvertRaw(rawInput, "base64")
	require.NoError(t, err)
	assert.NotEmpty(t, b64)

	// Error case with invalid raw
	_, err = svc.ConvertRaw("invalid-plain-text", "clash")
	assert.Error(t, err)
}
