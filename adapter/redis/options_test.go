package redis

import (
	"time"

	"github.com/go-redis/redis/v8"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara/adapter/internal/commontest"
)

var _ = Describe("Options", func() {
	newAdapter := func(opts ...Option) *Adapter {
		redisAddr := commontest.GetEnv("KIARA_TEST_REDIS_ADDR")
		redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
		return NewAdapter(redisClient, opts...)
	}

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
