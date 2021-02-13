package msgpack_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/genkami/kiara/codec/internal/commontest"
	codec "github.com/genkami/kiara/codec/json"
)

type account struct {
	Name string
	Age  int
}

var _ = Describe("Json", func() {
	Describe("Marshal", func() {
		It("converts the given data into JSON", func() {
			data := &account{Name: "Gura", Age: 9927}
			marshaled, err := codec.Codec.Marshal(data)
			Expect(err).NotTo(HaveOccurred())
			var unmarshaled account
			err = json.Unmarshal(marshaled, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(&unmarshaled).To(Equal(data))
		})
	})

	Describe("Unmarshal", func() {
		It("parses MessagPack", func() {
			data := &account{Name: "Gura", Age: 9927}
			marshaled, err := json.Marshal(data)
			Expect(err).NotTo(HaveOccurred())
			var unmarshaled account
			err = codec.Codec.Unmarshal(marshaled, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(&unmarshaled).To(Equal(data))
		})
	})

	commontest.AssertCodecCanMarshalAndUnmarshalAlmostEverything(codec.Codec)
})
