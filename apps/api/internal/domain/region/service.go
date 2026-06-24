package region

import "errors"

var ErrRegionNotFound = errors.New("region not found")

type Service struct {
	regions []Region
}

func NewService() *Service {
	return &Service{
		regions: []Region{
			{
				Code:        "world",
				Name:        "World",
				Description: "Global traffic region.",
				Bounds: Bounds{
					MinLatitude:  -90,
					MaxLatitude:  90,
					MinLongitude: -180,
					MaxLongitude: 180,
				},
			},
			{
				Code:        "caucasus",
				Name:        "Caucasus",
				Description: "Caucasus region including Azerbaijan, Georgia and Armenia.",
				Bounds: Bounds{
					MinLatitude:  38,
					MaxLatitude:  44,
					MinLongitude: 38,
					MaxLongitude: 51,
				},
			},
			{
				Code:        "cis",
				Name:        "CIS",
				Description: "Commonwealth of Independent States region.",
				Bounds: Bounds{
					MinLatitude:  35,
					MaxLatitude:  82,
					MinLongitude: 19,
					MaxLongitude: 180,
				},
			},
			{
				Code:        "ukraine",
				Name:        "Ukraine",
				Description: "Ukraine regional traffic area.",
				Bounds: Bounds{
					MinLatitude:  44,
					MaxLatitude:  53,
					MinLongitude: 22,
					MaxLongitude: 41,
				},
			},
			{
				Code:        "turkey",
				Name:        "Turkey",
				Description: "Turkey regional traffic area.",
				Bounds: Bounds{
					MinLatitude:  35,
					MaxLatitude:  43,
					MinLongitude: 25,
					MaxLongitude: 45,
				},
			},
		},
	}
}

func (s *Service) List() []Region {
	return s.regions
}

func (s *Service) GetByCode(code string) (Region, error) {
	for _, item := range s.regions {
		if item.Code == code {
			return item, nil
		}
	}

	return Region{}, ErrRegionNotFound
}
