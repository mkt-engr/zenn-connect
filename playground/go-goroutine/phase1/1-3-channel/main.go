package main

import (
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

func worker(jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fmt.Printf("start: 画像 %d を変換\n", job)
		thumbnail := convertImage(job)
		fmt.Printf("done: %s を作成\n", thumbnail)
	}
}

func main() {
	jobs := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)
	go worker(jobs, &wg)

	// jobsはバッファなしなので、workerが受け取るまで送信側のmainは待たされる
	fmt.Println("main: job 1 を送る")
	jobs <- 1
	fmt.Println("main: job 1 を送り終えた")

	fmt.Println("main: job 2 を送る")
	jobs <- 2 // workerがjob 1を処理中のため、ここで約3秒待たされる
	fmt.Println("main: job 2 を送り終えた")

	close(jobs)
	fmt.Println("main: workerの完了を待つ")
	wg.Wait()
	fmt.Println("main: 完了")
}
