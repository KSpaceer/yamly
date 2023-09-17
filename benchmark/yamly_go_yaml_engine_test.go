//go:build bench_yamly_go_yaml_engine

package benchmark

import (
	"gopkg.in/yaml.v3"
	"testing"

	_ "github.com/KSpaceer/yamly"
	_ "github.com/KSpaceer/yamly/engines/goyaml"
)

func BenchmarkYamly_GoYAML_Engine_Unmarshal_Large(b *testing.B) {
	b.SetBytes(int64(len(largeDataText)))
	for i := 0; i < b.N; i++ {
		var s LargeStruct
		err := yaml.Unmarshal(largeDataText, &s)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkYamly_GoYAML_Engine_Unmarshal_Small(b *testing.B) {
	b.SetBytes(int64(len(smallDataText)))
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		err := yaml.Unmarshal(smallDataText, &s)
		if err != nil {
			b.Error(err)
		}
	}
}
