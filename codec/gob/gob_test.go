package gob_test

import (
	"bytes"
	"encoding/gob"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	codec "github.com/genkami/kiara/codec/gob"
	"github.com/genkami/kiara/codec/internal/commontest"
)

type account struct {
	Name string
	Age  int
}

var _ = Describe("Gob", func() {
	Describe("Marshal", func() {
		It("converts the given data into gob", func() {
			data := &account{Name: "Gura", Age: 9927}
			marshaled, err := codec.Codec.Marshal(data)
			Expect(err).NotTo(HaveOccurred())
			var unmarshaled account
			err = gob.NewDecoder(bytes.NewReader(marshaled)).Decode(&unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(&unmarshaled).To(Equal(data))
		})
	})

	Describe("Unmarshal", func() {
		It("parses gob", func() {
			data := &account{Name: "Gura", Age: 9927}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(data)
			Expect(err).NotTo(HaveOccurred())
			marshaled := buf.Bytes()
			var unmarshaled account
			err = codec.Codec.Unmarshal(marshaled, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(&unmarshaled).To(Equal(data))
		})
	})

	commontest.AssertCodecCanMarshalAndUnmarshalAlmostEverything(codec.Codec)
})
