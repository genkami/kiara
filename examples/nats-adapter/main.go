package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/genkami/kiara"
	adapter "github.com/genkami/kiara/adapter/nats"
	"github.com/nats-io/nats.go"
)

type Message struct {
	From string
	Body string
}

func main() {
	var natsUrl string
	flag.StringVar(&natsUrl, "nats-url", "nats://localhost:4222", "URL to NATS server")
	flag.Parse()

	var err error
	conn, err := nats.Connect(natsUrl)
	if err != nil {
		panic(err)
	}
	pubsub := kiara.NewPubSub(adapter.NewAdapter(conn))
	defer pubsub.Close()

	channel := make(chan Message, 5)
	sub, err := pubsub.Subscribe("room:123", channel)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	ctx := context.Background()
	msg := &Message{From: "Risu", Body: "Hello, NATS!"}
	err = pubsub.Publish(ctx, "room:123", msg)
	if err != nil {
		panic(err)
	}

	sent := <-channel
	fmt.Printf("%s: %s\n", sent.From, sent.Body)
}
