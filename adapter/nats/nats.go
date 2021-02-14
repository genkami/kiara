// Package nats provides a NATS adapter for Kiara.
package nats

import (
	"errors"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/genkami/kiara/types"
)

// This is the length of channels to receive messages arrived from NATS.
// There is no use changing this number because messages received via such channels
// are immediately sent to another channels.
// You can configure the length of this "another channels" with DeliveredChannelSize().
const receivedNatsMsgChSize = 10

var (
	// This error is reported via Adapter.Errors() when the adapter can't deliver
	// succeeding messages arrived from NATS because Adapter.Delivered() is already full.
	ErrSlowConsumer = errors.New("slow consumer")

	// This error is returned by Adapter.Subscribe() when the topic is already subscribed.
	ErrAlreadySubscribed = errors.New("already subscribed")
)

// Adapter is an adapter that sends messages through NATS.
type Adapter struct {
	conn              *nats.Conn
	receivedNatsMsgCh chan *nats.Msg
	pipe              *types.Pipe

	done   chan struct{}
	doneWg sync.WaitGroup
	opts   options

	subsLock sync.Mutex
	subs     map[string]*nats.Subscription
}

var _ types.Adapter = &Adapter{}

// NewAdapter creates a new Adapter.
func NewAdapter(conn *nats.Conn, options ...Option) *Adapter {
	opts := defaultOptions()
	for _, o := range options {
		o.apply(&opts)
	}
	a := &Adapter{
		conn:              conn,
		receivedNatsMsgCh: make(chan *nats.Msg, receivedNatsMsgChSize),
		done:              make(chan struct{}),
		opts:              opts,
		subs:              map[string]*nats.Subscription{},
	}
	return a
}

func (a *Adapter) Start(pipe *types.Pipe) {
	a.pipe = pipe
	a.conn.SetErrorHandler(a.natsErrorHandler())
	a.conn.SetDisconnectErrHandler(a.natsConnErrorHandler())
	a.doneWg.Add(1)
	go a.run()
}

func (a *Adapter) run() {
	defer a.doneWg.Done()
	ticker := time.NewTicker(a.opts.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.done:
			return
		case msg := <-a.pipe.Publish:
			err := a.conn.Publish(msg.Topic, msg.Payload)
			if err != nil {
				select {
				case a.pipe.Errors <- err:
				default:
					// discard
				}
			}
		case natsMsg := <-a.receivedNatsMsgCh:
			msg := &types.Message{Topic: natsMsg.Subject, Payload: natsMsg.Data}
			select {
			case a.pipe.Delivered <- msg:
			default:
				select {
				case a.pipe.Errors <- ErrSlowConsumer:
				default:
					// discard
				}
			}
		case <-ticker.C:
			err := a.conn.Flush()
			if err != nil {
				select {
				case a.pipe.Errors <- err:
				default:
					// discard
				}
			}
		}
	}
}

func (a *Adapter) Subscribe(topic string) error {
	a.subsLock.Lock()
	defer a.subsLock.Unlock()
	if _, ok := a.subs[topic]; ok {
		return ErrAlreadySubscribed
	}

	sub, err := a.conn.ChanSubscribe(topic, a.receivedNatsMsgCh)
	if err != nil {
		return err
	}
	a.subs[topic] = sub
	err = a.conn.Flush()
	if err != nil {
		select {
		case a.pipe.Errors <- err:
		default:
			// discard
		}
	}
	return nil
}

func (a *Adapter) Unsubscribe(topic string) error {
	a.subsLock.Lock()
	defer a.subsLock.Unlock()
	if sub, ok := a.subs[topic]; ok {
		err := sub.Unsubscribe()
		if err != nil {
			return err
		}
		delete(a.subs, topic)
	}
	return nil
}

func (a *Adapter) Stop() {
	close(a.done)
	a.doneWg.Wait()
	a.conn.Close()
}

func (a *Adapter) natsErrorHandler() nats.ErrHandler {
	return func(_ *nats.Conn, _ *nats.Subscription, err error) {
		select {
		case a.pipe.Errors <- err:
		default:
			// discard
		}
	}
}

func (a *Adapter) natsConnErrorHandler() nats.ConnErrHandler {
	return func(_ *nats.Conn, err error) {
		select {
		case a.pipe.Errors <- err:
		default:
			// discard
		}
	}
}
