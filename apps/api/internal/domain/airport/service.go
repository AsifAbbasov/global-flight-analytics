package airport

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) List(ctx context.Context) ([]Airport, error) {
	return s.repository.List(ctx)
}

func (s *Service) GetByICAO(ctx context.Context, icao string) (Airport, error) {
	return s.repository.GetByICAO(ctx, icao)
}
