package redis

import (
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Options", func() {
	newAdapter := func(opts ...Option) *Adapter {
		redisAddr := os.Getenv("KIARA_TEST_REDIS_ADDR")
		redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
		return NewAdapter(redisClient, opts...)
	}

	Describe("PublishChannelSize", func() {
		Context("when the option is not set", func() {
			It("uses the default value", func() {
				adapter := newAdapter()
				Expect(cap(adapter.publishCh)).To(Equal(defaultPublishChannelSize))
			})
		})

		Context("when the option is set", func() {
			It("uses the give value", func() {
				size := 123
				adapter := newAdapter(PublishChannelSize(size))
				Expect(cap(adapter.publishCh)).To(Equal(size))
			})
		})
	})

	Describe("DeliveredChannelSize", func() {
		Context("when the option is not set", func() {
			It("uses the default value", func() {
				adapter := newAdapter()
				Expect(cap(adapter.deliveredCh)).To(Equal(defaultDeliveredChannelSize))
			})
		})

		Context("when the option is set", func() {
			It("uses the give value", func() {
				size := 123
				adapter := newAdapter(DeliveredChannelSize(size))
				Expect(cap(adapter.deliveredCh)).To(Equal(size))
			})
		})
	})

	Describe("ErrorChannelSize", func() {
		Context("when the option is not set", func() {
			It("uses the default value", func() {
				adapter := newAdapter()
				Expect(cap(adapter.errorCh)).To(Equal(defaultErrorChannelSize))
			})
		})

		Context("when the option is set", func() {
			It("uses the give value", func() {
				size := 123
				adapter := newAdapter(ErrorChannelSize(size))
				Expect(cap(adapter.errorCh)).To(Equal(size))
			})
		})
	})

	Describe("SubscriptionTimeout", func() {
		Context("when the option is not set", func() {
			It("uses the default value", func() {
				adapter := newAdapter()
				Expect(adapter.opts.subscriptionTimeout).To(Equal(defaultSubscriptionTimeout))
			})
		})

		Context("when the option is set", func() {
			It("uses the give value", func() {
				to := 1 * time.Minute
				adapter := newAdapter(SubscriptionTimeout(to))
				Expect(adapter.opts.subscriptionTimeout).To(Equal(to))
			})
		})
	})

	Describe("PublishTimeout", func() {
		Context("when the option is not set", func() {
			It("uses the default value", func() {
				adapter := newAdapter()
				Expect(adapter.opts.publishTimeout).To(Equal(defaultPublishTimeout))
			})
		})

		Context("when the option is set", func() {
			It("uses the give value", func() {
				to := 1 * time.Minute
				adapter := newAdapter(PublishTimeout(to))
				Expect(adapter.opts.publishTimeout).To(Equal(to))
			})
		})
	})
})
