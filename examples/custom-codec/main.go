package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/genkami/kiara"
	adapter "github.com/genkami/kiara/adapter/redis"
	"github.com/genkami/watson"
	"github.com/go-redis/redis/v8"
)

type Message struct {
	From string
	Body string
}

type WatsonCodec struct{}

func (_ *WatsonCodec) Marshal(v interface{}) ([]byte, error) {
	return watson.Marshal(v)
}

func (_ *WatsonCodec) Unmarshal(src []byte, v interface{}) error {
	return watson.Unmarshal(src, v)
}

func main() {
	var redisAddr string
	flag.StringVar(&redisAddr, "redis-addr", "localhost:6379", "Redis address")
	flag.Parse()

	var err error
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	pubsub := kiara.NewPubSub(
		adapter.NewAdapter(redisClient),
		kiara.WithCodec(&WatsonCodec{}),
	)
	defer pubsub.Close()

	channel := make(chan Message, 5)
	sub, err := pubsub.Subscribe("room:123", channel)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	ctx := context.Background()
	msg := &Message{From: "Amelia", Body: "Do you know that this message was marshaled into WATSON?"}
	err = pubsub.Publish(ctx, "room:123", msg)
	if err != nil {
		panic(err)
	}

	sent := <-channel
	fmt.Printf("%s: %s\n", sent.From, sent.Body)
}
