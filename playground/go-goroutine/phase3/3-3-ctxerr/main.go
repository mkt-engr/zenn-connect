package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type Job struct{ ID int }

// process は1件の仕事を小さなステップに分けて進め、ステップごとにキャンセルを確認する
func process(ctx context.Context, job Job) error {
	for step := 1; step <= 3; step++ {
		select {
		case <-ctx.Done():
			return ctx.Err() // キャンセルまたは期限切れの理由を返す
		default:
		}

		time.Sleep(500 * time.Millisecond)
	}

	// job 2 は、キャンセルとは関係のない通常のエラーで失敗する想定
	if job.ID == 2 {
		return errors.New("外部APIがエラーを返した")
	}
	return nil
}

func worker(ctx context.Context, jobs <-chan Job, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker: stop requested:", ctx.Err())
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			if err := process(ctx, job); err != nil {
				if ctx.Err() != nil {
					fmt.Printf("worker: job %d を中断: %v\n", job.ID, ctx.Err())
					return // キャンセルまたは期限切れによる終了
				}

				log.Printf("worker: job %d failed: %v", job.ID, err)
				continue // 通常のJobエラーなら次のJobへ
			}
			fmt.Printf("worker: job %d done\n", job.ID)
		}
	}
}

func main() {
	// 3.5秒で期限切れになるContext（WithCancel + cancel() なら context.Canceled になる）
	ctx, cancel := context.WithTimeout(context.Background(), 3500*time.Millisecond)
	defer cancel()

	jobs := make(chan Job)
	var wg sync.WaitGroup
	wg.Add(1)
	go worker(ctx, jobs, &wg)

	for id := 1; id <= 3; id++ {
		select {
		case jobs <- Job{ID: id}:
		case <-ctx.Done():
		}
	}
	close(jobs)
	wg.Wait()

	fmt.Println("main: ctx.Err() =", ctx.Err())
}
