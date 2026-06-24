package traffic

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetCurrent(ctx context.Context) ([]CurrentTrafficItem, error) {
	return s.repository.GetCurrent(ctx)
}
