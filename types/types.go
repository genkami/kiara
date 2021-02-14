// Package types provides types and interfaces that are needed to implement backend adapters.
package types

// Message represents a message that is sent over Adapters.
type Message struct {
	// Topic is a topic or a channel through which messages are sent.
	// Restrictions of topics are dependent on the concrete implementation of the Adapter.
	Topic string

	// Payload is the payload of the message.
	Payload []byte
}

// Pipe is a pipeline through which kiara.PubSub communicates with Adapters.
type Pipe struct {
	// Publish is a messages that are about to be published.
	// Adapters must watch this channel and send messages from this channel to backend message brokers.
	Publish <-chan *Message

	// Delivered is a messages that are sent from backend message brokers.
	// When messages are sent from Adapter's backend message brokers, Adapters must sent them to this channel.
	// When sending to this channel blocks, Adapters should not wait, discard succeeding messages, and send
	// errors to Errors channel to indicate that messages are dropped.
	Delivered chan<- *Message

	// Errors are asynchronous erros that occurred in Adapters.
	Errors chan<- error
}

// Adapter is an abstract interface to send and receive Messsages.
type Adapter interface {
	// Start starts communicating with `kiara.PubSub` through the given Pipe.
	Start(*Pipe)

	// Subscribe subscribes a topic.
	// It may return an error when the topic is already subscribed.
	Subscribe(topic string) error

	// Unsubscribe unsubscribes a topic.
	Unsubscribe(topic string) error

	// Stop stops the adapter and releases resources that an adapter has.
	Stop()
}

// Codec converts an arbitrary object into a byte slice.
type Codec interface {
	// Marshal converts `v` into a byte slice.
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal parses src into `v`.
	Unmarshal(src []byte, v interface{}) error
}
