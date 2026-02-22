package txfs

import (
	"fmt"
	"testing"
)

func BenchmarkIsDeleted(b *testing.B) {
	for _, deletes := range []int{1, 10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("deletes_%d", deletes), func(b *testing.B) {
			fs := &FS{
				changes:   map[string]changeKind{},
				tombstone: make(map[string]struct{}, deletes),
			}
			for i := 0; i < deletes; i++ {
				fs.tombstone[fmt.Sprintf("dir_%05d", i)] = struct{}{}
			}
			target := fmt.Sprintf("dir_%05d/file.txt", deletes-1)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if !fs.isDeleted(target) {
					b.Fatalf("expected target to be tombstoned")
				}
			}
		})
	}
}

func BenchmarkMarkDelete(b *testing.B) {
	for _, existing := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("existing_%d", existing), func(b *testing.B) {
			fs := &FS{
				changes:   make(map[string]changeKind, existing+1),
				tombstone: make(map[string]struct{}, 1),
			}
			for j := 0; j < existing; j++ {
				fs.changes[fmt.Sprintf("dir_%05d/file.txt", j)] = changeReplace
			}
			// Warm once so repeated calls benchmark steady-state scan cost.
			fs.markDelete("dir_99999")

			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				fs.markDelete("dir_99999")
			}
		})
	}
}
