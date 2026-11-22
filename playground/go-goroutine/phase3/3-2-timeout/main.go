package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

const steps = 8 // 1件の仕事は8ステップ（1ステップ1秒）かかる想定

// process は長い仕事を1秒ずつのステップに分けて進める。
// cooperate が true のときだけ、ステップごとにctxを確認してキャンセルに協力する
func process(ctx context.Context, id, job int, cooperate bool) error {
	for step := 1; step <= steps; step++ {
		if cooperate {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		time.Sleep(1 * time.Second)
		fmt.Printf("worker %d: job %d step %d/%d\n", id, job, step, steps)
	}
	return nil
}

func worker(workCtx context.Context, id int, jobs <-chan int, wg *sync.WaitGroup, cooperate bool) {
	defer wg.Done()

	for {
		select {
		case <-workCtx.Done():
			fmt.Printf("worker %d: stop requested\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			fmt.Printf("worker %d: start job %d\n", id, job)
			if err := process(workCtx, id, job, cooperate); err != nil {
				if workCtx.Err() != nil {
					// キャンセルまたは期限切れによる終了
					fmt.Printf("worker %d: job %d を中断 (%v)\n", id, job, workCtx.Err())
					return
				}
				// 通常のJobエラーならWorker自体は終了せず次のJobへ
				fmt.Printf("worker %d: job %d failed: %v\n", id, job, err)
				continue
			}
			fmt.Printf("worker %d: done job %d\n", id, job)
		}
	}
}

func main() {
	// 使い方: go run ./phase3/3-2-timeout ignore    … 処理がctxを確認しない
	//         go run ./phase3/3-2-timeout cooperate … 処理がctxを確認して中断する
	mode := "ignore"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	cooperate := mode == "cooperate"
	fmt.Println("mode:", mode)

	// 受付を止める合図（実際のサーバーではSIGTERMを受けたときに相当する）
	stopCtx, stop := context.WithCancel(context.Background())
	defer stop()
	// Workerや仕事にキャンセルを伝えるContext（shutdownCtxとは別に用意する）
	workCtx, workCancel := context.WithCancel(context.Background())
	defer workCancel()

	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go worker(workCtx, i, jobs, &wg, cooperate)
	}

	go func() {
		defer close(jobs)
		for job := 1; ; job++ {
			select {
			case <-stopCtx.Done():
				fmt.Println("main: 新しい仕事の受付を止めた")
				return
			case jobs <- job:
				fmt.Printf("main: job %d を渡した\n", job)
			}
		}
	}()

	// 動作確認しやすいよう、1秒後に自動でシャットダウンを開始する
	time.AfterFunc(1*time.Second, stop)

	<-stopCtx.Done()
	fmt.Println("main: shutdown requested（処理中の仕事の完了を待つ）")

	// 終了期限は2秒。仕事は8秒かかるので、必ずタイムアウトする
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("main: all workers stopped")
		return
	case <-shutdownCtx.Done():
		fmt.Println("main: shutdown timeout:", shutdownCtx.Err())
	}

	// タイムアウトしても、Workerは自動では止まらない。workCtxをキャンセルして終了を伝える
	fmt.Println("main: workCancel() を呼ぶ")
	workCancel()

	// キャンセル後の様子を3秒だけ見る
	select {
	case <-done:
		fmt.Println("main: all workers stopped")
	case <-time.After(3 * time.Second):
		fmt.Println("main: 3秒待ってもworkerが止まらない。mainが終わるとプロセスごと終了する")
	}
}
