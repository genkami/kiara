package main

import (
	"context"
	"fmt"

	"github.com/genkami/kiara"
	adapter "github.com/genkami/kiara/adapter/redis"
	"github.com/go-redis/redis/v8"
)

type Message struct {
	From string
	Body string
}

func main() {
	var err error
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	pubsub := kiara.NewPubSub(adapter.NewAdapter(redisClient))
	defer pubsub.Close()

	channel := make(chan Message, 5)
	sub, err := pubsub.Subscribe("room:123", channel)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	ctx := context.Background()
	msg := &Message{From: "birb", Body: "cock-a-doodle-doo"}
	err = pubsub.Publish(ctx, "room:123", msg)
	if err != nil {
		panic(err)
	}

	sent := <-channel
	fmt.Printf("%s: %s\n", sent.From, sent.Body)
}
