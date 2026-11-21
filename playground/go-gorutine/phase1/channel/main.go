package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fmt.Println("start job:", job)
		time.Sleep(3 * time.Second) // 1つのジョブに3秒かかる想定
		fmt.Println("done job:", job)
	}
}

func main() {
	jobs := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)
	go worker(jobs, &wg)

	jobs <- 1
	jobs <- 2

	close(jobs)
	wg.Wait()
}
