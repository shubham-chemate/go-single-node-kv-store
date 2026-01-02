package main

import (
	"fmt"
	"time"
)

func generateValues() <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)

		for i := 2; i <= 100; i += 2 {
			ch <- i
			time.Sleep(1 * time.Second)
		}
	}()
	return ch
}

func GeneratorPatternMain() {
	ch := generateValues()

	for val := range ch {
		fmt.Printf("got val: %d\n", val)
	}

}
