---
title: Description
tags: [Dev.mentor, golang]

---

# Description

## Code
``` go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/trace"
	"time"
)

type (
	Bean       int
	GroundBean int
	Water      int
	HotWater   int
	Coffee     int
)

const (
	GramBeans          Bean       = 1
	GramGroundBeans    GroundBean = 1
	MilliLiterWater    Water      = 1
	MilliLiterHotWater HotWater   = 1
	CupsCoffee         Coffee     = 1
)

func (w Water) String() string {
	return fmt.Sprintf("%d[ml] water", int(w))
}

func (hw HotWater) String() string {
	return fmt.Sprintf("%d[ml] hot water", int(hw))
}

func (b Bean) String() string {
	return fmt.Sprintf("%d[g] beans", int(b))
}

func (gb GroundBean) String() string {
	return fmt.Sprintf("%d[g] ground beans", int(gb))
}

func (cups Coffee) String() string {
	return fmt.Sprintf("%d cup(s) coffee", int(cups))
}

// 沖泡 1 杯咖啡所需的水量
func (cups Coffee) Water() Water {
	return Water(180*cups) / MilliLiterWater
}

// 沖泡 1 杯咖啡所需的熱水量
func (cups Coffee) HotWater() HotWater {
	return HotWater(180*cups) / MilliLiterHotWater
}

// 沖泡 1 杯咖啡所需的咖啡豆量
func (cups Coffee) Beans() Bean {
	return Bean(20*cups) / GramBeans
}

// 沖泡 1 杯咖啡所需咖啡粉量
func (cups Coffee) GroundBeans() GroundBean {
	return GroundBean(20*cups) / GramGroundBeans
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatalln("Error:", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalln("Error:", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		log.Fatalln("Error:", err)
	}
	defer trace.Stop()

	RunMyProgram()
}

func RunMyProgram() {
	// 預計沖泡多少杯咖啡
	const amountCoffee = 20 * CupsCoffee

	ctx, task := trace.NewTask(context.Background(), "make coffee")
	defer task.End()

	// 材料
	water := amountCoffee.Water()
	beans := amountCoffee.Beans()

	fmt.Println(water)
	fmt.Println(beans)

	// 熱水
	var hotWater HotWater
	for water > 0 {
		water -= 600 * MilliLiterWater
		hotWater += boil(ctx, 600*MilliLiterWater)
	}
	fmt.Println(hotWater)

	// 咖啡粉
	var groundBeans GroundBean
	for beans > 0 {
		beans -= 20 * GramBeans
		groundBeans += grind(ctx, 20*GramBeans)
	}
	fmt.Println(groundBeans)

	// 沖泡咖啡
	var coffee Coffee
	cups := 4 * CupsCoffee
	for hotWater >= cups.HotWater() && groundBeans >= cups.GroundBeans() {
		hotWater -= cups.HotWater()
		groundBeans -= cups.GroundBeans()
		coffee += brew(ctx, cups.HotWater(), cups.GroundBeans())
	}

	fmt.Println(coffee)
}

// 燒開水
func boil(ctx context.Context, water Water) HotWater {
	defer trace.StartRegion(ctx, "boil").End()
	time.Sleep(400 * time.Millisecond)
	return HotWater(water)
}

// 研磨
func grind(ctx context.Context, beans Bean) GroundBean {
	defer trace.StartRegion(ctx, "grind").End()
	time.Sleep(200 * time.Millisecond)
	return GroundBean(beans)
}

// 沖泡
func brew(ctx context.Context, hotWater HotWater, groundBeans GroundBean) Coffee {
	defer trace.StartRegion(ctx, "brew").End()
	time.Sleep(1 * time.Second)
	// 少量者優先處理
	cups1 := Coffee(hotWater / (1 * CupsCoffee).HotWater())
	cups2 := Coffee(groundBeans / (1 * CupsCoffee).GroundBeans())
	if cups1 < cups2 {
		return cups1
	}
	return cups2
}

```

## 已知條件
- 花費400ms可以煮600ml的熱水
- 花費200ms可以磨20g的咖啡豆
- 煮咖啡需要花費180ml的熱水加上20g的咖啡豆
- 花費1s可以獲得4杯咖啡
- 可以同時做不同種類的事情，但同一種事情一個時間點只能執行一次(假設所有器具只有一個)

## 目標
1. 將boil、grind、brew改成可以同時執行。
2. 但是brew啟動條件是必須要有180ml的熱水及20g磨好的豆子。
3. 要煮出20杯咖啡 -> 400ms * 6 + 1000ms = 3400ms -> 預估秒數

# 解題
## 想法
1. 用goroutine煮水：-> 用ch_hotwater紀錄目前已煮的熱水 -> if water > 0 && 沒有使用boil -> start boil.
2. 用goroutine磨豆 -> 用ch_groundbean紀錄目前已磨咖啡豆 -> if beans > 0 && 沒有使用grind -> start grind.
3. 用goroutine泡咖啡 -> 持續檢查if ch_hotwater > 180 && if ch_groundbean > 20 && 沒有正在使用brew -> start brew.

