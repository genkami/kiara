package kiara

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/genkami/kiara/types"
)

var (
	// This error is reported through PubSub.Errors() when a channel through which messages are sent is full.
	ErrSlowConsumer = errors.New("consumer is too slow; message discarded")

	// This error is returned when a given context is cancelled.
	ErrCancelled = errors.New("cancelled")

	// This error is returned when the second argument of PubSub.Subscribe() is not a channel or the direction of the channel is not <-.
	ErrArgumentMustBeChannel = errors.New("argument must be a channel")
)

// PubSub provides a way to send and receive arbitrary data.
type PubSub struct {
	adapter     types.Adapter
	opts        options
	publishCh   chan *types.Message
	deliveredCh chan *types.Message
	errorCh     chan error
	done        chan struct{}
	doneWg      sync.WaitGroup
	state       pubSubState
}

// NewPubSub creates a new PubSub and starts its underlying adapter.
func NewPubSub(adapter types.Adapter, options ...Option) *PubSub {
	opts := defaultOptions()
	for _, o := range options {
		o.apply(&opts)
	}

	publishCh := make(chan *types.Message, opts.publishChSize)
	deliveredCh := make(chan *types.Message, opts.deliveredChSize)
	errorCh := make(chan error, opts.errorChSize)
	pipe := &types.Pipe{
		Publish:   publishCh,
		Delivered: deliveredCh,
		Errors:    errorCh,
	}

	p := &PubSub{
		adapter:     adapter,
		opts:        opts,
		publishCh:   publishCh,
		deliveredCh: deliveredCh,
		errorCh:     errorCh,
		done:        make(chan struct{}),
		state: pubSubState{
			subs: map[string]subscriptionSet{},
		},
	}
	adapter.Start(pipe)
	p.doneWg.Add(1)
	go p.run()
	return p
}

func (p *PubSub) run() {
	defer p.doneWg.Done()
	for {
		select {
		case <-p.done:
			return
		case msg := <-p.deliveredCh:
			p.deliver(msg)
		}
	}
}

// Close stops the PubSub and releases its resources.
// It also stop its underlying adapter so we don't need stopping adapters manually.
func (p *PubSub) Close() {
	close(p.done)
	p.doneWg.Wait()
	p.adapter.Stop()
}

// deliver delivers a message to all channels that are subscribing to a message's topic.
func (p *PubSub) deliver(msg *types.Message) {
	// Getting subscriptionSet and delivering messages to all its channels must be done with `state.lock` `RLock`ed
	// in order to guarantee that no messages are sent after `Unsubscribe`d.
	p.state.lock.RLock()
	defer p.state.lock.RUnlock()
	channels, ok := p.state.subs[msg.Topic]
	if !ok {
		return
	}
	channels = channels.Copy()
	channels.ForEach(func(ch interface{}) {
		p.deliverTo(ch, msg.Payload)
	})
}

// deliverTo parses a message and delivers it to the given channel.
// We do not share the parsed result with all channels that want the result in order
// to prevent the result from accidentally being accessed concurrently.
func (p *PubSub) deliverTo(channel interface{}, payload []byte) {
	chanVal := reflect.ValueOf(channel)
	elemType := chanVal.Type().Elem()
	var dataVal reflect.Value
	if elemType.Kind() != reflect.Ptr {
		dataVal = reflect.New(elemType)
	} else {
		dataVal = reflect.New(elemType.Elem())
	}
	// The type of `dataVal` is either `*elemType` or `elemType` itself here
	// in order to avoid creating a pointer to pointer.
	// Note that the type of `dataVal` is different from `elemType` if and
	// only if `elemType.Kind() != reflect.Ptr`
	err := p.opts.codec.Unmarshal(payload, dataVal.Interface())
	if err != nil {
		select {
		case p.errorCh <- err:
		default:
			// discard
		}
		return
	}
	if elemType.Kind() != reflect.Ptr {
		// As we described before, in this case the type of `dataVal` is
		// `*elemType`. So we should `Indirect` it so that `dataVal` can be
		// sent to `chanVal` (whose type is `chan<- elemType`).
		dataVal = reflect.Indirect(dataVal)
	}
	chosen, _, _ := reflect.Select([]reflect.SelectCase{
		reflect.SelectCase{Dir: reflect.SelectSend, Chan: chanVal, Send: dataVal},
		reflect.SelectCase{Dir: reflect.SelectDefault},
	})
	if chosen == 1 { // default:
		select {
		case p.errorCh <- ErrSlowConsumer:
		default:
			// discard
		}
	}
}

