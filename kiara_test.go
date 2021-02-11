package kiara_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara"
	"github.com/genkami/kiara/adapter/inmemory"
)

const defaultChSize = 10

var (
	timeoutExpectedNotToExceed = 300 * time.Millisecond
	timeoutExpectedToExceed    = 10 * time.Millisecond
)

var _ = Describe("Kiara", func() {
	Describe("Publish", func() {
		var (
			broker *inmemory.Broker
			pubsub *kiara.PubSub
		)

		BeforeEach(func() {
			broker = inmemory.NewBroker()
			adapter := inmemory.NewAdapter(broker)
			pubsub = kiara.NewPubSub(adapter)
		})

		AfterEach(func() {
			pubsub.Close()
			broker.Close()
		})

		Context("when a topic is subscribed", func() {
			It("sends a message to the subscriber", func() {
				topic := "room:123"
				ch := make(chan interface{}, defaultChSize)
				sub, err := pubsub.Subscribe(topic, ch)
				Expect(err).NotTo(HaveOccurred())
				defer sub.Unsubscribe()
				ctx, cancel := context.WithTimeout(context.Background(), timeoutExpectedNotToExceed)
				defer cancel()
				var sent int = 123
				err = pubsub.Publish(ctx, topic, sent)
				Expect(err).NotTo(HaveOccurred())
				var received int
				select {
				case receivedAny := <-ch:
					received = receivedAny.(int)
				case <-time.After(timeoutExpectedNotToExceed):
					Fail("timeout")
				}
				Expect(received).To(Equal(sent))
			})
		})

		Context("when a topic is not subscribed", func() {
			It("does not send any message to the subscriber", func() {
				topic := "room:123"
				anotherTopic := "room:321"
				ch := make(chan interface{}, defaultChSize)
				sub, err := pubsub.Subscribe(anotherTopic, ch)
				Expect(err).NotTo(HaveOccurred())
				defer sub.Unsubscribe()
				ctx, cancel := context.WithTimeout(context.Background(), timeoutExpectedNotToExceed)
				defer cancel()
				var sent int = 123
				err = pubsub.Publish(ctx, topic, sent)
				Expect(err).NotTo(HaveOccurred())
				select {
				case msg := <-ch:
					Fail(fmt.Sprintf("expected no message but got %v", msg))
				case <-time.After(timeoutExpectedToExceed):
					// OK
				}
			})
		})
	})
})