## Code
```go=
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/trace"
	"sync"
	"time"
)

type (
	Bean       int
	GroundBean int
	Water      int
	HotWater   int
	Coffee     int
)

const (
	GramBeans          Bean       = 1
	GramGroundBeans    GroundBean = 1
	MilliLiterWater    Water      = 1
	MilliLiterHotWater HotWater   = 1
	CupsCoffee         Coffee     = 1
)

func (w Water) String() string {
	return fmt.Sprintf("%d[ml] water", int(w))
}

func (hw HotWater) String() string {
	return fmt.Sprintf("%d[ml] hot water", int(hw))
}

func (b Bean) String() string {
	return fmt.Sprintf("%d[g] beans", int(b))
}

func (gb GroundBean) String() string {
	return fmt.Sprintf("%d[g] ground beans", int(gb))
}

func (cups Coffee) String() string {
	return fmt.Sprintf("%d cup(s) coffee", int(cups))
}

// 沖泡 1 杯咖啡所需的水量
func (cups Coffee) Water() Water {
	return Water(180*cups) / MilliLiterWater
}

// 沖泡 1 杯咖啡所需的熱水量
func (cups Coffee) HotWater() HotWater {
	return HotWater(180*cups) / MilliLiterHotWater
}

// 沖泡 1 杯咖啡所需的咖啡豆量
func (cups Coffee) Beans() Bean {
	return Bean(20*cups) / GramBeans
}

// 沖泡 1 杯咖啡所需咖啡粉量
func (cups Coffee) GroundBeans() GroundBean {
	return GroundBean(20*cups) / GramGroundBeans
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatalln("Error:", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalln("Error:", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		log.Fatalln("Error:", err)
	}
	defer trace.Stop()

	RunMyProgram()
}

func RunMyProgram() {
	// 預計沖泡多少杯咖啡
	const amountCoffee = 20 * CupsCoffee

	ctx, task := trace.NewTask(context.Background(), "make coffee")
	defer task.End()

	// 材料
	water := amountCoffee.Water()
	beans := amountCoffee.Beans()

	fmt.Println(water)
	fmt.Println(beans)

	var wg sync.WaitGroup
	quit := make(chan bool)
	// 熱水
	var hotWater HotWater
	ch_HotWater := make(chan HotWater)
	var mu_water sync.Mutex
	for water > 0 {
		water -= 600 * MilliLiterWater
		wg.Add(1)
		go func() {
			mu_water.Lock()
			defer mu_water.Unlock()
			boil(ctx, 600*MilliLiterWater, ch_HotWater, &wg)
		}()
	}

	// 咖啡粉
	var groundBeans GroundBean
	ch_GroundBean := make(chan GroundBean)
	var mu_bean sync.Mutex
	for beans > 0 {
		beans -= 20 * GramBeans
		wg.Add(1)
		go func() {
			mu_bean.Lock()
			defer mu_bean.Unlock()
			grind(ctx, 20*GramBeans, ch_GroundBean, &wg)
		}()
	}
	var coffee Coffee
	cups := 1 * CupsCoffee
	go func() {
		for {
			select {
			case hw := <-ch_HotWater:
				hotWater += hw
				if hotWater >= cups.HotWater() && groundBeans >= cups.GroundBeans() {
					wg.Add(1)
					go func() {
						temp_coffee := min(Coffee(hotWater/cups.HotWater()), Coffee(groundBeans/cups.GroundBeans()))
						coffee += brew(ctx, temp_coffee.HotWater(), temp_coffee.GroundBeans(), &wg)
						hotWater -= temp_coffee.HotWater()
						groundBeans -= temp_coffee.GroundBeans()
					}()
				}
			case gb := <-ch_GroundBean:
				groundBeans += gb
				if hotWater >= cups.HotWater() && groundBeans >= cups.GroundBeans() {
					wg.Add(1)
					go func() {
						temp_coffee := min(Coffee(hotWater/cups.HotWater()), Coffee(groundBeans/cups.GroundBeans()))
						coffee += brew(ctx, temp_coffee.HotWater(), temp_coffee.GroundBeans(), &wg)
						hotWater -= temp_coffee.HotWater()
						groundBeans -= temp_coffee.GroundBeans()
					}()
				}
			case <-quit:
				return
			}
		}
	}()

	wg.Wait()
	quit <- true
	close(ch_HotWater)
	close(ch_GroundBean)
	close(quit)
	fmt.Println(hotWater)
	fmt.Println(groundBeans)
	fmt.Println(coffee)
}

// 燒開水
func boil(ctx context.Context, water Water, ch_HotWater chan<- HotWater, wg *sync.WaitGroup) {
	defer wg.Done()
	defer trace.StartRegion(ctx, "boil").End()
	time.Sleep(400 * time.Millisecond)
	ch_HotWater <- HotWater(water)
}

// 研磨
func grind(ctx context.Context, beans Bean, ch_GroundBean chan<- GroundBean, wg *sync.WaitGroup) {
	defer wg.Done()
	defer trace.StartRegion(ctx, "grind").End()
	time.Sleep(200 * time.Millisecond)
	ch_GroundBean <- GroundBean(beans)
}

// 沖泡
func brew(ctx context.Context, hotWater HotWater, groundBeans GroundBean, wg *sync.WaitGroup) Coffee {
	defer wg.Done()
	defer trace.StartRegion(ctx, "brew").End()
	time.Sleep(1 * time.Second)
	// 少量者優先處理
	cups1 := Coffee(hotWater / (1 * CupsCoffee).HotWater())
	cups2 := Coffee(groundBeans / (1 * CupsCoffee).GroundBeans())
	if cups1 < cups2 {
		return cups1
	}
	return cups2
}

func min(x, y Coffee) Coffee {
	if x > y {
		return y
	}
	return x
}

```