package main

import (
	"fmt"
	"runtime"
	"sync"
)

/*
The upstream stages closes their outbound channel when they have sent all their values downstream.
The downstream stages keep receiving values from the inbound channel until the channel is closed.
This pattern allows each receiving stage to be written as range loop.
*** quotation Deepak kumar Gunjetti

Tüm stageler kendi channellarını olustururlar, channellarındaki degerlerin hepsini asagıya gonderırler ve channellarını kapatırlar.
Bu şekilde bir sonraki stage range ile channel kapanana kadar gelen değerleri alabilirler.
*/

func generator(done chan struct{}, nums ...int) chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}()
	return out
}

func square(done chan struct{}, in chan int) chan int {
	out := make(chan int)
	go func() {
		defer close(out)

		for n := range in {
			select {
			case out <- n * n:
			case <-done:
				return
			}
		}
	}()
	return out
}

func merge(done chan struct{}, cs ...chan int) chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	output := func(c <-chan int) {
		defer wg.Done()

		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	done := make(chan struct{})
	in := generator(done, 10000, 10, 100)

	c1 := square(done, in)
	c2 := square(done, in)

	out := merge(done, c1, c2)

	for i := range out {
		fmt.Println(i)
	}
	close(done)

	fmt.Println("Number of active goroutines : ", runtime.NumGoroutine())
}
