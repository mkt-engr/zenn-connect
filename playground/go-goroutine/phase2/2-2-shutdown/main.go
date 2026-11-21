package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// convertImage は画像変換を想定した関数。受け取った数値を画像IDとして使い、サムネイルのファイル名を返す
func convertImage(imageID int) string {
	// ここでは重い処理を想定して3秒待機しているだけで、実際の処理（API呼び出し・DB保存・画像変換など）は行っていない。
	time.Sleep(3 * time.Second)
	return fmt.Sprintf("image_%d_thumb.jpg", imageID)
}

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
			fmt.Printf("worker %d: start 画像 %d を変換\n", id, job)
			thumbnail := convertImage(job)
			fmt.Printf("worker %d: done %s を作成\n", id, thumbnail)
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
				fmt.Println("main: 新しい仕事の受付を止めた")
				return
			case jobs <- job:
				fmt.Printf("main: job %d を渡した\n", job)
			}
		}
	}()

	// 動作確認しやすいよう、2秒後に自動でシャットダウンを開始する
	time.AfterFunc(2*time.Second, stop)

	<-ctx.Done()
	fmt.Println("main: shutdown requested（処理中の仕事の完了を待つ）")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("main: all workers stopped")
	case <-shutdownCtx.Done():
		fmt.Println("main: shutdown timeout:", shutdownCtx.Err())
	}
}
