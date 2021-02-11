# Kiara

![phoenix](./doc/img/phoenix.png)

Kiara is a Go equivalent of Phoenix PubSub that makes it easy for Go applications to communicate with each other.

# (WIP) Basic Usage (with Redis Backend) (WIP)

``` go
import (
    "fmt"
    "context"

    "github.com/go-redis/redis/v8"
    "github.com/genkami/kiara"
    "github.com/genkami/kiara/adapter/redis"
)

type Message struct {
    From string
    Body string
}

func main() {
    var err error
    redis := redis.NewClient(&redis.Options{Addr: "locslhost:6379"})
    pubsub := kiara.NewPubSub(redis.NewAdapter(redis))
    defer pubsub.Close()

    channel := make(chan Message, 5)
    sub, err = pubsub.Subscribe("room:123", channel)
    if err != nil {
        panic(err)
    }

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
