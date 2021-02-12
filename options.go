package kiara

import (
	"github.com/genkami/kiara/codec/msgpack"
	"github.com/genkami/kiara/types"
)

// options is a configuration of PubSub.
type options struct {
	errorChSize int
	codec       types.Codec
}

func defaultOptions() options {
	return options{
		errorChSize: 10,
		codec:       msgpack.Codec, // TODO: change this to gob
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

func WithCustomCodec(codec types.Codec) Option {
	return optionFunc(func(opts *options) {
		opts.codec = codec
	})
}
