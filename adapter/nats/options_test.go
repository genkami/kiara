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

	Describe("FlushInterval", func() {
		Context("when the option is not set", func() {
			It("uses the default value", func() {
				adapter := newAdapter()
				Expect(adapter.opts.flushInterval).To(Equal(defaultFlushInterval))
			})
		})
	})
})
