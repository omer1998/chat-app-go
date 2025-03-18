package app

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/chat"
)

type Client struct {
	url    string
	conn   *websocket.Conn
	config *Config
}

func NewClient(url string, config *Config) *Client {
	return &Client{url: url, config: config}
}
func (c *Client) Handshake(uiWriteMsg uiMsgHandler, uiUpdateUser uiUpdateUsers) error {

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
	data, err = json.Marshal(c.config.User)
	if err != nil {
		return fmt.Errorf("error marshling user struct: %s", err.Error())
	}

	// send this data to the server

	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("error writing user data: %s", err.Error())
	}

	// read another messhage from server
	_, _, err = c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("error reading msg from connection %s: ", err.Error())
	}
	// uiWriteMsg("system", string(data))

	// fmt.Println("message from server: ", string(data))
	// msgsChan := make(chan string)
	// msgHnadler("--> connected")
	c.ReadIncMessages(uiWriteMsg, uiUpdateUser)

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

func (c *Client) ReadIncMessages(uiWriteMsg uiMsgHandler, uiUpdateUsers uiUpdateUsers) {
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
			// fmt.Println("recieved msg >>>>> ", outMessage.Msg)
			// here we need to look if the from user is present in our contact or not
			// if present push message in channal directly
			// if not present we need to do two things :
			// 1/ update our contact list in config json file
			// 2/ update the ui to show the new user

			usr, err := c.config.LookUpUser(outMessage.From.Id)
			if err != nil {
				if errors.Is(err, errUserNotFound) {
					// here we add this user to our contact
					usr := user{Id: outMessage.From.Id, Name: outMessage.From.Name}
					err := c.config.UpdateContact(usr)
					// here we need to update the ui (terminal user interface)
					if err == nil {
						uiUpdateUsers(user{Id: outMessage.From.Id, Name: outMessage.From.Name})

						// we also need to add this incoming message to the messages that relate to this user
						// the whole idea here; we need to save incoming messages from this user in messages field of this user
						// in this function also we persist this msg in specific file related to this user
						// if err := c.config.AddMessage(
						// 	outMessage.From.Id, fmt.Sprintf("\n%s : %s", outMessage.From.Name, outMessage.Msg)); err != nil {
						// 	fmt.Printf("error adding msg to user: %s", err.Error())
						// 	return
						// }
						if err := c.config.AddMessage(usr.Id, fmt.Sprintf("%s: %s\n", usr.Name, outMessage.Msg)); err != nil {
							uiWriteMsg("system", err.Error())

						}
						continue

					} else {
						uiWriteMsg("system", err.Error())

					}

				}
			} // no error mean the user is found
			// add message to this user
			if err := c.config.AddMessage(usr.Id, fmt.Sprintf("\n%s: %s", usr.Name, outMessage.Msg)); err != nil {
				uiWriteMsg("system", err.Error())

			}
			msgsChan <- outMessage

			// fmt.Println("\nmessage from server: ", outMessage.Msg)

		}
	}()
	for {
		msg := <-msgsChan
		uiWriteMsg(msg.From.Name, msg.Msg)
	}
}
