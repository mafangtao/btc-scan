package main

import (
	"fmt"
	"sync"
	"time"
)

func test() {
	ch1 := make(chan int, 100)
	ch2 := make(chan string, 100)

	go func() {
		for i := 0; i < 10; i++ {
			ch1 <- i
		}
		close(ch1)

		for i := 0; i < 3; i++ {
			ch2 <- fmt.Sprintf("mmmm-%d", i)
		}
		close(ch2)
	}()

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(d int) {
			defer wg.Done()

			over := false
			for !over {
				select {
				case c1, ok := <-ch1:
					if ok {
						fmt.Printf("%d c1= %d\n", d, c1)
						time.Sleep(time.Second / 10)
					} else {
						fmt.Printf("%d c1 over\n", d)
						over = true
					}
				default:

				}
			}

			//fmt.Printf("%d c1---------------------\n", d)

			//over = false
			//for !over {
			//	select {
			//	case c2, ok := <-ch2:
			//		if ok {
			//			fmt.Printf("%d c2 = %s\n",d, c2)
			//		} else {
			//			fmt.Printf("d%d c2 over\n", d)
			//			over = true
			//		}
			//	}
			//}

			//fmt.Printf("%d c2 ---------------------\n", d)

		}(i)
	}

	wg.Wait()

	fmt.Println("ok")

}

func main() {
	test()
}
