package nats

import (
	"time"
)

var (
	defaultFlushInterval = 1 * time.Second
)

// options is a configuration of Adapter.
type options struct {
	flushInterval time.Duration
}

func defaultOptions() options {
	return options{
		flushInterval: defaultFlushInterval,
	}
}

// Option configures the Adapter.
type Option interface {
	apply(opts *options)
}

type optionFunc func(*options)

func (f optionFunc) apply(opts *options) {
	f(opts)
}

// FlushInterval sets the interval for flushing.
func FlushInterval(interval time.Duration) Option {
	return optionFunc(func(opts *options) {
		opts.flushInterval = interval
	})
}
