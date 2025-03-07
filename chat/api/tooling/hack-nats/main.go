package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	subject := "chat-msg"
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Printf("error creating connection: %s", err.Error())
		os.Exit(1)
	}
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}
	js.AddStream(&nats.StreamConfig{
		Name:     subject,
		Subjects: []string{subject},
	})
	defer nc.Close()
	var wg sync.WaitGroup

	wg.Add(1)
	go subscriber(subject, &wg)

	wg.Add(1)
	go subscriber(subject, &wg)

	if err = publisher("hello from the other side", subject); err != nil {
		fmt.Printf("error pub: %s", err.Error())
	}

	wg.Wait()

}

func publisher(msg string, subject string) error {
	nc, _ := nats.Connect(nats.DefaultURL)

	err := nc.Publish(subject, []byte(msg))
	if err != nil {
		return err
	}
	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	_, err = js.Publish(subject, []byte("Hello, JetStream!"))
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil

}
func subscriber(subject string, wg *sync.WaitGroup) error {
	defer wg.Done()
	nc, _ := nats.Connect(nats.DefaultURL)
	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	sub, err := js.SubscribeSync(subject)
	if err != nil {
		return err
	}
	msg, err := sub.NextMsg(time.Second * 5)
	if err != nil {
		return err
	}

	log.Printf("Received message: %s", string(msg.Data))
	return nil

}
