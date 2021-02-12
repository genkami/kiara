package commontest

import (
	"context"
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara"
	"github.com/genkami/kiara/adapter/inmemory"
	"github.com/genkami/kiara/types"
)

const defaultChSize = 10

var timeoutExpectedNotToExceed = 300 * time.Millisecond

func receiveMessage(channel interface{}) (interface{}, bool) {
	chanVal := reflect.ValueOf(channel)
	chanType := chanVal.Type()
	if chanType.Kind() != reflect.Chan {
		panic(fmt.Sprintf("expected channel but got %s", chanType))
	}
	if chanType.ChanDir()&reflect.RecvDir == 0 {
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

func AssertCodecCanMarshalAndUnmarshal(codec types.Codec, sent, channel interface{}) {
	It("can marshal and unmarshal the message", func() {
		broker := inmemory.NewBroker()
		defer broker.Close()
		pubsub := kiara.NewPubSub(
			inmemory.NewAdapter(broker),
			kiara.WithCodec(codec),
		)
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

func AssertCodecCanMarshalAndUnmarshalAlmostEverything(codec types.Codec) {
	Context("when message is int", func() {
		var sent int = 123
		channel := make(chan int, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is *int", func() {
		var sent int = 123
		channel := make(chan *int, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, &sent, channel)
	})

	Context("when message is int8", func() {
		var sent int8 = 123
		channel := make(chan int8, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is uint", func() {
		var sent uint = 123
		channel := make(chan uint, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is uint8", func() {
		var sent uint8 = 123
		channel := make(chan uint8, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is float32", func() {
		var sent float32 = 1.23
		channel := make(chan float32, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is float64", func() {
		var sent float64 = 1.23
		channel := make(chan float64, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is string", func() {
		var sent string = "hello"
		channel := make(chan string, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is *string", func() {
		var sent string = "hello"
		channel := make(chan *string, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, &sent, channel)
	})

	Context("when message is slice", func() {
		var sent []string = []string{"hello", "world"}
		channel := make(chan []string, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is pointer to slice", func() {
		var sent []string = []string{"hello", "world"}
		channel := make(chan *[]string, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, &sent, channel)
	})

	Context("when message is array", func() {
		var sent [3]int = [3]int{1, 2, 3}
		channel := make(chan [3]int, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is pointer to array", func() {
		var sent [3]int = [3]int{1, 2, 3}
		channel := make(chan *[3]int, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, &sent, channel)
	})

	Context("when message is defined type", func() {
		var sent myInt = 123
		channel := make(chan myInt, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is pointer to defined type", func() {
		var sent myInt = 123
		channel := make(chan *myInt, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, &sent, channel)
	})

	Context("when message is struct", func() {
		var sent myStruct = myStruct{ID: 123, Company: "KFP"}
		channel := make(chan myStruct, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})

	Context("when message is pointer to struct", func() {
		var sent *myStruct = &myStruct{ID: 123, Company: "KFP"}
		channel := make(chan *myStruct, defaultChSize)
		AssertCodecCanMarshalAndUnmarshal(codec, sent, channel)
	})
}

type myInt int

type myStruct struct {
	ID      int
	Company string
}
