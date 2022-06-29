package main
import ("testing")

func BenchmarkAdd(b *testing.B ){
	for i:=0;i<b.N;i++ {
		add(1,2)
	}
}