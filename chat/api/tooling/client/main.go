package main

import (
	app "github.com/omer1998/chat-app-go.git/chat/api/tooling/chat"
)

func main() {
	addr := "ws://localhost:3000/connect"
	// here we retrieve the user info which is me from the config file
	config, err := app.NewConfig()
	if err != nil {
		panic(err)
	}

	// fmt.Printf("config file path: %s", config.FilePath)
	client := app.NewClient(addr, config)
	defer client.Close()
	myApp := app.NewApp(client, config)
	if err := myApp.Run(); err != nil {
		panic(err)
	}

}
