package main

import (
	"fmt"
	"sync"
	"time"
)

func createWork(workCh chan<- int) {
	for i := range 100 {
		workCh <- i
	}
	close(workCh)
}

type sq struct {
	num   int
	numSq int
	err   error
}

func worker(num int, opCh chan<- sq, wg *sync.WaitGroup, limiter <-chan struct{}) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("worker crashed: %d [%v]\n", num, r)
		}
	}()
	defer func() { <-limiter }()

	time.Sleep(1 * time.Second)

	if num == 5 {
		opCh <- sq{num, -1, fmt.Errorf("error while processing")}
		panic("ah 5 it is!")
	}

	opCh <- sq{num, num * num, nil}
}

func printWork(opCh <-chan sq) {
	for op := range opCh {
		fmt.Printf("Received Output: %d square is %d\n", op.num, op.numSq)
	}
}

func WorkerPoolPatternMain() {
	workCh := make(chan int)
	opCh := make(chan sq)

	go createWork(workCh)
	go printWork(opCh)

	var wg sync.WaitGroup

	limit := 5
	goroutineLimiter := make(chan struct{}, limit)

	for work := range workCh {
		wg.Add(1)
		goroutineLimiter <- struct{}{}
		go worker(work, opCh, &wg, goroutineLimiter)
	}

	wg.Wait()
}
