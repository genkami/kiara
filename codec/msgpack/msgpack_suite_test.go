package msgpack_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMsgpack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Msgpack Suite")
}
