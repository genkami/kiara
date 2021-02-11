package kiara_test

import (
	"context"
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara"
	"github.com/genkami/kiara/adapter/inmemory"
)

func receiveMessage(channel interface{}) (interface{}, bool) {
	chanVal := reflect.ValueOf(channel)
	chanType := chanVal.Type()
	if chanType.Kind() != reflect.Chan {
		panic(fmt.Sprintf("expected channel but got %s", chanType))
	}
	if chanType.ChanDir()|reflect.RecvDir == 0 {
		panic(fmt.Sprintf("expected <-chan T but got %s", chanType))
	}
	timeAfter := time.After(timeoutExpectedNotToExceed)
	timeAfterVal := reflect.ValueOf(timeAfter)
	chosen, recv, recvOk := reflect.Select([]reflect.SelectCase{
		reflect.SelectCase{Dir: reflect.SelectRecv, Chan: chanVal},
		reflect.SelectCase{Dir: reflect.SelectRecv, Chan: timeAfterVal},
	})
	if chosen != 0 {
		return nil, false
	}
	if !recvOk {
		panic("BUG: message must be received")
	}
	return recv.Interface(), true
}

var _ = Describe("Codec", func() {
	AssertMessageIsSent := func(sent, channel interface{}) {
		It("can marshal and unmarshal the message", func() {
			broker := inmemory.NewBroker()
			defer broker.Close()
			pubsub := kiara.NewPubSub(inmemory.NewAdapter(broker))
			defer pubsub.Close()

			topic := "room:123"
			sub, err := pubsub.Subscribe(topic, channel)
			Expect(err).NotTo(HaveOccurred())
			defer sub.Unsubscribe()

			ctx, cancel := context.WithTimeout(context.Background(), timeoutExpectedNotToExceed)
			defer cancel()
			err = pubsub.Publish(ctx, topic, sent)
			Expect(err).NotTo(HaveOccurred())

			msg, ok := receiveMessage(channel)
			if !ok {
				Fail("timeout")
			}
			Expect(msg).To(Equal(sent))
		})
	}

	Context("when message is int", func() {
		var sent int = 123
		channel := make(chan int, 10)
		AssertMessageIsSent(sent, channel)
	})
})
