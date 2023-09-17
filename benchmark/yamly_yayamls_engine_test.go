//go:build bench_yamly_yayamls_engine

package benchmark

import (
	_ "github.com/KSpaceer/yamly"
	_ "github.com/KSpaceer/yamly/engines/yayamls"
	"testing"
)

func BenchmarkYamly_YAYAMLS_Engine_Unmarshal_Large(b *testing.B) {
	b.SetBytes(int64(len(largeDataText)))
	for i := 0; i < b.N; i++ {
		var s LargeStruct
		err := s.UnmarshalYAML(largeDataText)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkYamly_YAYAMLS_Engine_Unmarshal_Small(b *testing.B) {
	b.SetBytes(int64(len(smallDataText)))
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		err := s.UnmarshalYAML(smallDataText)
		if err != nil {
			b.Error(err)
		}
	}
}
