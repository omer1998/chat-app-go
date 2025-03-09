package main

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {

	nc, err := nats.Connect("demo.nats.io")
	if err != nil {
		fmt.Print("error nats connection: %s", err.Error())
		return
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		fmt.Printf("error creating jetstream: %s", err.Error())
		return
	}
	sub := "omerfaris"
	// creating stream
	s, err := js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
		Name:     sub,
		Subjects: []string{sub},
	})
	if err != nil {
		fmt.Printf("error creating stream: %s", err.Error())
		return
	}
	js.Publish(context.Background(), sub, []byte("hello from the other side"))
	//create or update consummer
	_, err = s.CreateOrUpdateConsumer(context.Background(), jetstream.ConsumerConfig{
		Name:      "omerconsumer",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		fmt.Printf("error creating consumer: %s", err.Error())
		return
	}
	cons, err := s.Consumer(context.Background(), "omerconsumer")
	if err != nil {
		fmt.Printf("error retrieving consumer: %s", err.Error())
		return
	}
	msg, err := cons.Next()
	if err != nil {
		fmt.Printf("error consume message: %s", err.Error())
		return
	}
	fmt.Printf("msg recived: %s", string(msg.Data()))

}

func consumer(id int) {

}
