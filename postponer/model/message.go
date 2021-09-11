package model

import "time"

type Message struct {
	ID      string
	Queue   string
	Body    string
	FiresAt time.Time
}
