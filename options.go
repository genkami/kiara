package kiara

import (
	"github.com/genkami/kiara/codec/gob"
	"github.com/genkami/kiara/types"
)

const defaultErrorChannelSize = 100

// options is a configuration of PubSub.
type options struct {
	errorChSize int
	codec       types.Codec
}

func defaultOptions() options {
	return options{
		errorChSize: defaultErrorChannelSize,
		codec:       gob.Codec,
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

// ErrorChannelSize sets a size of a channel through which async errors are reported.
func ErrorChannelSize(size int) Option {
	return optionFunc(func(opts *options) {
		opts.errorChSize = size
	})
}
