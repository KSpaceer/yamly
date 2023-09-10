//go:build bench_go_yaml

package benchmark

import (
	"gopkg.in/yaml.v3"
	"testing"
)

func BenchmarkGoYAML_Unmarshal_Large(b *testing.B) {
	b.SetBytes(int64(len(largeDataText)))
	for i := 0; i < b.N; i++ {
		var s LargeStruct
		err := yaml.Unmarshal(largeDataText, &s)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkGoYAML_Unmarshal_Small(b *testing.B) {
	b.SetBytes(int64(len(smallDataText)))
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		err := yaml.Unmarshal(smallDataText, &s)
		if err != nil {
			b.Error(err)
		}
	}
}
