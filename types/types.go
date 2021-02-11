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

// Adapter is an abstract interface to send and receive Messsages.
type Adapter interface {
	// Publish returns a channel through which messages are published.
	Publish() chan<- *Message

	// Delivered returns a channel through which messages are delivered.
	Delivered() <-chan *Message

	// Errors returns a channel through which asyncronous errors are reported.
	Errors() <-chan error

	// Subscribe subscribes a topic.
	// It may return an error when the topic is already subscribed.
	Subscribe(topic string) error

	// Unsubscribe unsubscribes a topic.
	Unsubscribe(topic string) error

	// Close stops the adapter and releases resources that an adapter has.
	Close()
}

// Codec converts an arbitrary object into a byte slice.
type Codec interface {
	// Marshal converts `v` into a byte slice.
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal parses src into `v`.
	Unmarshal(src []byte, v interface{}) error
}
