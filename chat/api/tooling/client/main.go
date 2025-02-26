package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/omer1998/chat-app-go.git/chat/app/domain/chatapp"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/errs"
)

func main() {

	addr := "ws://localhost:3000/connect"
	dialer := websocket.DefaultDialer
	clientConn, _, err := dialer.Dial(addr, nil)
	if err != nil {
		fmt.Printf("error connecting to webserver %s", err.Error())
		return
	}
	defer clientConn.Close()
	err = hack2(clientConn)
	if err != nil {
		log.Fatal(err)
	}

}

func hack2(conn *websocket.Conn) error {

	_, data, err := conn.ReadMessage()
	if err != nil {
		return errs.Newf(errs.Internal, "error reading msg from connection %s: ", err.Error())
	}
	if string(data) != "HELLO" {
		return fmt.Errorf("error unexpected handshake message")
	}

	user := chatapp.User{
		Id:   uuid.New(),
		Name: "Omer Faris",
	}
	// we need to marshal this struct (object) to byte
	data, err = json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error marshling user struct: %s", err.Error())
	}

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
