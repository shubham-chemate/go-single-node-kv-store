package main

import (
	"fmt"
	"time"
)

func upstreamGenerator(input []int) <-chan int {
	ch := make(chan int)
	go func() {
		for _, x := range input {
			ch <- x
			time.Sleep(1 * time.Second)
		}
		close(ch)
	}()
	return ch
}

func filterEven(srcCh <-chan int) <-chan int {
	outCh := make(chan int)
	go func() {
		defer close(outCh)
		for val := range srcCh {
			if val%2 == 0 {
				outCh <- val
			}
		}
	}()
	return outCh
}

func square(srcCh <-chan int) <-chan int {
	outCh := make(chan int)
	go func() {
		defer close(outCh)
		for val := range srcCh {
			outCh <- val * val
		}
	}()
	return outCh
}

func half(srcCh <-chan int) <-chan int {
	outCh := make(chan int)
	go func() {
		defer close(outCh)
		for val := range srcCh {
			outCh <- val / 2
		}
	}()
	return outCh
}

func PipelinePatternMain() {
	srcCh := upstreamGenerator([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})

	evenCh := filterEven(srcCh)
	sqCh := square(evenCh)
	halfCh := half(sqCh)

	for val := range halfCh {
		fmt.Printf("val %d\n", val)
	}
}
