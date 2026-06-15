package global

import (
	"os"
	"sync"

	"github.com/hibiken/asynq"
)

var (
	nce    sync.Once
	asynqC *asynq.Client
)

var redisAddr = os.Getenv("REDIS_URL")
func GetTaskQueue() *asynq.Client {
	nce.Do(func() {
		asynqC = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, 
			Password: os.Getenv("REDIS_PASSWORD"), Username: os.Getenv("REDIS_USERNAME")})
	})
	return asynqC
}
