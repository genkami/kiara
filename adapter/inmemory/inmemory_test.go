package inmemory_test

import (
	"time"

	. "github.com/onsi/ginkgo"

	"github.com/genkami/kiara/adapter/inmemory"
	"github.com/genkami/kiara/adapter/internal/commontest"
	"github.com/genkami/kiara/types"
)

var (
	timeout      = 1 * time.Second
	shortTimeout = 10 * time.Millisecond
)

type env struct {
	broker *inmemory.Broker
}

func (e *env) Setup() {
	e.broker = inmemory.NewBroker()
}

func (e *env) Teardown() {
	e.broker.Close()
}

func (e *env) NewAdapter() types.Adapter {
	return inmemory.NewAdapter(e.broker)
}

var _ = Describe("Inmemory", func() {
	commontest.AssertAdapterIsImplementedCorrectly(&env{})
})
