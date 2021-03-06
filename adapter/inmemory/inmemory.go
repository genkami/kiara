// Package inmemory provides a simple pubsub adapter mainly aimed at testing.
package inmemory

import (
	"errors"
	"log"
	"sync"

	"github.com/genkami/kiara/types"
)

var (
	ErrAlreadySubscribed = errors.New("already subscribed")
	ErrNotSubscribed     = errors.New("not subscribed")
)

// Broker is a simple message broker.
type Broker struct {
	lock     sync.RWMutex
	messages chan *types.Message
	adapters adapterSet
	opts     brokerOptions
	done     chan struct{}
}

// NewBroker creates a new Broker.
func NewBroker() *Broker {
	opts := defaultBrokerOptions()
	b := &Broker{
		adapters: newAdapterSet(),
		messages: make(chan *types.Message, opts.messagesChSize),
		opts:     opts,
		done:     make(chan struct{}),
	}
	go b.run()
	return b
}

func (b *Broker) run() {
	for {
		select {
		case <-b.done:
			return
		case msg := <-b.messages:
			b.send(msg)
		}
	}
}

func (b *Broker) send(msg *types.Message) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	b.adapters.ForEach(func(a *Adapter) {
		a.noticed <- msg
	})
}

func (b *Broker) registerAdapter(a *Adapter) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.adapters.Has(a) {
		log.Panicf("BUG: the same adapter registered more than once: %v", a)
	}
	b.adapters.Add(a)
}

func (b *Broker) unregisterAdapter(a *Adapter) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if !b.adapters.Has(a) {
		log.Panicf("BUG: unregistering Adapter that is not registered: %v", a)
	}
	b.adapters.Delete(a)
}

// Close stops the broker and releases its resources.
func (b *Broker) Close() {
	close(b.done)
}

// brokerOptions is a configuration of Broker.
type brokerOptions struct {
	messagesChSize int
}

// Currently we do not provide any methods to configure these parameters because no one wants to do this.
func defaultBrokerOptions() brokerOptions {
	return brokerOptions{
		messagesChSize: 10,
	}
}

// adapterSet is a set of Adapters.
type adapterSet map[*Adapter]struct{}

// newAdapterSet returns an empty set of adapters.
func newAdapterSet() adapterSet {
	return adapterSet(map[*Adapter]struct{}{})
}

func (set adapterSet) Add(a *Adapter) {
	set[a] = struct{}{}
}

func (set adapterSet) Delete(a *Adapter) {
	delete(set, a)
}

func (set adapterSet) Has(a *Adapter) bool {
	_, ok := set[a]
	return ok
}

func (set adapterSet) ForEach(fn func(*Adapter)) {
	for a := range set {
		fn(a)
	}
}

// Adapter is an adapter that sends messages through Broker.
type Adapter struct {
	broker  *Broker
	subLock sync.RWMutex
	topics  topicSet
	pipe    *types.Pipe
	noticed chan *types.Message
	done    chan struct{}
	opts    adapterOptions
}

var _ types.Adapter = &Adapter{}

func NewAdapter(broker *Broker) *Adapter {
	opts := defaultAdapterOptions()
	a := &Adapter{
		broker:  broker,
		topics:  newTopicSet(),
		noticed: make(chan *types.Message, opts.noticedChSize),
		done:    make(chan struct{}),
		opts:    opts,
	}
	return a
}

func (a *Adapter) Start(pipe *types.Pipe) {
	a.pipe = pipe
	a.broker.registerAdapter(a)
	go a.run()
}

func (a *Adapter) run() {
	for {
		select {
		case <-a.done:
			return
		case msg := <-a.noticed:
			a.deliver(msg)
		case msg := <-a.pipe.Publish:
			a.broker.messages <- msg
		}
	}
}

func (a *Adapter) deliver(msg *types.Message) {
	a.subLock.RLock()
	if a.topics.Has(msg.Topic) {
		a.subLock.RUnlock()
		a.pipe.Delivered <- msg
	} else {
		a.subLock.RUnlock()
	}
}

func (a *Adapter) Subscribe(topic string) error {
	a.subLock.Lock()
	defer a.subLock.Unlock()
	if a.topics.Has(topic) {
		return ErrAlreadySubscribed
	}
	a.topics.Add(topic)
	return nil
}

func (a *Adapter) Unsubscribe(topic string) error {
	a.subLock.Lock()
	defer a.subLock.Unlock()
	if !a.topics.Has(topic) {
		return ErrNotSubscribed
	}
	a.topics.Delete(topic)
	return nil
}

func (a *Adapter) Stop() {
	a.broker.unregisterAdapter(a)
	close(a.done)
}

// adapterOptions is a configuration of Adapter.
type adapterOptions struct {
	noticedChSize int
}

// Currently we do not provide any methods to configure these parameters because no one wants to do this.
func defaultAdapterOptions() adapterOptions {
	return adapterOptions{
		noticedChSize: 10,
	}
}

// topicSet is a set of topics.
type topicSet map[string]struct{}

func newTopicSet() topicSet {
	return topicSet(map[string]struct{}{})
}

func (set topicSet) Add(topic string) {
	set[topic] = struct{}{}
}

func (set topicSet) Delete(topic string) {
	delete(set, topic)
}

func (set topicSet) Has(topic string) bool {
	_, ok := set[topic]
	return ok
}
