package tasks

import (
	"os"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/hibiken/asynq"
)

func Run(app *global.App) (*asynq.Server, error) {
	mux := asynq.NewServeMux()

	var srv = asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     os.Getenv("REDIS_URL"),
			Password: os.Getenv("REDIS_PASSWORD"),
			Username: os.Getenv("REDIS_USERNAME"),
		},
		asynq.Config{
			Concurrency: 10,
		},
	)

	if err := srv.Run(mux); err != nil {
		return nil, err
	}
	return srv, nil
}
