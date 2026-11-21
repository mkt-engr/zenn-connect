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

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fmt.Printf("worker %d: start 画像 %d を変換\n", id, job)
		thumbnail := convertImage(job)
		fmt.Printf("worker %d: done %s を作成\n", id, thumbnail)
	}
}

func main() {
	jobs := make(chan int)
	var wg sync.WaitGroup

	// workerを3つだけ起動する。仕事が10件あっても、同時に処理されるのは最大3件
	workerCount := 3

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	for job := 1; job <= 10; job++ {
		// jobsはバッファなしなので、3つのworkerがすべて処理中の間、mainはここで待たされる
		fmt.Printf("main: job %d を送る\n", job)
		jobs <- job
	}

	close(jobs)
	fmt.Println("main: workerの完了を待つ")
	wg.Wait()
	fmt.Println("main: 完了")
}
