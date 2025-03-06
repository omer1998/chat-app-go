package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/omer1998/chat-app-go.git/chat/app/domain/chatapp"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/chat"
)

func main() {

	// TODO: handle the shutdown of all goroutine by using context
	// 1:  create context
	// 2: create go routing listening for inturruption signal;
	// 	once catch signal --> cnacel context
	// 	close connection
	// 3 pass conetext to the hack func:
	// 4: make all go routine to check/listen for context.Done() as a case
	//  and another case would be the main function of the goroutine
	// once context is done --> return
	//
	addr := "ws://localhost:3000/connect"
	dialer := websocket.DefaultDialer
	clientConn, _, err := dialer.Dial(addr, nil)
	if err != nil {
		fmt.Printf("error connecting to webserver %s", err.Error())
		return
	}
	defer clientConn.Close()

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
	var toUserId uuid.UUID
	switch id {
	case 0:
		userId, _ = uuid.Parse(usersId[0])
		toUserId, _ = uuid.Parse(usersId[1])
		name = names[0]
	case 1:
		userId, _ = uuid.Parse(usersId[1])
		toUserId, _ = uuid.Parse(usersId[0])
		name = names[1]

	}
	meUser := chatapp.User{Id: userId, Name: name}
	toUser := chatapp.User{Id: toUserId, Name: name}

	if err := hack2(clientConn, meUser, toUser); err != nil {
		fmt.Println("error hack2 %w", err)
		return
	}

}

func hack2(conn *websocket.Conn, meUser chatapp.User, toUser chatapp.User) error {
	fmt.Println("start hack2")

	_, data, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("error reading msg from connection %s: ", err.Error())
	}
	if string(data) != "HELLO" {
		return fmt.Errorf("error unexpected handshake message")
	}

	// we need to marshal this struct (object) to byte
	data, err = json.Marshal(meUser)
	if err != nil {
		return fmt.Errorf("error marshling user struct: %s", err.Error())
	}

	fmt.Println("user from clien %w", meUser)
	// send this data to the server

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("error writing user data: %s", err.Error())
	}

	// read another messhage from server
	_, data, err = conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("error reading msg from connection %s: ", err.Error())
	}

	fmt.Println("message from server: ", string(data))

	go func() {
		for {
			defer conn.Close()
			_, data, err := conn.ReadMessage()
			if err != nil {

				fmt.Printf("error reading msg: %s ", err.Error())
				return
			}
			var outMessage chat.OutMessage
			if err := json.Unmarshal(data, &outMessage); err != nil {
				fmt.Printf("error unmarshling msg: %s ", err.Error())
				return
			}
			fmt.Println("\nmessage from server: ", outMessage.Msg)
			// fmt.Println("msg type: ", msgType)

		}
	}()

	for {
		fmt.Printf("message> ")

		reader := bufio.NewReader(os.Stdin)
		line, _, _ := reader.ReadLine()
		msg := chat.InMessage{

			ToId: toUser.Id,
			Msg:  string(line),
		}
		err := conn.WriteJSON(msg)
		if err != nil {
			conn.Close()

			return fmt.Errorf("error writing message to %s, err : %s", toUser.Id.String(), err.Error())

		}

	}

	// here we want to send msg to specific user
	return nil
}
func hack1(clientConn *websocket.Conn) {

	err := clientConn.WriteMessage(websocket.TextMessage, []byte("hello from client ..."))
	if err != nil {
		fmt.Printf("error writing to server %s", err.Error())
	}
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	doneChan := make(chan struct{})
	go readMessage(clientConn, doneChan)
	for {
		select {
		case <-doneChan:
			fmt.Println("done")
			return
		case tick := <-ticker.C:
			err = clientConn.WriteMessage(websocket.TextMessage, []byte(tick.GoString()))
			if err != nil {
				fmt.Printf("error writing to server %s", err.Error())
				return
			}
		case <-shutdownGracefully():
			fmt.Println("interrupt signal..!!!")
			clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing now because of system signal"))
			clientConn.Close()
			return
		}

	}
}

func readMessage(con *websocket.Conn, doneChan chan struct{}) {
	defer close(doneChan)

	for {
		_, data, err := con.ReadMessage()
		if err != nil {
			fmt.Printf("error reading message %s", err.Error())
			return
		}
		fmt.Printf("data from server is %s \n", string(data))
	}
}

func shutdownGracefully() chan os.Signal {
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT)
	return shutdownChan
}
