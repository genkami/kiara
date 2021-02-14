package redis

import (
	"time"
)

var (
	defaultSubscriptionTimeout = 3 * time.Second
	defaultPublishTimeout      = 3 * time.Second
)

// option is a configuration of Adapter.
type options struct {
	subscriptionTimeout time.Duration
	publishTimeout      time.Duration
}

func defaultOptions() options {
	return options{
		subscriptionTimeout: defaultSubscriptionTimeout,
		publishTimeout:      defaultPublishTimeout,
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

// SubscriptionTimeout sets the timeout for a subscription request.
func SubscriptionTimeout(timeout time.Duration) Option {
	return optionFunc(func(opts *options) {
		opts.subscriptionTimeout = timeout
	})
}

// PublishTimeout sets the timeout for a publish request.
func PublishTimeout(timeout time.Duration) Option {
	return optionFunc(func(opts *options) {
		opts.publishTimeout = timeout
	})
}
