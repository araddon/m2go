package m2go

import (
	//"fmt"
	"testing"
)

func BenchmarkTnet(b *testing.B) {
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		tnet, err := NewTnet("5:hello,")
		if err != nil || tnet.Value.(string) != "hello" {
			panic(err)
		}
	}
}
