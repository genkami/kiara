package kiara_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKiara(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kiara Suite")
}
