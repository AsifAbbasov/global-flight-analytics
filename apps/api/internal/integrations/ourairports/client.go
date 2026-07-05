package ourairports

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

const AirportsCSVURL = "https://davidmegginson.github.io/ourairports-data/airports.csv"

var ErrClientTimeoutRequired = errors.New(
	"OurAirports client timeout must be greater than zero",
)

type ClientConfig struct {
	Timeout        time.Duration
	AirportsCSVURL string
	CountryCodes   []string
	Now            func() time.Time
}

type Client struct {
	httpClient     *http.Client
	airportsCSVURL string
	countryCodes   []string
	now            func() time.Time
}

type ConditionalRequest struct {
	ETag         string
	LastModified string
}

type LoadResult struct {
	Airports []airport.ImportRecord

	RetrievedAt time.Time
	CheckedAt   time.Time

	ETag         string
	LastModified string

	NotModified bool
}

func NewClient(
	config ClientConfig,
) (*Client, error) {
	if config.Timeout <= 0 {
		return nil, ErrClientTimeoutRequired
	}

	airportsCSVURL := strings.TrimSpace(
		config.AirportsCSVURL,
	)
	if airportsCSVURL == "" {
		airportsCSVURL = AirportsCSVURL
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	countryCodes := append(
		[]string(nil),
		config.CountryCodes...,
	)

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		airportsCSVURL: airportsCSVURL,
		countryCodes:   countryCodes,
		now:            now,
	}, nil
}

func (client *Client) LoadAirports(
	ctx context.Context,
) (LoadResult, error) {
	return client.LoadAirportsConditional(
		ctx,
		ConditionalRequest{},
	)
}

func (client *Client) LoadAirportsConditional(
	ctx context.Context,
	conditionalRequest ConditionalRequest,
) (LoadResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		client.airportsCSVURL,
		nil,
	)
	if err != nil {
		return LoadResult{}, fmt.Errorf(
			"create OurAirports request: %w",
			err,
		)
	}

	request.Header.Set(
		"User-Agent",
		"global-flight-analytics-airports-import",
	)

	requestETag := strings.TrimSpace(
		conditionalRequest.ETag,
	)
	if requestETag != "" {
		request.Header.Set(
			"If-None-Match",
			requestETag,
		)
	}

	requestLastModified := strings.TrimSpace(
		conditionalRequest.LastModified,
	)
	if requestLastModified != "" {
		request.Header.Set(
			"If-Modified-Since",
			requestLastModified,
		)
	}

	response, err := client.httpClient.Do(
		request,
	)
	if err != nil {
		return LoadResult{}, fmt.Errorf(
			"download OurAirports airports CSV: %w",
			err,
		)
	}
	defer response.Body.Close()

	checkedAt := client.now().UTC()

	responseETag := strings.TrimSpace(
		response.Header.Get(
			"ETag",
		),
	)

	responseLastModified := strings.TrimSpace(
		response.Header.Get(
			"Last-Modified",
		),
	)

	if response.StatusCode == http.StatusNotModified {
		return LoadResult{
			CheckedAt: checkedAt,
			ETag: firstNonEmpty(
				responseETag,
				requestETag,
			),
			LastModified: firstNonEmpty(
				responseLastModified,
				requestLastModified,
			),
			NotModified: true,
		}, nil
	}

	if response.StatusCode != http.StatusOK {
		return LoadResult{}, fmt.Errorf(
			"download OurAirports airports CSV: unexpected HTTP status %s",
			response.Status,
		)
	}

	items, err := client.parseAirports(
		response,
		checkedAt,
	)
	if err != nil {
		return LoadResult{}, fmt.Errorf(
			"parse OurAirports airports CSV: %w",
			err,
		)
	}

	return LoadResult{
		Airports:     items,
		RetrievedAt:  checkedAt,
		CheckedAt:    checkedAt,
		ETag:         responseETag,
		LastModified: responseLastModified,
		NotModified:  false,
	}, nil
}

func (client *Client) parseAirports(
	response *http.Response,
	retrievedAt time.Time,
) ([]airport.ImportRecord, error) {
	if len(client.countryCodes) == 0 {
		return ParseAirportsCSV(
			response.Body,
			retrievedAt,
		)
	}

	return ParseAirportsCSVForCountryCodes(
		response.Body,
		retrievedAt,
		client.countryCodes,
	)
}

func firstNonEmpty(
	values ...string,
) string {
	for _, value := range values {
		trimmedValue := strings.TrimSpace(
			value,
		)

		if trimmedValue != "" {
			return trimmedValue
		}
	}

	return ""
}
