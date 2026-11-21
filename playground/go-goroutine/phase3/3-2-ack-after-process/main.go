package main

import (
	"errors"
	"fmt"
	"time"
)

// Message は ACK/NACK を持つメッセージの概念的なモデル
type Message struct {
	ID int
}

func (m *Message) Ack()  { fmt.Printf("message %d: ACK（完了扱い）\n", m.ID) }
func (m *Message) Nack() { fmt.Printf("message %d: NACK（再配信される）\n", m.ID) }

func process(m *Message) error {
	fmt.Printf("message %d: start processing\n", m.ID)
	// ここでは重い処理を想定して3秒待機しているだけで、実際の処理（API呼び出し・DB保存・画像変換など）は行っていない。
	time.Sleep(3 * time.Second)

	// 動作確認しやすいよう、IDが偶数のメッセージは失敗させる
	if m.ID%2 == 0 {
		return errors.New("process failed")
	}
	fmt.Printf("message %d: processed\n", m.ID)
	return nil
}

// 処理が成功してからACKする
func handleMessage(m *Message) {
	if err := process(m); err != nil {
		m.Nack()
		return
	}
	m.Ack()
}

func main() {
	for id := 1; id <= 4; id++ {
		handleMessage(&Message{ID: id})
	}
}
