package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func fib(number float64, ch_i chan int, ch_v chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	x, y := 1.0, 1.0
	for i := 0; i < int(number); i++ {
		x, y = y, x+y
	}

	r := rand.Intn(3)
	time.Sleep(time.Duration(r) * time.Second)
	ch_i <- int(number)
	ch_v <- int(x)

}

func main() {
	start := time.Now()
	ch_i := make(chan int)
	ch_v := make(chan int)
	var wg sync.WaitGroup
	go func() {
		for i := 1; i < 15; i++ {
			wg.Add(1)
			go fib(float64(i), ch_i, ch_v, &wg)
		}
		wg.Wait()
		close(ch_i)
		close(ch_v)
	}()
	result := make(map[int]int)
	for key := range ch_i {
		value := <-ch_v
		result[key] = value
	}
	for i := 1; i < 15; i++ {
		fmt.Printf("Fib(%v): %v\n", i, result[i])
	}
	elapsed := time.Since(start)
	fmt.Printf("Done! It took %v seconds!\n", elapsed.Seconds())
}
