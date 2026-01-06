package main

import (
	"fmt"
	"sync"
	"time"
)

func generateWork(input []int) <-chan int {
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

func handleCh(inCh <-chan int, outCh chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for val := range inCh {
		outCh <- val
	}
}

func fanIn(inChs ...<-chan int) <-chan int {
	opCh := make(chan int)

	var wg sync.WaitGroup
	wg.Add(len(inChs))

	for _, inCh := range inChs {
		go handleCh(inCh, opCh, &wg)
	}

	// wg is needed to give the completion signal through opCh to the consuming goroutines
	go func() {
		wg.Wait()
		close(opCh)
	}()

	return opCh
}

func FanInPatternMain() {
	iCh1 := generateWork([]int{1, 2, 3, 4, 5})
	iCh2 := generateWork([]int{6, 7, 8, 9, 10})

	opCh := fanIn(iCh1, iCh2)

	for {
		val, ok := <-opCh
		if !ok {
			break
		}
		fmt.Printf("value: %d\n", val)
	}
}
