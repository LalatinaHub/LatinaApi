package converter

import (
	"fmt"
	"strings"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/LalatinaHub/common/subconverter"
)

// ConverterService transforms proxy nodes into various client-compatible profile configurations.
type ConverterService interface {
	ToClash(nodes []model.ProxyNode, template string) (string, error)
	ToSingbox(nodes []model.ProxyNode, profile string, template string) (string, error)
	ToSFA(nodes []model.ProxyNode, template string) (string, error)
	ToBFR(nodes []model.ProxyNode, template string) (string, error)
	ToBase64(nodes []model.ProxyNode) string
	ToRawString(nodes []model.ProxyNode) string
	ConvertRaw(rawConfig string, format string) (string, error)
}

type converterService struct{}

// NewConverterService returns a ConverterService implementation powered by common/subconverter.
func NewConverterService() ConverterService {
	return &converterService{}
}

func (s *converterService) ToClash(nodes []model.ProxyNode, template string) (string, error) {
	conv := subconverter.New(nodes)
	return conv.ToClash(template)
}

func (s *converterService) ToSingbox(nodes []model.ProxyNode, profile string, template string) (string, error) {
	conv := subconverter.New(nodes)
	return conv.ToSingbox(profile, template)
}

func (s *converterService) ToSFA(nodes []model.ProxyNode, template string) (string, error) {
	conv := subconverter.New(nodes)
	return conv.ToSFA(template)
}

func (s *converterService) ToBFR(nodes []model.ProxyNode, template string) (string, error) {
	conv := subconverter.New(nodes)
	return conv.ToBFR(template)
}

func (s *converterService) ToBase64(nodes []model.ProxyNode) string {
	conv := subconverter.New(nodes)
	return conv.ToBase64()
}

func (s *converterService) ToRawString(nodes []model.ProxyNode) string {
	conv := subconverter.New(nodes)
	return conv.ToRawString()
}

func (s *converterService) ConvertRaw(rawConfig string, format string) (string, error) {
	conv, err := subconverter.NewFromRaw(rawConfig)
	if err != nil {
		return "", fmt.Errorf("failed to parse raw configurations: %w", err)
	}

	switch strings.ToLower(format) {
	case "clash", "clash.meta", "mihomo":
		return conv.ToClash("")
	case "singbox", "sing-box":
		return conv.ToSingbox("standard", "")
	case "sfa":
		return conv.ToSFA("")
	case "bfr":
		return conv.ToBFR("")
	case "raw":
		return conv.ToRawString(), nil
	case "base64", "b64":
		return conv.ToBase64(), nil
	default:
		// Default to Base64
		return conv.ToBase64(), nil
	}
}
