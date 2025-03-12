package main

import (
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/omer1998/chat-app-go.git/chat/api/tooling/chat"
	"github.com/omer1998/chat-app-go.git/chat/app/domain/chatapp"
)

func main() {
	addr := "ws://localhost:3000/connect"
	id, _ := strconv.Atoi(os.Args[1])
	usersId := []string{
		"1b0deec0-5c9d-42bd-98ee-59e39d5a8105",
		"6913f142-ffe2-49c7-a222-5a1b1638992b",
	}
	names := []string{
		"omer faris",
		"ahmed faris",
	}
	var name string
	var userId uuid.UUID
	// var toUserId uuid.UUID
	switch id {
	case 0:
		userId, _ = uuid.Parse(usersId[0])
		name = names[0]
	case 1:
		userId, _ = uuid.Parse(usersId[1])
		name = names[1]

	}
	meUser := chatapp.User{Id: userId, Name: name}
	client := chat.NewClient(addr, meUser)
	defer client.Close()
	myApp := chat.NewApp(client)
	if err := myApp.Run(); err != nil {
		panic(err)
	}
}
