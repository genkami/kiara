package commontest

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara/types"
)

var (
	timeoutExpectedNotToExceed = 3 * time.Second
	timeoutExpectedToExceed    = 10 * time.Millisecond
)

func GetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		Fail(fmt.Sprintf("environment variable %s not set", name))
	}
	return value
}

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

	newPipe := func() (chan *types.Message, chan *types.Message, chan error, *types.Pipe) {
		publish := make(chan *types.Message, 10)
		delivered := make(chan *types.Message, 10)
		errors := make(chan error, 10)
		pipe := &types.Pipe{
			Publish:   publish,
			Delivered: delivered,
			Errors:    errors,
		}
		return publish, delivered, errors, pipe
	}

	Describe("Publish", func() {
		Context("when the adapter itself is subscribing to the topic", func() {
			It("sends a message", func() {
				publish, delivered, errors, pipe := newPipe()
				adapter := env.NewAdapter()
				adapter.Start(pipe)
				defer adapter.Stop()
				topic := "kfpemployees"
				err := adapter.Subscribe(topic)
				Expect(err).NotTo(HaveOccurred())
				payload := []byte("kikkeriki~~~")
				publish <- &types.Message{Topic: topic, Payload: payload}
				select {
				case <-time.After(timeoutExpectedNotToExceed):
					select {
					case err := <-errors:
						Expect(err).NotTo(HaveOccurred())
					default:
						Fail("message disappeared")
					}
				case msg := <-delivered:
					Expect(msg).To(Equal(&types.Message{Topic: topic, Payload: payload}))
				}
			})
		})

		Context("when the adapter is not subscribing to the topic", func() {
			It("does not send a message", func() {
				publish, delivered, _, pipe := newPipe()
				adapter := env.NewAdapter()
				adapter.Start(pipe)
				defer adapter.Stop()
				topic := "kfpemployees"
				payload := []byte("kikkeriki~~~")
				publish <- &types.Message{Topic: topic, Payload: payload}
				select {
				case <-time.After(timeoutExpectedToExceed):
				case <-delivered:
					Fail("unexpected message arrived")
				}
			})
		})

		Context("when another adapter is subscribing to the topic", func() {
			It("sends a message", func() {
				publish, _, _, pubPipe := newPipe()
				pub := env.NewAdapter()
				pub.Start(pubPipe)
				defer pub.Stop()

				_, delivered, errors, subPipe := newPipe()
				sub := env.NewAdapter()
				sub.Start(subPipe)
				defer sub.Stop()

				topic := "kfpemployees"
				err := sub.Subscribe(topic)
				Expect(err).NotTo(HaveOccurred())
				payload := []byte("kikkeriki~~~")
				publish <- &types.Message{Topic: topic, Payload: payload}
				select {
				case <-time.After(timeoutExpectedNotToExceed):
					select {
					case err := <-errors:
						Expect(err).NotTo(HaveOccurred())
					default:
						Fail("message disappeared")
					}
				case msg := <-delivered:
					Expect(msg).To(Equal(&types.Message{Topic: topic, Payload: payload}))
				}
			})
		})

		Context("when more than one adapters are subscribing to the topic", func() {
			It("sends a message to all the adapters", func() {
				publish, _, _, pubPipe := newPipe()
				pub := env.NewAdapter()
				pub.Start(pubPipe)
				defer pub.Stop()
				topic := "kfpemployees"

				size := 3
				deliveredList := make([]chan *types.Message, 0, size)
				errorsList := make([]chan error, 0, size)
				for i := 0; i < size; i++ {
					_, delivered, errors, pipe := newPipe()
					sub := env.NewAdapter()
					sub.Start(pipe)
					defer sub.Stop()
					err := sub.Subscribe(topic)
					Expect(err).NotTo(HaveOccurred())

					deliveredList = append(deliveredList, delivered)
					errorsList = append(errorsList, errors)
				}

				payload := []byte("kikkeriki~~~")
				publish <- &types.Message{Topic: topic, Payload: payload}
				for i, delivered := range deliveredList {
					errors := errorsList[i]
					select {
					case <-time.After(timeoutExpectedNotToExceed):
						select {
						case err := <-errors:
							Expect(err).NotTo(HaveOccurred())
						default:
							Fail("message disappeared")
						}
					case msg := <-delivered:
						Expect(msg).To(Equal(&types.Message{Topic: topic, Payload: payload}))
					}
				}
			})
		})
	})
}
