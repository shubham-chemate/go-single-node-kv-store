package main

import (
	"fmt"
	"time"
)

func generateInput(input []int) <-chan int {
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

func fanOut(inCh <-chan int) <-chan int {
	outCh := make(chan int)
	go func() {
		defer close(outCh)
		for val := range inCh {
			outCh <- val
		}
	}()
	return outCh
}

func FanOutMain() {
	inCh := generateInput([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	outCh1 := fanOut(inCh)
	outCh2 := fanOut(inCh)
	outCh3 := fanOut(inCh)
	outCh4 := fanOut(inCh)

	// checking until every channel is closed -> nil channel pattern
	for outCh1 != nil || outCh2 != nil || outCh3 != nil || outCh4 != nil {
		select {
		case val1, ok := <-outCh1:
			if !ok {
				// closure
				outCh1 = nil
				fmt.Println("outCh1 closed")
				continue
			}
			fmt.Printf("got %d from outCh1\n", val1)
		case val2, ok := <-outCh2:
			if !ok {
				// closure
				outCh2 = nil
				fmt.Println("outCh2 closed")
				continue
			}
			fmt.Printf("got %d from outCh2\n", val2)
		case val3, ok := <-outCh3:
			if !ok {
				// closure
				outCh3 = nil
				fmt.Println("outCh3 closed")
				continue
			}
			fmt.Printf("got %d from outCh3\n", val3)
		case val4, ok := <-outCh4:
			if !ok {
				// closure
				outCh4 = nil
				fmt.Println("outCh4 closed")
				continue
			}
			fmt.Printf("got %d from outCh4\n", val4)
		}
	}
}
