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
	var remain_hw HotWater
	var remain_gb GroundBean
	cups := 4 * CupsCoffee
	var mu_coffee sync.Mutex
	go func() {
		for {
			select {
			case hw := <-ch_HotWater:
				hotWater += hw
				remain_hw += hw
			case gb := <-ch_GroundBean:
				groundBeans += gb
				remain_gb += gb
			case <-quit:
				return
			}
			if remain_hw >= cups.HotWater() && remain_gb >= cups.GroundBeans() {
				wg.Add(1)
				remain_hw -= cups.HotWater()
				remain_gb -= cups.GroundBeans()
				go func() {
					mu_coffee.Lock()
					defer mu_coffee.Unlock()
					coffee += brew(ctx, cups.HotWater(), cups.GroundBeans())
					wg.Done()
				}()
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
