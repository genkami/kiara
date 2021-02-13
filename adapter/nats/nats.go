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

	publishCh   chan *types.Message
	deliveredCh chan *types.Message
	errorCh     chan error

	done   chan struct{}
	doneWg sync.WaitGroup
	opts   options

	subs *subscriptions
}

var _ types.Adapter = &Adapter{}

// NewAdapter creates a new Adapter.
func NewAdapter(conn *nats.Conn) *Adapter {
	opts := defaultOptions()
	// TODO: configure options
	a := &Adapter{
		conn:              conn,
		receivedNatsMsgCh: make(chan *nats.Msg, receivedNatsMsgChSize),
		publishCh:         make(chan *types.Message, opts.publishChSize),
		deliveredCh:       make(chan *types.Message, opts.deliveredChSize),
		errorCh:           make(chan error, opts.errorChSize),
		done:              make(chan struct{}),
		opts:              opts,
		subs:              newSubscriptions(),
	}
	conn.SetErrorHandler(a.natsErrorHandler())
	conn.SetDisconnectErrHandler(a.natsConnErrorHandler())
	a.doneWg.Add(1)
	go a.run()
	return a
}

func (a *Adapter) run() {
	defer a.doneWg.Done()
	ticker := time.NewTicker(a.opts.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.done:
			return
		case msg := <-a.publishCh:
			err := a.conn.Publish(msg.Topic, msg.Payload)
			if err != nil {
				select {
				case a.errorCh <- err:
				default:
					// discard
				}
			}
		case natsMsg := <-a.receivedNatsMsgCh:
			msg := &types.Message{Topic: natsMsg.Subject, Payload: natsMsg.Data}
			select {
			case a.deliveredCh <- msg:
			default:
				select {
				case a.errorCh <- ErrSlowConsumer:
				default:
					// discard
				}
			}
		case <-ticker.C:
			err := a.conn.Flush()
			if err != nil {
				select {
				case a.errorCh <- err:
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
	a.subs.lock.Lock()
	defer a.subs.lock.Unlock()
	if _, ok := a.subs.subs[topic]; ok {
		return ErrAlreadySubscribed
	}

	sub, err := a.conn.ChanSubscribe(topic, a.receivedNatsMsgCh)
	if err != nil {
		return err
	}
	a.subs.subs[topic] = sub
	return nil
}

func (a *Adapter) Unsubscribe(topic string) error {
	a.subs.lock.Lock()
	defer a.subs.lock.Unlock()
	if sub, ok := a.subs.subs[topic]; ok {
		err := sub.Unsubscribe()
		if err != nil {
			return err
		}
		delete(a.subs.subs, topic)
	}
	return nil
}

// Close closes an adapter and its underlying NATS connection.
func (a *Adapter) Close() {
	close(a.done)
	a.doneWg.Wait()
	a.conn.Close()
}

func (a *Adapter) natsErrorHandler() nats.ErrHandler {
	return func(_ *nats.Conn, _ *nats.Subscription, err error) {
		select {
		case a.errorCh <- err:
		default:
			// discard
		}
	}
}

func (a *Adapter) natsConnErrorHandler() nats.ConnErrHandler {
	return func(_ *nats.Conn, err error) {
		select {
		case a.errorCh <- err:
		default:
			// discard
		}
	}
}

// subscriptions manages topics (or subject in NATS term).
type subscriptions struct {
	lock sync.Mutex

	// keys are subjects that Adapter is subscribing.
	subs map[string]*nats.Subscription
}

func newSubscriptions() *subscriptions {
	return &subscriptions{
		subs: map[string]*nats.Subscription{},
	}
}

// options is a configuration of Adapter.
type options struct {
	publishChSize   int
	deliveredChSize int
	errorChSize     int
	flushInterval   time.Duration
}

func defaultOptions() options {
	return options{
		publishChSize:   100,
		deliveredChSize: 100,
		errorChSize:     100,
		flushInterval:   100 * time.Millisecond,
	}
}
