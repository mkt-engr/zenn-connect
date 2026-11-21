package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func worker(ctx context.Context, id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d: stop requested\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			// 開始した仕事は完了まで処理する
			fmt.Printf("worker %d processing job %d\n", id, job)
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func main() {
	// Ctrl+C (SIGINT) / SIGTERM でキャンセルされるContext
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go worker(ctx, i, jobs, &wg)
	}

	// 仕事を送る側もContextを見る（新しい仕事の受付を止める）
	go func() {
		defer close(jobs)
		for job := 1; ; job++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- job:
			}
		}
	}()

	// 動作確認しやすいよう、2秒後に自動でシャットダウンを開始する
	time.AfterFunc(2*time.Second, stop)

	<-ctx.Done()
	fmt.Println("shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("all workers stopped")
	case <-shutdownCtx.Done():
		fmt.Println("shutdown timeout:", shutdownCtx.Err())
	}
}
