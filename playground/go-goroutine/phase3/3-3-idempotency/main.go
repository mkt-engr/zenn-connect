package main

import (
	"fmt"
	"time"
)

// Message は注文イベントの概念的なモデル
type Message struct {
	OrderID int
}

// charge は決済API呼び出しを想定した処理
func charge(orderID int) {
	// ここでは重い処理を想定して3秒待機しているだけで、実際の処理（API呼び出し・DB保存・画像変換など）は行っていない。
	time.Sleep(3 * time.Second)
	fmt.Printf("  注文 %d: 決済した\n", orderID)
}

// 冪等でない例：届くたびに毎回決済する
func handleNaive(m Message) {
	fmt.Printf("注文 %d を受信\n", m.OrderID)
	charge(m.OrderID)
}

// 冪等な例：処理済みの注文IDは何もしない
// processed は、注文IDの一意制約を持つordersテーブルの代わり
var processed = map[int]bool{}

func handleIdempotent(m Message) {
	fmt.Printf("注文 %d を受信\n", m.OrderID)
	if processed[m.OrderID] {
		fmt.Printf("  注文 %d: 処理済みなのでスキップ\n", m.OrderID)
		return
	}
	charge(m.OrderID)
	processed[m.OrderID] = true
}

func main() {
	// 注文1のACKがブローカーに届かず、注文1が再配信された状況を想定する
	deliveries := []Message{{OrderID: 1}, {OrderID: 2}, {OrderID: 1}}

	fmt.Println("=== 冪等でない場合 ===")
	for _, m := range deliveries {
		handleNaive(m)
	}

	fmt.Println("=== 冪等な場合 ===")
	for _, m := range deliveries {
		handleIdempotent(m)
	}
}
