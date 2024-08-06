package main

import (
	"fmt"
	"time"
)

func fib(ch chan<- int, quit chan bool) {
	x, y := 1, 1
	for {
		select {
		case ch <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("Enter quit!")
			close(quit)
			close(ch)
			return
		}
	}

}

func main() {
	start := time.Now()
	ch := make(chan int)
	quit := make(chan bool)
	go fib(ch, quit)
	var input string
	for {
		fmt.Scanf("%s", &input)
		if input == "quit" {
			quit <- true
			break
		} else {
			fmt.Println(<-ch)
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("Done! It took %v seconds!\n", elapsed.Seconds())
}
