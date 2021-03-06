package redis_test

import (
	"github.com/go-redis/redis/v8"
	. "github.com/onsi/ginkgo"

	"github.com/genkami/kiara/adapter/internal/commontest"
	adapter "github.com/genkami/kiara/adapter/redis"
	"github.com/genkami/kiara/types"
)

var redisAddr string

var _ = BeforeSuite(func() {
	redisAddr = commontest.GetEnv("KIARA_TEST_REDIS_ADDR")
})

type env struct{}

func (e *env) Setup() {
}

func (e *env) Teardown() {
}

func (e *env) NewAdapter() types.Adapter {
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	return adapter.NewAdapter(redisClient)
}

var _ = Describe("Redis", func() {
	commontest.AssertAdapterIsImplementedCorrectly(&env{})
})
