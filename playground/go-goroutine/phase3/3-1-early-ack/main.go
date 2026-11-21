package main

import (
	"fmt"
	"time"
)

// Message は ACK/NACK を持つメッセージの概念的なモデル
type Message struct {
	ID int
}

func (m *Message) Ack() {
	fmt.Printf("message %d: ACK（ブローカーは完了扱いにする）\n", m.ID)
}

func process(m *Message) {
	fmt.Printf("message %d: start processing\n", m.ID)
	// ここでは重い処理を想定して3秒待機しているだけで、実際の処理（API呼び出し・DB保存・画像変換など）は行っていない。
	time.Sleep(3 * time.Second)
	fmt.Printf("message %d: processed\n", m.ID) // プロセスが先に止まるため、この行は表示されない
}

// 悪い例：処理をgoroutineに渡した直後にACKしてしまう
func handleMessage(m *Message) {
	go process(m)
	m.Ack() // 処理が終わる前にACKしている
}

func main() {
	handleMessage(&Message{ID: 1})

	// 処理中にプロセスが停止した状況（デプロイ・障害・クラッシュなど）を想定する
	time.Sleep(1 * time.Second)
	fmt.Println("main: プロセスが停止する")
	// この時点でmessage 1はACK済みなので再配信されず、処理は失われる
}
