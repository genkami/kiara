package commontest

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara/types"
)

var (
	timeoutExpectedNotToExceed = 300 * time.Millisecond
	timeoutExpectedToExceed    = 10 * time.Millisecond
)

type AdapterEnv interface {
	Setup()
	Teardown()
	NewAdapter() types.Adapter
}

func AssertAdapterIsImplementedCorrectly(env AdapterEnv) {
	BeforeEach(func() {
		env.Setup()
	})

	AfterEach(func() {
		env.Teardown()
	})

	Describe("Publish", func() {
		Context("when the adapter itself is subscribing to the topic", func() {
			It("sends a message", func() {
				adapter := env.NewAdapter()
				defer adapter.Close()
				topic := "kfpemployees"
				err := adapter.Subscribe(topic)
				Expect(err).NotTo(HaveOccurred())
				payload := []byte("kikkeriki~~~")
				adapter.Publish() <- &types.Message{Topic: topic, Payload: payload}
				select {
				case <-time.After(timeoutExpectedNotToExceed):
					Fail("message disappeared")
				case msg := <-adapter.Delivered():
					Expect(msg).To(Equal(&types.Message{Topic: topic, Payload: payload}))
				}
			})
		})

		Context("when the adapter is not subscribing to the topic", func() {
			It("does not send a message", func() {
				adapter := env.NewAdapter()
				defer adapter.Close()
				topic := "kfpemployees"
				payload := []byte("kikkeriki~~~")
				adapter.Publish() <- &types.Message{Topic: topic, Payload: payload}
				select {
				case <-time.After(timeoutExpectedToExceed):
				case <-adapter.Delivered():
					Fail("unexpected message arrived")
				}
			})
		})

		Context("when another adapter is subscribing to the topic", func() {
			It("sends a message", func() {
				pub := env.NewAdapter()
				defer pub.Close()
				sub := env.NewAdapter()
				defer sub.Close()
				topic := "kfpemployees"
				err := sub.Subscribe(topic)
				Expect(err).NotTo(HaveOccurred())
				payload := []byte("kikkeriki~~~")
				pub.Publish() <- &types.Message{Topic: topic, Payload: payload}
				select {
				case <-time.After(timeoutExpectedNotToExceed):
					Fail("message disappeared")
				case msg := <-sub.Delivered():
					Expect(msg).To(Equal(&types.Message{Topic: topic, Payload: payload}))
				}
			})
		})

		Context("when more than one adapters are subscribing to the topic", func() {
			It("sends a message to all the adapters", func() {
				pub := env.NewAdapter()
				defer pub.Close()
				topic := "kfpemployees"

				size := 3
				subs := make([]types.Adapter, 0, size)
				for i := 0; i < size; i++ {
					sub := env.NewAdapter()
					err := sub.Subscribe(topic)
					Expect(err).NotTo(HaveOccurred())
					defer sub.Close()
					subs = append(subs, sub)
				}

				payload := []byte("kikkeriki~~~")
				pub.Publish() <- &types.Message{Topic: topic, Payload: payload}
				for _, sub := range subs {
					select {
					case <-time.After(timeoutExpectedNotToExceed):
						Fail("message disappeared")
					case msg := <-sub.Delivered():
						Expect(msg).To(Equal(&types.Message{Topic: topic, Payload: payload}))
					}
				}
			})
		})
	})
}
