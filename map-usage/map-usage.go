package main

import (
	"fmt"
	"runtime"
)

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Heap alloc: %v MiB\n", m.Alloc/1024/1024)
}

func TestMapShrink() {
	PrintMemUsage()

	mp := make(map[string]string)
	for i := range 10_000_000 {
		val := ""
		for range 10 {
			val = val + fmt.Sprintf("value of %d", i)
		}
		mp[fmt.Sprint(i)] = val
	}

	PrintMemUsage()

	// clear(mp)
	for k := range mp {
		delete(mp, k)
	}

	runtime.GC()
	PrintMemUsage()
}

func main() {
	TestMapShrink()
}
