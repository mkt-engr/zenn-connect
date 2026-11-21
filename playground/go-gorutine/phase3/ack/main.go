package main

import (
	"errors"
	"fmt"
)

// Message は ACK/NACK を持つメッセージの概念的なモデル
type Message struct {
	ID    int
	acked bool
}

func (m *Message) Ack()  { fmt.Printf("message %d: ACK\n", m.ID) }
func (m *Message) Nack() { fmt.Printf("message %d: NACK (再配信される)\n", m.ID) }

func process(m *Message) error {
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
