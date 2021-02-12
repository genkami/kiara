package kiara

import (
	"github.com/genkami/kiara/codec/gob"
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

func WithCodec(codec types.Codec) Option {
	return optionFunc(func(opts *options) {
		opts.codec = codec
	})
}
