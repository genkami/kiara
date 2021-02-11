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

	Describe("Subscribe", func() {
		Context("when the second argument is `chan T`", func() {
			It("can subscribe to the topic", func() {
				var ch chan int = make(chan int, defaultChSize)
				sub, err := pubsub.Subscribe("room:123", ch)
				Expect(err).NotTo(HaveOccurred())
				sub.Unsubscribe()
			})
		})

		Context("when the second argument is `chan<- T`", func() {
			It("can subscribe to the topic", func() {
				var ch chan<- int = make(chan int, defaultChSize)
				sub, err := pubsub.Subscribe("room:123", ch)
				Expect(err).NotTo(HaveOccurred())
				sub.Unsubscribe()
			})
		})

		Context("when the second argument is `<-chan T`", func() {
			It("returns an error", func() {
				var ch <-chan int = make(chan int, defaultChSize)
				_, err := pubsub.Subscribe("room:123", ch)
				Expect(err).To(MatchError(kiara.ErrArgumentMustBeChannel))
			})
		})

		Context("when the second argument is not a channel", func() {
			It("returns an error", func() {
				var ch []int = make([]int, defaultChSize)
				_, err := pubsub.Subscribe("room:123", ch)
				Expect(err).To(MatchError(kiara.ErrArgumentMustBeChannel))
			})
		})
	})

	Describe("Publish", func() {
		Context("when a topic is subscribed", func() {
			It("sends a message to the subscriber", func() {
				topic := "room:123"
				ch := make(chan int, defaultChSize)
				sub, err := pubsub.Subscribe(topic, ch)
				Expect(err).NotTo(HaveOccurred())
				defer sub.Unsubscribe()
				ctx, cancel := context.WithTimeout(context.Background(), timeoutExpectedNotToExceed)
				defer cancel()
				var sent int = 123
				err = pubsub.Publish(ctx, topic, sent)
				Expect(err).NotTo(HaveOccurred())
				select {
				case received := <-ch:
					Expect(received).To(Equal(sent))
				case <-time.After(timeoutExpectedNotToExceed):
					Fail("timeout")
				}
			})
		})

		Context("when a topic is not subscribed", func() {
			It("does not send any message to the subscriber", func() {
				topic := "room:123"
				anotherTopic := "room:321"
				ch := make(chan int, defaultChSize)
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

		Context("when a topic was subscribed but is now unsubscribed", func() {
			It("stops sending messages to the subscriber", func() {
				topic := "room:123"
				unsubscribed := false
				ch := make(chan int, defaultChSize)
				sub, err := pubsub.Subscribe(topic, ch)
				Expect(err).NotTo(HaveOccurred())
				defer func() {
					if !unsubscribed {
						sub.Unsubscribe()
					}
				}()

				ctx, cancel := context.WithTimeout(context.Background(), timeoutExpectedNotToExceed)
				defer cancel()
				var sent int = 123
				err = pubsub.Publish(ctx, topic, sent)
				Expect(err).NotTo(HaveOccurred())
				select {
				case received := <-ch:
					Expect(received).To(Equal(sent))
				case <-time.After(timeoutExpectedNotToExceed):
					Fail("timeout")
				}

				unsubscribed = true
				err = sub.Unsubscribe()
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel = context.WithTimeout(context.Background(), timeoutExpectedNotToExceed)
				defer cancel()
				sent = 456
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

		Context("when a topic is subscribed to by multiple channels", func() {
			It("sends a message to all subscribers", func() {
				topic := "room:123"

				n := 3
				chs := make([]chan int, 0, n)
				for i := 0; i < n; i++ {
					ch := make(chan int, defaultChSize)
					sub, err := pubsub.Subscribe(topic, ch)
					Expect(err).NotTo(HaveOccurred())
					defer sub.Unsubscribe()
					chs = append(chs, ch)
				}

				ctx, cancel := context.WithTimeout(context.Background(), timeoutExpectedNotToExceed)
				defer cancel()
				var sent int = 123
				err := pubsub.Publish(ctx, topic, sent)
				Expect(err).NotTo(HaveOccurred())
				for i, ch := range chs {
					select {
					case received := <-ch:
						Expect(received).To(Equal(sent))
					case <-time.After(timeoutExpectedNotToExceed):
						Fail(fmt.Sprintf("%d: timeout", i))
					}
				}
			})
		})

		Context("when a topic is subscribed to by multiple channels and some of then stops subscribing", func() {
			It("sends a message only to the channels that continues subscribing", func() {
				topic := "room:123"

				subAUnsubscribed := false
				chA := make(chan int, defaultChSize)
				subA, err := pubsub.Subscribe(topic, chA)
				Expect(err).NotTo(HaveOccurred())
				defer func() {
					if !subAUnsubscribed {
						subA.Unsubscribe()
					}
				}()

				chB := make(chan int, defaultChSize)
				subB, err := pubsub.Subscribe(topic, chB)
				Expect(err).NotTo(HaveOccurred())
				defer subB.Unsubscribe()

				ctx, cancel := context.WithTimeout(context.Background(), timeoutExpectedNotToExceed)
				defer cancel()
				var sent int = 123
				err = pubsub.Publish(ctx, topic, sent)
				Expect(err).NotTo(HaveOccurred())

				// ensure that all `ch`s are subscribing to the topic
				for i, ch := range []chan int{chA, chB} {
					select {
					case received := <-ch:
						Expect(received).To(Equal(sent))
					case <-time.After(timeoutExpectedNotToExceed):
						Fail(fmt.Sprintf("%d: timeout", i))
					}
				}

				subAUnsubscribed = true
				err = subA.Unsubscribe()
				Expect(err).NotTo(HaveOccurred())

				ctx, cancel = context.WithTimeout(context.Background(), timeoutExpectedNotToExceed)
				defer cancel()
				sent = 456
				err = pubsub.Publish(ctx, topic, sent)
				Expect(err).NotTo(HaveOccurred())

				select {
				case msg := <-chA:
					Fail(fmt.Sprintf("A: expected no message but got %v", msg))
				case <-time.After(timeoutExpectedToExceed):
					// OK
				}

				select {
				case received := <-chB:
					Expect(received).To(Equal(sent))
				case <-time.After(timeoutExpectedNotToExceed):
					Fail("timeout")
				}
			})
		})
	})
})
