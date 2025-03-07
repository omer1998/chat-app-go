package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type User struct {
	Id       uuid.UUID
	Name     string
	Conn     *websocket.Conn
	LastPing time.Time
	LastPong time.Time
}

type Connection struct {
	LastPong time.Time
	LastPing time.Time
	Conn     *websocket.Conn
}

type InMessage struct {
	ToId uuid.UUID `json:"toId"`
	// From User   `json:"from"`
	Msg string `json:"msg"`
}
type InMessageBus struct {
	From User      `json:"from"`
	ToId uuid.UUID `json:"toiD"`
	Msg  string    `json:"msg"`
}
type OutMessage struct {
	From User   `json:"to"`
	Msg  string `json:"msg"`
}
