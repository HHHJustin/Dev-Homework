# Fibonacci 1

# Question

下面程式是計算Fabonacci數列的function：

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

func fib(number float64) float64 {
    x, y := 1.0, 1.0
    for i := 0; i < int(number); i++ {
        x, y = y, x+y
    }

    r := rand.Intn(3)
    time.Sleep(time.Duration(r) * time.Second)

    return x
}

func main() {
    start := time.Now()

    for i := 1; i < 15; i++ {
        n := fib(float64(i))
    fmt.Printf("Fib(%v): %v\n", i, n)
    }

    elapsed := time.Since(start)
    fmt.Printf("Done! It took %v seconds!\n", elapsed.Seconds())
}
```

執行結果：

```go
Fib(1): 1
Fib(2): 2
Fib(3): 3
Fib(4): 5
Fib(5): 8
Fib(6): 13
Fib(7): 21
Fib(8): 34
Fib(9): 55
Fib(10): 89
Fib(11): 144
Fib(12): 233
Fib(13): 377
Fib(14): 610
Done! It took 13.0068773 seconds!
```

必須花費13秒多的時間，使用併發的方式加速程式的運行速度。

# Answer & Result

```go
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
	for i := 1; i < 15; i++ {
		wg.Add(1)
		go fib(float64(i), ch_i, ch_v, &wg)
	}
	go func() {
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
```

輸出結果為：

```go
Fib(1): 1
Fib(2): 2
Fib(3): 3
Fib(4): 5
Fib(5): 8
Fib(6): 13
Fib(7): 21
Fib(8): 34
Fib(9): 55
Fib(10): 89
Fib(11): 144
Fib(12): 233
Fib(13): 377
Fib(14): 610
Done! It took 2.0037537 seconds!
```

# Explain

```go
ch_i := make(chan int)
ch_v := make(chan int)
```

- 使用兩個unbuffer channel(ch_i及ch_v)，分別用來存放index及Fib所計算出來的Fib value。

```go
var wg sync.WaitGroup
for i := 1; i < 15; i++ {
		wg.Add(1)
		go fib(float64(i), ch_i, ch_v, &wg)
	}
```

- 使用WaitGroup防止前台比後台先跑完。
- 將每次gib都當作一次goroutine，並個goroutine WaitGroup 都 + 1。

```go
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
```

- fib function input新增ch_i, ch_v, wg變數。
- 依據number計算出數值之後，分別將number也就是代表費式數列的index放在ch_i、將數值存放在ch_v通道中。
- 每次goroutine結束之後，用wg.Done將WaitGroup - 1。

```go
	go func() {
		wg.Wait()
		close(ch_i)
		close(ch_v)
	}()
```

- 同時也將wg.Wait(), close通道變成goroutine，等待WaitGroup變成0之後，關閉兩個通道。

```go
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
```

- 為了讓Fib可以依序顯示，利用map將記錄其index及value。最後依序將其Print出來得到結果。