package chat

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/omer1998/chat-app-go.git/chat/app/domain/chatapp"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/chat"
)

type Client struct {
	url  string
	conn *websocket.Conn
	user chatapp.User
}

func NewClient(url string, user chatapp.User) *Client {
	return &Client{url: url, user: user}
}
func (c *Client) Handshake(msgHnadler messageHandler) error {

	dialer := websocket.DefaultDialer
	clientConn, _, err := dialer.Dial(c.url, nil)
	if err != nil {
		return fmt.Errorf("error connecting to webserver %s", err.Error())
	}
	c.conn = clientConn
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("error reading msg from connection %s: ", err.Error())
	}
	if string(data) != "HELLO" {
		return fmt.Errorf("error unexpected handshake message")
	}
	// here the user send his data
	// we need to marshal this struct (object) to byte
	data, err = json.Marshal(c.user)
	if err != nil {
		return fmt.Errorf("error marshling user struct: %s", err.Error())
	}

	// send this data to the server

	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("error writing user data: %s", err.Error())
	}

	// read another messhage from server
	_, data, err = c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("error reading msg from connection %s: ", err.Error())
	}
	msgHnadler("system", string(data))

	// fmt.Println("message from server: ", string(data))
	// msgsChan := make(chan string)
	// msgHnadler("--> connected")
	c.ReadIncMessages(msgHnadler)

	return nil
}
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Send(msg chat.InMessage) error {
	if msg.Msg == "" {
		return fmt.Errorf("empty msg not allowed")
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal err: %w", err)
	}
	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("send error: %w", err)
	}

	return nil
}

func (c *Client) ReadIncMessages(msgHandler messageHandler) {
	msgsChan := make(chan chat.OutMessage)
	go func() {
		for {
			_, data, err := c.conn.ReadMessage()
			if err != nil {

				fmt.Printf("error reading msg: %s ", err.Error())
				return
			}
			var outMessage chat.OutMessage
			if err := json.Unmarshal(data, &outMessage); err != nil {
				fmt.Printf("error unmarshling msg: %s ", err.Error())
				return
			}
			msgsChan <- outMessage
			// fmt.Println("\nmessage from server: ", outMessage.Msg)

		}
	}()
	for {
		msg := <-msgsChan
		msgHandler(msg.From.Name, msg.Msg)
	}
}
