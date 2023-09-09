//go:build bench_yayamls

package benchmark

import (
	_ "github.com/KSpaceer/yayamls"
	"testing"
)

func BenchmarkYAYAMLS_Unmarshal_Large(b *testing.B) {
	b.Skip()
	b.SetBytes(int64(len(largeDataText)))
	for i := 0; i < b.N; i++ {
		var s LargeStruct
		err := s.UnmarshalYAML(largeDataText)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkYAYAMLS_Unmarshal_Small(b *testing.B) {
	b.SetBytes(int64(len(smallDataText)))
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		err := s.UnmarshalYAML(smallDataText)
		if err != nil {
			b.Error(err)
		}
	}
}
