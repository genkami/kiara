# Kiara

![ci status](https://github.com/genkami/kiara/workflows/Test/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/genkami/kiara.svg)](https://pkg.go.dev/github.com/genkami/kiara)

![phoenix](./doc/img/phoenix.png)

Kiara is a Go equivalent of Phoenix PubSub that makes it easy for Go applications to communicate with each other.

## Basic Usage (with Redis Backend)

``` go
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
```

## Codec
By default, messages are marshaled into gob format. You can specify which format does Kiara use to marshal and unmarshal messages by passing `WithCodec()` to `NewPubSub()`.

``` go
import "github.com/genkami/kiara/codec/msgpack"

pubsub := kiara.NewPubSub(
    adapter.NewAdapter(redisClient),
    kiara.WithCodec(msgpack.Codec),
)
```

## Custom Codec
You can implement your own codec by simply implementing `Marshal` and `Unmarshal`. For example, if you want to encode messages into [WATSON](https://github.com/genkami/watson), you have to implement WATSON codec like this:

``` go
import 	"github.com/genkami/watson"

type WatsonCodec struct{}

func (_ *WatsonCodec) Marshal(v interface{}) ([]byte, error) {
	return watson.Marshal(v)
}

func (_ *WatsonCodec) Unmarshal(src []byte, v interface{}) error {
	return watson.Unmarshal(src, v)
}
```


## Examples
* [Basic Usage](./examples/basic-usage/README.md)
* [Custom Codec (WATSON × Kiara)](./examples/custom-codec/README.md)
