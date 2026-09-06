package converter

import (
	"testing"
)

func BenchmarkToClash(b *testing.B) {
	svc := NewConverterService()
	nodes := sampleNodes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ToClash(nodes, "")
	}
}

func BenchmarkToSingbox(b *testing.B) {
	svc := NewConverterService()
	nodes := sampleNodes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ToSingbox(nodes, "standard", "")
	}
}

func BenchmarkToBase64(b *testing.B) {
	svc := NewConverterService()
	nodes := sampleNodes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.ToBase64(nodes)
	}
}
