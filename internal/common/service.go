package common

import "github.com/Cork-Holdings/gp_payment_orchestration/internal/global"

type Service struct {
	App *global.App
}

func New(app *global.App) *Service {
	return &Service{App: app}
}
