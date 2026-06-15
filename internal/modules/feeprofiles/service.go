package feeprofiles

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

func (s *Service) CreateFeeProfile(m *FeeProfile) (*FeeProfile, error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return repo.CreateOne[FeeProfile](s.app, *m)
}

func (s *Service) GetFeeProfile(id string) (*FeeProfile, error) {
	return repo.GetOne[FeeProfile](s.app, map[string]any{"id": id})
}

func (s *Service) GetFeeProfiles(opts *repo.Options) ([]FeeProfile, error) {
	return repo.GetMany[FeeProfile](s.app, opts)
}

func (s *Service) UpdateFeeProfile(id string, updates any) (*FeeProfile, error) {
	return repo.UpdateOne[FeeProfile](s.app, map[string]any{"id": id}, updates)
}

func (s *Service) DeleteFeeProfile(id string) error {
	return repo.DeleteOne[FeeProfile](s.app, map[string]any{"id": id})
}
