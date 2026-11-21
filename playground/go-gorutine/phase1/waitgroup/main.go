package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("start job")
	time.Sleep(3 * time.Second) // 1つのジョブに3秒かかる想定
	fmt.Println("done job")
}

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go worker(&wg)

	wg.Wait()
}