// Publish publishes `data` to the underlying message broker.
// This means `data` is sent to every channels that is `Subscribe`ing the same topic as the given one.
// It returns an error when it cannot prepare publishing due to marshaling error or being cancelled by `ctx`.
// Any other errors are reported asynchronously via PubSub.Errors().
func (p *PubSub) Publish(ctx context.Context, topic string, data interface{}) error {
	payload, err := p.opts.codec.Marshal(data)
	if err != nil {
		return err
	}
	msg := &types.Message{Topic: topic, Payload: payload}
	select {
	case p.publishCh <- msg:
	case <-ctx.Done():
		return ErrCancelled
	}
	return nil
}

// Errors returns a channel through which asynchronous errors are reported.
// When the channel is full, subsequent errors are discarded.
func (p *PubSub) Errors() <-chan error {
	return p.errorCh
}

// Subscribe binds a channel to the given topic.
// This means any messages that are `Publish`ed toghther with the same topic are
// sent to the given channel.
//
// A `channel` must be the type of `chan T` or `chan<- T` where `T` is any type that can be
// `Unmarshal`ed by the codec of the `PubSub`.
//
// Note that PubSub internally passes *T to its internal codec when T is not a pointer.
// In most cases you don't have to care about it but it may be confusing when the ccodec assumes that the data implements certain interfaces.
//
// It's ok to subscribe to one topic more than one times.
// In this case, messages are broadcasted to all channels that are subscribing to the topic.
func (p *PubSub) Subscribe(topic string, channel interface{}) (*Subscription, error) {
	chanType := reflect.TypeOf(channel)
	if chanType.Kind() != reflect.Chan {
		return nil, ErrArgumentMustBeChannel
	}
	if chanType.ChanDir()&reflect.SendDir == 0 {
		return nil, ErrArgumentMustBeChannel
	}

	p.state.lock.Lock()
	defer p.state.lock.Unlock()
	alreadySubscribed := false
	channels, ok := p.state.subs[topic]
	if !ok {
		channels = newSubscriptionSet()
		p.state.subs[topic] = channels
	}
	if len(channels) > 0 {
		alreadySubscribed = true
	}
	channels.Add(channel)
	if !alreadySubscribed {
		// this must be called with `state.lock` locked in order to avoid
		// race condition where all channels are removed from `state.subs` but
		// `p` continues subscribing to the topic.
		err := p.adapter.Subscribe(topic)
		if err != nil {
			return nil, err
		}
	}
	return &Subscription{
		topic:   topic,
		channel: channel,
		pubSub:  p,
	}, nil
}

func (p *PubSub) unsubscribe(topic string, channel interface{}) error {
	p.state.lock.Lock()
	defer p.state.lock.Unlock()
	channels, ok := p.state.subs[topic]
	if !ok {
		return nil
	}
	channels.Delete(channel)
	if channels.Len() <= 0 {
		delete(p.state.subs, topic)
		// this must be called with `state.lock` locked in order to avoid
		// race condition where some channels are added to `state.subs` but
		// `p` stops subscribing to the topic.
		return p.adapter.Unsubscribe(topic)
	}
	return nil
}

// pubSubState is an internal state of PubSub that cannot be accessed concurrently.
type pubSubState struct {
	lock sync.RWMutex
	subs map[string]subscriptionSet
}

// Subscription binds a channel to specific topic.
type Subscription struct {
	topic   string
	channel interface{} // guaranteed to be a channel
	pubSub  *PubSub
}

// Unsubscribe removes a binding from corresponding channel to its associated topic.
// Once `Unsubscribe` is returned, it is guaranteed that no more messages are sent to the channel.
func (s *Subscription) Unsubscribe() error {
	return s.pubSub.unsubscribe(s.topic, s.channel)
}

// subscriptionSet is a set of subscriptions that PubSub should deliver messages to.
// The key is guaranteed to be channels.
type subscriptionSet map[interface{}]struct{}

func newSubscriptionSet() subscriptionSet {
	return subscriptionSet(map[interface{}]struct{}{})
}

func (set subscriptionSet) Add(ch interface{}) {
	set[ch] = struct{}{}
}

func (set subscriptionSet) Delete(ch interface{}) {
	delete(set, ch)
}

func (set subscriptionSet) Copy() subscriptionSet {
	clone := newSubscriptionSet()
	for k, v := range set {
		clone[k] = v
	}
	return clone
}

func (set subscriptionSet) ForEach(fn func(interface{})) {
	for ch := range set {
		fn(ch)
	}
}

func (set subscriptionSet) Len() int {
	return len(set)
}
