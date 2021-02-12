package kiara

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara/adapter/inmemory"
	"github.com/genkami/kiara/codec/gob"
	"github.com/genkami/kiara/codec/msgpack"
)

type account struct {
	Name string
	Age  int
}

var _ = Describe("Options", func() {
	var (
		broker *inmemory.Broker
	)

	BeforeEach(func() {
		broker = inmemory.NewBroker()
	})

	AfterEach(func() {
		broker.Close()
	})

	newPubSub := func(opts ...Option) *PubSub {
		adapter := inmemory.NewAdapter(broker)
		return NewPubSub(adapter, opts...)
	}

	Describe("WithCodec", func() {
		Context("when the option is not set", func() {
			It("uses the gob codec", func() {
				pubsub := newPubSub()
				data := &account{Name: "Gura", Age: 9927}
				marshaled, err := pubsub.opts.codec.Marshal(data)
				Expect(err).NotTo(HaveOccurred())
				var unmarshaled account
				err = gob.Codec.Unmarshal(marshaled, &unmarshaled)
				Expect(err).NotTo(HaveOccurred())
				Expect(&unmarshaled).To(Equal(data))
			})
		})

		Context("when the option is set", func() {
			It("uses the given codec", func() {
				pubsub := newPubSub(WithCodec(msgpack.Codec))
				data := &account{Name: "Gura", Age: 9927}
				marshaled, err := pubsub.opts.codec.Marshal(data)
				Expect(err).NotTo(HaveOccurred())
				var unmarshaled account
				err = msgpack.Codec.Unmarshal(marshaled, &unmarshaled)
				Expect(err).NotTo(HaveOccurred())
				Expect(&unmarshaled).To(Equal(data))
			})
		})
	})

	Describe("ErrorChannelSize", func() {
		Context("when the option is not set", func() {
			It("uses the default size", func() {
				pubsub := newPubSub()
				Expect(cap(pubsub.errorCh)).To(Equal(defaultErrorChannelSize))
			})
		})

		Context("when the option is set", func() {
			It("uses the given size", func() {
				size := 445
				pubsub := newPubSub(ErrorChannelSize(size))
				Expect(cap(pubsub.errorCh)).To(Equal(size))
			})
		})
	})
})
