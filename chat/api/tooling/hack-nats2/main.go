package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	var wg sync.WaitGroup
	subject := "msg"
	nc, err := nats.Connect(nats.DefaultURL)

	if err != nil {
		log.Fatalf("error connection %s", err.Error())
	}
	defer nc.Close()
	for i := 1; i <= 5; i++ {
		message := []byte("Hello NATS " + strconv.Itoa(i))
		err = nc.Publish(subject, message)
		if err != nil {
			log.Fatalf("error publishing message: %s", err.Error())
		}
		log.Printf("Published message %d: %s\n", i, message)
		time.Sleep(1 * time.Second)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		nc.Subscribe(subject, func(msg *nats.Msg) {
			fmt.Printf("recieved msg: %s", string(msg.Data))
		})

	}()

	wg.Wait()

}
