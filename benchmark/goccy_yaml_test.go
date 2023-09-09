//go:build bench_goccy

package benchmark

import (
	"github.com/goccy/go-yaml"
	"testing"
)

func BenchmarkGoccyYAML_Unmarshal_Large(b *testing.B) {
	b.Skip()
	b.SetBytes(int64(len(largeDataText)))
	for i := 0; i < b.N; i++ {
		var s LargeStruct
		err := yaml.Unmarshal(largeDataText, &s)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkGoccyYAML_Unmarshal_Small(b *testing.B) {
	b.SetBytes(int64(len(smallDataText)))
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		err := yaml.Unmarshal(smallDataText, &s)
		if err != nil {
			b.Error(err)
		}
	}
}
