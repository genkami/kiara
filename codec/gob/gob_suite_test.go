package gob_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGob(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gob Suite")
}
