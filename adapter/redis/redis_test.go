package redis_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"

	"github.com/genkami/kiara/adapter/internal/commontest"
	adapter "github.com/genkami/kiara/adapter/redis"
	"github.com/genkami/kiara/types"
	"github.com/go-redis/redis/v8"
)

var redisAddr string

func getEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		Fail(fmt.Sprintf("environment variable %s not set", name))
	}
	return value
}

var _ = BeforeSuite(func() {
	redisAddr = getEnv("KIARA_TEST_REDIS_ADDR")
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
