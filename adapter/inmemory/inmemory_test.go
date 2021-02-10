package inmemory_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara/adapter/inmemory"
	"github.com/genkami/kiara/types"
)

var (
	timeout      = 1 * time.Second
	shortTimeout = 10 * time.Millisecond
)

var _ = Describe("Inmemory", func() {
	var broker *inmemory.Broker

	BeforeEach(func() {
		broker = inmemory.NewBroker()
	})

	AfterEach(func() {
		broker.Close()
	})

	Describe("Publish", func() {
		Context("when the adapter itself is subscribing to the topic", func() {
			It("sends a message", func() {
				a := inmemory.NewAdapter(broker)
				defer a.Close()
				topic := "kfpemployees"
				err := a.Subscribe(topic)
				Expect(err).NotTo(HaveOccurred())
				payload := []byte("kikkeriki~~~")
				a.Publish() <- &types.Message{Topic: topic, Payload: payload}
				select {
				case <-time.After(timeout):
					Fail("message disappeared")
				case msg := <-a.Delivered():
					Expect(msg).To(Equal(&types.Message{Topic: topic, Payload: payload}))
				}
			})
		})

		Context("when the adapter is not subscribing to the topic", func() {
			It("does not send a message", func() {
				a := inmemory.NewAdapter(broker)
				defer a.Close()
				topic := "kfpemployees"
				payload := []byte("kikkeriki~~~")
				a.Publish() <- &types.Message{Topic: topic, Payload: payload}
				select {
				case <-time.After(shortTimeout):
				case <-a.Delivered():
					Fail("unexpected message arrived")
				}
			})
		})

		Context("when another adapter is subscribing to the topic", func() {
			It("sends a message", func() {
				pub := inmemory.NewAdapter(broker)
				defer pub.Close()
				sub := inmemory.NewAdapter(broker)
				defer sub.Close()
				topic := "kfpemployees"
				err := sub.Subscribe(topic)
				Expect(err).NotTo(HaveOccurred())
				payload := []byte("kikkeriki~~~")
				pub.Publish() <- &types.Message{Topic: topic, Payload: payload}
				select {
				case <-time.After(timeout):
					Fail("message disappeared")
				case msg := <-sub.Delivered():
					Expect(msg).To(Equal(&types.Message{Topic: topic, Payload: payload}))
				}
			})
		})

		Context("when more than one adapters are subscribing to the topic", func() {
			It("sends a message to all the adapters", func() {
				pub := inmemory.NewAdapter(broker)
				defer pub.Close()
				topic := "kfpemployees"

				size := 3
				subs := make([]*inmemory.Adapter, 0, size)
				for i := 0; i < size; i++ {
					sub := inmemory.NewAdapter(broker)
					err := sub.Subscribe(topic)
					Expect(err).NotTo(HaveOccurred())
					defer sub.Close()
					subs = append(subs, sub)
				}

				payload := []byte("kikkeriki~~~")
				pub.Publish() <- &types.Message{Topic: topic, Payload: payload}
				for _, sub := range subs {
					select {
					case <-time.After(timeout):
						Fail("message disappeared")
					case msg := <-sub.Delivered():
						Expect(msg).To(Equal(&types.Message{Topic: topic, Payload: payload}))
					}
				}
			})
		})
	})
})
