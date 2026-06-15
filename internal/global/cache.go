package global

import (
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	once   sync.Once
)

func GetCache() *redis.Client {
	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr:        os.Getenv("REDIS_URL"),
			Password:    os.Getenv("REDIS_PASSWORD"),
			Username: os.Getenv("REDIS_USERNAME"),
			DB:          0,
			DialTimeout: time.Second * 30,
		})
	})
	return client
}
