package nats

import (
	"github.com/nats-io/nats.go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara/adapter/internal/commontest"
)

var _ = Describe("Options", func() {
	newAdapter := func(opts ...Option) *Adapter {
		natsUrl := commontest.GetEnv("KIARA_TEST_NATS_URL")
		conn, err := nats.Connect(natsUrl)
		Expect(err).NotTo(HaveOccurred())
		return NewAdapter(conn, opts...)
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

	Describe("FlushInterval", func() {
		Context("when the option is not set", func() {
			It("uses the default value", func() {
				adapter := newAdapter()
				Expect(adapter.opts.flushInterval).To(Equal(defaultFlushInterval))
			})
		})
	})
})
