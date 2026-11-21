package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func worker(ctx context.Context, jobs <-chan int) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped:", ctx.Err())
			return
		case job := <-jobs:
			fmt.Println("start job:", job)
			time.Sleep(3 * time.Second) // 1つのジョブに3秒かかる想定
			fmt.Println("done job:", job)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		worker(ctx, jobs)
	}()

	jobs <- 1
	jobs <- 2
	time.Sleep(100 * time.Millisecond)

	cancel()
	wg.Wait()
}
