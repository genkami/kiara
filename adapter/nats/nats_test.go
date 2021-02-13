package nats_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/nats-io/nats.go"

	"github.com/genkami/kiara/adapter/internal/commontest"
	adapter "github.com/genkami/kiara/adapter/nats"
	"github.com/genkami/kiara/types"
)

var natsUrl string

var _ = BeforeSuite(func() {
	natsUrl = commontest.GetEnv("KIARA_TEST_NATS_URL")
})

type env struct{}

func (e *env) Setup() {
}

func (e *env) Teardown() {
}

func (e *env) NewAdapter() types.Adapter {
	conn, err := nats.Connect(natsUrl)
	Expect(err).NotTo(HaveOccurred())
	return adapter.NewAdapter(conn)
}

var _ = Describe("Nats", func() {
	commontest.AssertAdapterIsImplementedCorrectly(&env{})
})
