package aircraft

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) List(ctx context.Context) ([]Aircraft, error) {
	return s.repository.List(ctx)
}

func (s *Service) GetByICAO24(ctx context.Context, icao24 string) (Aircraft, error) {
	return s.repository.GetByICAO24(ctx, icao24)
}
