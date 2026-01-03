## Concurrency Patterns

### Generator Pattern

- Using blocking nature of channel, we are getting values from a channel as we want them

```go
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
```

### Fan In Pattern

- We take values from multiple streams and produce it so output stream

```go
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
```

- notice the importance of wg in fanIn method, it helps to tell the consumer of output channel that we are done with sending values
- also we have separate go routine for waitgroup otherwise we won't be able to return output channel (will get blocked)

### Fan Out Pattern

- one input channel is consumed by many output channles

```go
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
```

- the for loop in main function is extemely important
- we must stop the loop when we stop receiving data from all output channels, that scenario is handles very beautifully here

## Pipeline Pattern

- the idea is to process the output produced by one stage as an input to other stage
- something like ETL, we can scale each stage independently using something like worker pool pattern

## Worker Pool Pattern

- here we have set of workers that handles the work given by upstream input
- for each record in input stream we spawn a new go routine to handle that record
- here the common scenario is, we can create million+ goroutines for large input
- we can limit goroutine using blocking nature of channels (buffered)

```go
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
```