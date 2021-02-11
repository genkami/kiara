package msgpack_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmihailenco/msgpack/v5"

	codec "github.com/genkami/kiara/codec/msgpack"
)

type account struct {
	Name string
	Age  int
}

var _ = Describe("Msgpack", func() {
	Describe("Marshal", func() {
		It("converts the given data into MessagePack", func() {
			data := &account{Name: "Gura", Age: 9927}
			marshaled, err := codec.Codec.Marshal(data)
			Expect(err).NotTo(HaveOccurred())
			var unmarshaled account
			err = msgpack.Unmarshal(marshaled, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(&unmarshaled).To(Equal(data))
		})
	})

	Describe("Unmarshal", func() {
		It("parses MessagPack", func() {
			data := &account{Name: "Gura", Age: 9927}
			marshaled, err := msgpack.Marshal(data)
			Expect(err).NotTo(HaveOccurred())
			var unmarshaled account
			err = codec.Codec.Unmarshal(marshaled, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(&unmarshaled).To(Equal(data))
		})
	})
})
