package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type User struct {
	Id       string
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
	ToId string `json:"toId"`
	// From User   `json:"from"`
	Msg string `json:"msg"`
}
type InMessageBus struct {
	CapId    uuid.UUID `json:"capId"`
	FromId   string    `json:"fromId"`
	FromName string    `json:"fromName"`
	ToId     string    `json:"toiD"`
	Msg      string    `json:"msg"`
}
type OutMessage struct {
	From User   `json:"to"`
	Msg  string `json:"msg"`
}
