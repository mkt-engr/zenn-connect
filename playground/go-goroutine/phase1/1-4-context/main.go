package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// convertImage は画像変換を想定した関数。受け取った数値を画像IDとして使い、サムネイルのファイル名を返す
func convertImage(imageID int) string {
	// ここでは重い処理を想定して3秒待機しているだけで、実際の処理（API呼び出し・DB保存・画像変換など）は行っていない。
	time.Sleep(3 * time.Second)
	return fmt.Sprintf("image_%d_thumb.jpg", imageID)
}

func worker(ctx context.Context, id int, jobs <-chan int) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d: stopped: %v\n", id, ctx.Err())
			return
		case job := <-jobs:
			fmt.Printf("worker %d: start 画像 %d を変換\n", id, job)
			thumbnail := convertImage(job)
			fmt.Printf("worker %d: done %s を作成\n", id, thumbnail)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan int)

	// workerを2つ起動する。2つの仕事を同時に処理できる
	var wg sync.WaitGroup
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(ctx, id, jobs)
		}(i)
	}

	fmt.Println("main: job 1 を送る")
	jobs <- 1
	fmt.Println("main: job 2 を送る")
	jobs <- 2
	time.Sleep(100 * time.Millisecond)

	fmt.Println("main: cancelする（job 1, 2は処理中）")
	cancel()
	wg.Wait()
	fmt.Println("main: 完了")
}
