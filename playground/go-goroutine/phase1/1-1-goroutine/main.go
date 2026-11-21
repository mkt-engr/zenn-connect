package main

import (
	"fmt"
	"time"
)

func worker() {
	fmt.Println("start job")
	// ここでは重い処理を想定して3秒待機しているだけで、実際の処理（API呼び出し・DB保存・画像変換など）は行っていない。
	time.Sleep(3 * time.Second)
	fmt.Println("done job") // mainが先に終わるため、この行は表示されない
}

func main() {
	// goroutineを起動するだけで、終了は待たない
	go worker()

	fmt.Println("main: ジョブをworkerに任せた")
	fmt.Println("main: 終了する")
	// mainが終わるとプログラム全体が終了し、動作中のworkerも途中で打ち切られる
}
