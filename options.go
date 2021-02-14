package kiara

import (
	"github.com/genkami/kiara/codec/gob"
	"github.com/genkami/kiara/types"
)

const (
	defaultPublishChannelSize   = 100
	defaultDeliveredChannelSize = 100
	defaultErrorChannelSize     = 100
)

// options is a configuration of PubSub.
type options struct {
	publishChSize   int
	deliveredChSize int
	errorChSize     int
	codec           types.Codec
}

func defaultOptions() options {
	return options{
		publishChSize:   defaultPublishChannelSize,
		deliveredChSize: defaultDeliveredChannelSize,
		errorChSize:     defaultErrorChannelSize,
		codec:           gob.Codec,
	}
}

// Option configures PubSub.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(opts *options) {
	f(opts)
}

// WithCodec specifies a codec that PubSub uses to marshal and unmarshal messages.
//
// By default, messages are marshaled into gob format.
func WithCodec(codec types.Codec) Option {
	return optionFunc(func(opts *options) {
		opts.codec = codec
	})
}

// PublishChannelSize sets the size of a channel that contains messages that will be sent later.
func PublishChannelSize(size int) Option {
	return optionFunc(func(opts *options) {
		opts.publishChSize = size
	})
}

// DeliveredChannelSize sets the size of a channel that contains messages that are sent by the backend and about to be delivered.
func DeliveredChannelSize(size int) Option {
	return optionFunc(func(opts *options) {
		opts.deliveredChSize = size
	})
}

// ErrorChannelSize sets a size of a channel through which async errors are reported.
func ErrorChannelSize(size int) Option {
	return optionFunc(func(opts *options) {
		opts.errorChSize = size
	})
}
