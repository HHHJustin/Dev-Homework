# Fibonacci 2

# Answer & Result

```go
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
```

輸出結果為：

```go
1

1

2
quit
Enter quit!
Done! It took 1.735835917 seconds!
```

# Explain

```go
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
```

- select
	- 當ch傳入x時，更新x及y的數值
	- 當quit傳入數值，關閉通道並且結束fib的goroutin

```go
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
```

- 當input是quit時，quit通道傳入true，fib的goroutine則會關閉。
- else：將會把ch通道中的數值打印出來。