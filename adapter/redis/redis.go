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
	ErrAlreadySubscribed = errors.New("already subscribed")
	ErrSlowConsumer      = errors.New("slow consumer")
)

// RedisClient is an abstract interface for Redis client.
type RedisClient interface {
	Publish(context.Context, string, interface{}) *redis.IntCmd
	Subscribe(context.Context, ...string) *redis.PubSub
	Close() error
}

// Adapter is an adapter that sends message through Redis PubSub
type Adapter struct {
	client      RedisClient
	pubSub      *redis.PubSub
	publishCh   chan *types.Message
	deliveredCh chan *types.Message
	errorCh     chan error
	done        chan struct{}
	doneWg      sync.WaitGroup
	opts        options
}

// NewAdapter returns a new Adapter.
func NewAdapter(client RedisClient, options ...Option) *Adapter {
	opts := defaultOptions()
	for _, o := range options {
		o.apply(&opts)
	}
	a := &Adapter{
		client: client,
		// NOTE: client.Subscribe() does not block when channels is not given.
		pubSub:      client.Subscribe(context.Background()),
		publishCh:   make(chan *types.Message, opts.publishChSize),
		deliveredCh: make(chan *types.Message, opts.deliveredChSize),
		errorCh:     make(chan error, opts.errorChSize),
		done:        make(chan struct{}),
		opts:        opts,
	}
	a.doneWg.Add(1)
	go a.run()
	return a
}

func (a *Adapter) run() {
	defer a.doneWg.Done()
	msgCh := a.pubSub.Channel()
	for {
		select {
		case <-a.done:
			return
		case msg := <-a.publishCh:
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), a.opts.publishTimeout)
				defer cancel()
				err := a.client.Publish(ctx, msg.Topic, string(msg.Payload)).Err()
				if err != nil {
					select {
					case a.errorCh <- err:
					default:
						// discard
					}
				}
			}()
		case m := <-msgCh:
			msg := &types.Message{Topic: m.Channel, Payload: []byte(m.Payload)}
			select {
			case a.deliveredCh <- msg:
			default:
				select {
				case a.errorCh <- ErrSlowConsumer:
				default:
					// discard
				}
			}
		}
	}
}

func (a *Adapter) Publish() chan<- *types.Message {
	return a.publishCh
}

func (a *Adapter) Delivered() <-chan *types.Message {
	return a.deliveredCh
}

func (a *Adapter) Errors() <-chan error {
	return a.errorCh
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

func (a *Adapter) Close() {
	close(a.done)
	a.doneWg.Wait()
	a.pubSub.Close()
	a.client.Close()
}

var _ types.Adapter = &Adapter{}
