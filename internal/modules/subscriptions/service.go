package subscriptions

import (
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/repo"
	"github.com/google/uuid"
)

type Service struct {
	app *global.App
}

func NewService(app *global.App) *Service {
	return &Service{app: app}
}

func (s *Service) CreateSubscription(m *Subscription) (*Subscription, error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return repo.CreateOne[Subscription](s.app, *m)
}

func (s *Service) GetSubscription(id string) (*Subscription, error) {
	return repo.GetOne[Subscription](s.app, map[string]any{"id": id})
}

func (s *Service) GetSubscriptions(opts *repo.Options) ([]Subscription, error) {
	return repo.GetMany[Subscription](s.app, opts)
}

func (s *Service) UpdateSubscription(id string, updates any) (*Subscription, error) {
	return repo.UpdateOne[Subscription](s.app, map[string]any{"id": id}, updates)
}

func (s *Service) DeleteSubscription(id string) error {
	return repo.DeleteOne[Subscription](s.app, map[string]any{"id": id})
}
