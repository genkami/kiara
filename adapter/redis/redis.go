// Package redis provides a Redis adapter for Kiara.
package redis

import (
	"context"
	"errors"
	"sync"

	"github.com/go-redis/redis/v8"

	"github.com/genkami/kiara/types"
)

var (
	// This error is reported via types.Pipe.Errors when the adapter can't deliver
	// succeeding messages arrived from Redis because types.Pipe.Delivered is already full.
	ErrSlowConsumer = errors.New("slow consumer")
)

// RedisClient is an abstract interface for Redis client.
type RedisClient interface {
	Publish(context.Context, string, interface{}) *redis.IntCmd
	Subscribe(context.Context, ...string) *redis.PubSub
	Close() error
}

// Adapter is an adapter that sends message through Redis PubSub
type Adapter struct {
	client RedisClient
	pubSub *redis.PubSub
	pipe   *types.Pipe
	done   chan struct{}
	doneWg sync.WaitGroup
	opts   options
}

var _ types.Adapter = &Adapter{}

// NewAdapter returns a new Adapter.
func NewAdapter(client RedisClient, options ...Option) *Adapter {
	opts := defaultOptions()
	for _, o := range options {
		o.apply(&opts)
	}
	a := &Adapter{
		client: client,
		// NOTE: client.Subscribe() does not block when channels is not given.
		pubSub: client.Subscribe(context.Background()),
		done:   make(chan struct{}),
		opts:   opts,
	}
	return a
}

func (a *Adapter) Start(pipe *types.Pipe) {
	a.pipe = pipe
	a.doneWg.Add(1)
	go a.run()
}

func (a *Adapter) run() {
	defer a.doneWg.Done()
	msgCh := a.pubSub.Channel()
	for {
		select {
		case <-a.done:
			return
		case msg := <-a.pipe.Publish:
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), a.opts.publishTimeout)
				defer cancel()
				err := a.client.Publish(ctx, msg.Topic, string(msg.Payload)).Err()
				if err != nil {
					select {
					case a.pipe.Errors <- err:
					default:
						// discard
					}
				}
			}()
		case m := <-msgCh:
			msg := &types.Message{Topic: m.Channel, Payload: []byte(m.Payload)}
			select {
			case a.pipe.Delivered <- msg:
			default:
				select {
				case a.pipe.Errors <- ErrSlowConsumer:
				default:
					// discard
				}
			}
		}
	}
}

func (a *Adapter) Subscribe(topic string) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.opts.subscriptionTimeout)
	defer cancel()
	return a.pubSub.Subscribe(ctx, topic)
}

func (a *Adapter) Unsubscribe(topic string) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.opts.subscriptionTimeout)
	defer cancel()
	return a.pubSub.Unsubscribe(ctx, topic)
}

func (a *Adapter) Stop() {
	close(a.done)
	a.doneWg.Wait()
	a.pubSub.Close()
	a.client.Close()
}
