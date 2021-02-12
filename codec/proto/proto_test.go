package proto_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	"github.com/genkami/kiara"
	"github.com/genkami/kiara/adapter/inmemory"
	"github.com/genkami/kiara/codec/internal/commontest"
	pb "github.com/genkami/kiara/codec/internal/testproto"
	codec "github.com/genkami/kiara/codec/proto"
)

const defaultChSize = 10

var timeoutExpectedNotToExceed = 3 * time.Second

func AssertCodecCanMarshalAndUnmarshal(sent proto.Message, channel interface{}) {
	It("can marshal and unmarshal the message", func() {
		broker := inmemory.NewBroker()
		defer broker.Close()
		pubsub := kiara.NewPubSub(
			inmemory.NewAdapter(broker),
			kiara.WithCodec(codec.Codec),
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

		msg, ok := commontest.ReceiveMessage(channel)
		if !ok {
			select {
			case err := <-pubsub.Errors():
				Expect(err).NotTo(HaveOccurred())
			default:
				Fail("timeout")
			}
		}
		protoMsg, ok := msg.(proto.Message)
		if !ok {
			Fail(fmt.Sprintf("expected proto.Message but got %T", msg))
		}
		Expect(proto.Equal(sent, protoMsg)).To(BeTrue())
	})
}

var _ = Describe("Proto", func() {
	Describe("Marshal", func() {
		It("converts the given data into Protocol Buffers", func() {
			data := &pb.Channel{Name: "Callipe Mori Ch.", Subscribers: 1000000}
			marshaled, err := codec.Codec.Marshal(data)
			Expect(err).NotTo(HaveOccurred())
			var unmarshaled pb.Channel
			err = proto.Unmarshal(marshaled, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(proto.Equal(&unmarshaled, data)).To(BeTrue())
		})
	})

	Describe("Unmarshal", func() {
		It("parses Protocol Buffers", func() {
			data := &pb.Channel{Name: "Watson Amelia Ch.", Subscribers: 1000000}
			marshaled, err := proto.Marshal(data)
			Expect(err).NotTo(HaveOccurred())
			var unmarshaled pb.Channel
			err = codec.Codec.Unmarshal(marshaled, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(proto.Equal(&unmarshaled, data)).To(BeTrue())
		})
	})

	AssertCodecCanMarshalAndUnmarshal(
		&pb.Channel{Name: "Watson Amelia Ch.", Subscribers: 1000000},
		make(chan *pb.Channel, defaultChSize))
})
