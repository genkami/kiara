package redis

import (
	"time"
)

const (
	defaultPublishChannelSize   = 100
	defaultDeliveredChannelSize = 100
	defaultErrorChannelSize     = 100
)

var (
	defaultSubscriptionTimeout = 3 * time.Second
	defaultPublishTimeout      = 3 * time.Second
)

// option is a configuration of Adapter.
type options struct {
	publishChSize       int
	deliveredChSize     int
	errorChSize         int
	subscriptionTimeout time.Duration
	publishTimeout      time.Duration
}

func defaultOptions() options {
	return options{
		publishChSize:       defaultPublishChannelSize,
		deliveredChSize:     defaultDeliveredChannelSize,
		errorChSize:         defaultErrorChannelSize,
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

// ErrorChannelSize sets the size of a channel through which errors are reported.
func ErrorChannelSize(size int) Option {
	return optionFunc(func(opts *options) {
		opts.errorChSize = size
	})
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
