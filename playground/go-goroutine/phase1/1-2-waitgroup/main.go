package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("start job")
	// ここでは重い処理を想定して3秒待機しているだけで、実際の処理（API呼び出し・DB保存・画像変換など）は行っていない。
	time.Sleep(3 * time.Second)
	fmt.Println("done job")
}

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go worker(&wg)

	// workerが裏で動いている間、mainは別の仕事を続けられる
	fmt.Println("main: ジョブをworkerに任せた")
	for i := 1; i <= 2; i++ {
		time.Sleep(1 * time.Second)
		fmt.Println("main: 別の仕事をしている", i)
	}

	fmt.Println("main: workerの完了を待つ")
	wg.Wait()
	fmt.Println("main: 完了")
}
