package handlers

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

const (
	historicalMetricQuery          = "metric"
	historicalScopeQuery           = "scope"
	historicalGranularityQuery     = "granularity"
	historicalRegionCodeQuery      = "region_code"
	historicalAirportICAOQuery     = "airport_icao"
	historicalOriginICAOQuery      = "origin_icao"
	historicalDestinationICAOQuery = "destination_icao"
	historicalLimitQuery           = "limit"
	historicalBeforeWindowEndQuery = "before_window_end"
)

var (
	errHistoricalMetricInvalid = errors.New(
		"historical metric is required and must be supported",
	)
	errHistoricalScopeInvalid = errors.New(
		"historical scope is required and must be valid",
	)
	errHistoricalGranularityInvalid = errors.New(
		"historical granularity is required and must be supported",
	)
	errHistoricalLimitInvalid = errors.New(
		"historical history limit must be valid",
	)
	errHistoricalCursorInvalid = errors.New(
		"historical history cursor must be valid",
	)

	historicalAirportICAOPattern = regexp.MustCompile(
		`^[A-Z0-9]{4}$`,
	)
)

type historicalIntelligenceStore interface {
	GetLatest(
		context.Context,
		historicalaggregate.ListQuery,
	) (historicalaggregate.Record, error)
	List(
		context.Context,
		historicalaggregate.ListQuery,
	) (historicalaggregate.Page, error)
}

type HistoricalIntelligenceHandler struct {
	store historicalIntelligenceStore
}

func NewHistoricalIntelligenceHandler(
	store historicalIntelligenceStore,
) *HistoricalIntelligenceHandler {
	return &HistoricalIntelligenceHandler{
		store: store,
	}
}

func (handler *HistoricalIntelligenceHandler) GetLatest(
	ctx *fiber.Ctx,
) error {
	if handler.store == nil {
		return historicalIntelligenceUnavailable(
			ctx,
		)
	}

	query, err := parseHistoricalIntelligenceQuery(
		historicalIntelligenceQueryValues{
			Metric: ctx.Query(
				historicalMetricQuery,
			),
			Scope: ctx.Query(
				historicalScopeQuery,
			),
			Granularity: ctx.Query(
				historicalGranularityQuery,
			),
			RegionCode: ctx.Query(
				historicalRegionCodeQuery,
			),
			AirportICAO: ctx.Query(
				historicalAirportICAOQuery,
			),
			OriginICAO: ctx.Query(
				historicalOriginICAOQuery,
			),
			DestinationICAO: ctx.Query(
				historicalDestinationICAOQuery,
			),
		},
		false,
	)
	if err != nil {
		return historicalIntelligenceRequestError(
			ctx,
			err,
		)
	}

	record, err := handler.store.GetLatest(
		ctx.Context(),
		query,
	)
	if err != nil {
		return writeHistoricalIntelligenceError(
			ctx,
			err,
			"HISTORICAL_INTELLIGENCE_LOAD_FAILED",
			"Failed to load the latest Historical Intelligence aggregate",
		)
	}

	return response.OK(
		ctx,
		dto.ToHistoricalIntelligenceAggregateRecord(
			record,
		),
	)
}

func (handler *HistoricalIntelligenceHandler) ListHistory(
	ctx *fiber.Ctx,
) error {
	if handler.store == nil {
		return historicalIntelligenceUnavailable(
			ctx,
		)
	}

	query, err := parseHistoricalIntelligenceQuery(
		historicalIntelligenceQueryValues{
			Metric: ctx.Query(
				historicalMetricQuery,
			),
			Scope: ctx.Query(
				historicalScopeQuery,
			),
			Granularity: ctx.Query(
				historicalGranularityQuery,
			),
			RegionCode: ctx.Query(
				historicalRegionCodeQuery,
			),
			AirportICAO: ctx.Query(
				historicalAirportICAOQuery,
			),
			OriginICAO: ctx.Query(
				historicalOriginICAOQuery,
			),
			DestinationICAO: ctx.Query(
				historicalDestinationICAOQuery,
			),
			Limit: ctx.Query(
				historicalLimitQuery,
			),
			BeforeWindowEnd: ctx.Query(
				historicalBeforeWindowEndQuery,
			),
		},
		true,
	)
	if err != nil {
		return historicalIntelligenceRequestError(
			ctx,
			err,
		)
	}

	page, err := handler.store.List(
		ctx.Context(),
		query,
	)
	if err != nil {
		return writeHistoricalIntelligenceError(
			ctx,
			err,
			"HISTORICAL_INTELLIGENCE_HISTORY_LOAD_FAILED",
			"Failed to load Historical Intelligence aggregate history",
		)
	}

	return response.OK(
		ctx,
		dto.ToHistoricalIntelligenceAggregateHistory(
			page,
		),
	)
}

type historicalIntelligenceQueryValues struct {
	Metric          string
	Scope           string
	Granularity     string
	RegionCode      string
	AirportICAO     string
	OriginICAO      string
	DestinationICAO string

	Limit           string
	BeforeWindowEnd string
}

func parseHistoricalIntelligenceQuery(
	values historicalIntelligenceQueryValues,
	includePagination bool,
) (historicalaggregate.ListQuery, error) {
	metricName, err := parseHistoricalMetric(
		values.Metric,
	)
	if err != nil {
		return historicalaggregate.ListQuery{},
			err
	}

	scope, err := parseHistoricalScope(
		values,
	)
	if err != nil {
		return historicalaggregate.ListQuery{},
			err
	}

	granularity, err := parseHistoricalGranularity(
		values.Granularity,
	)
	if err != nil {
		return historicalaggregate.ListQuery{},
			err
	}

	query := historicalaggregate.ListQuery{
		SchemaVersion: historicalcontract.SchemaVersionV1,
		MetricName:    metricName,
		Scope:         scope,
		Granularity:   granularity,
	}

	if !includePagination {
		return query, nil
	}

	limit, err := parseHistoricalLimit(
		values.Limit,
	)
	if err != nil {
		return historicalaggregate.ListQuery{},
			err
	}
	beforeWindowEnd, err :=
		parseHistoricalBeforeWindowEnd(
			values.BeforeWindowEnd,
		)
	if err != nil {
		return historicalaggregate.ListQuery{},
			err
	}

	query.Limit = limit
	query.BeforeWindowEnd = beforeWindowEnd

	return query, nil
}

func parseHistoricalMetric(
	value string,
) (historicalcontract.MetricName, error) {
	normalized := historicalcontract.MetricName(
		strings.ToLower(
			strings.TrimSpace(value),
		),
	)
	for _, supported := range historicalcontract.SupportedMetricNames() {
		if normalized == supported {
			return normalized, nil
		}
	}

	return "", errHistoricalMetricInvalid
}

func parseHistoricalScope(
	values historicalIntelligenceQueryValues,
) (historicalcontract.Scope, error) {
	scopeType := historicalcontract.ScopeType(
		strings.ToLower(
			strings.TrimSpace(values.Scope),
		),
	)
	regionCode := strings.ToUpper(
		strings.TrimSpace(values.RegionCode),
	)
	airportICAO := strings.ToUpper(
		strings.TrimSpace(values.AirportICAO),
	)
	originICAO := strings.ToUpper(
		strings.TrimSpace(values.OriginICAO),
	)
	destinationICAO := strings.ToUpper(
		strings.TrimSpace(
			values.DestinationICAO,
		),
	)

	switch scopeType {
	case historicalcontract.ScopeTypeGlobal:
		if regionCode != "" ||
			airportICAO != "" ||
			originICAO != "" ||
			destinationICAO != "" {
			return historicalcontract.Scope{},
				errHistoricalScopeInvalid
		}

	case historicalcontract.ScopeTypeRegion:
		if regionCode == "" ||
			airportICAO != "" ||
			originICAO != "" ||
			destinationICAO != "" {
			return historicalcontract.Scope{},
				errHistoricalScopeInvalid
		}

	case historicalcontract.ScopeTypeAirport:
		if !historicalAirportICAOPattern.
			MatchString(airportICAO) ||
			regionCode != "" ||
			originICAO != "" ||
			destinationICAO != "" {
			return historicalcontract.Scope{},
				errHistoricalScopeInvalid
		}

	case historicalcontract.ScopeTypeRoute:
		if !historicalAirportICAOPattern.
			MatchString(originICAO) ||
			!historicalAirportICAOPattern.
				MatchString(destinationICAO) ||
			regionCode != "" ||
			airportICAO != "" {
			return historicalcontract.Scope{},
				errHistoricalScopeInvalid
		}

	default:
		return historicalcontract.Scope{},
			errHistoricalScopeInvalid
	}

	return historicalcontract.Scope{
		Type:                scopeType,
		RegionCode:          regionCode,
		AirportICAOCode:     airportICAO,
		OriginICAOCode:      originICAO,
		DestinationICAOCode: destinationICAO,
	}, nil
}

func parseHistoricalGranularity(
	value string,
) (historicalcontract.Granularity, error) {
	normalized := historicalcontract.Granularity(
		strings.ToLower(
			strings.TrimSpace(value),
		),
	)

	switch normalized {
	case historicalcontract.GranularityHour,
		historicalcontract.GranularityDay,
		historicalcontract.GranularityWeek,
		historicalcontract.GranularityCustom:
		return normalized, nil
	default:
		return "",
			errHistoricalGranularityInvalid
	}
}

func parseHistoricalLimit(
	value string,
) (int, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return historicalaggregate.
			DefaultListLimit, nil
	}

	limit, err := strconv.Atoi(normalized)
	if err != nil ||
		limit < 1 ||
		limit > historicalaggregate.
			MaximumListLimit {
		return 0, errHistoricalLimitInvalid
	}

	return limit, nil
}

func parseHistoricalBeforeWindowEnd(
	value string,
) (time.Time, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse(
		time.RFC3339Nano,
		normalized,
	)
	if err != nil {
		return time.Time{},
			errHistoricalCursorInvalid
	}

	return parsed.UTC(), nil
}

func historicalIntelligenceRequestError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(
		err,
		errHistoricalMetricInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_HISTORICAL_INTELLIGENCE_METRIC",
			"Historical Intelligence metric is required and must be supported",
		)

	case errors.Is(
		err,
		errHistoricalScopeInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_HISTORICAL_INTELLIGENCE_SCOPE",
			"Historical Intelligence scope and its identifying parameters are invalid",
		)

	case errors.Is(
		err,
		errHistoricalGranularityInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_HISTORICAL_INTELLIGENCE_GRANULARITY",
			"Historical Intelligence granularity must be hour, day, week, or custom",
		)

	case errors.Is(
		err,
		errHistoricalLimitInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_HISTORICAL_INTELLIGENCE_LIMIT",
			"Historical Intelligence history limit must be between one and one hundred",
		)

	case errors.Is(
		err,
		errHistoricalCursorInvalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_HISTORICAL_INTELLIGENCE_CURSOR",
			"Historical Intelligence history cursor must be a valid RFC 3339 timestamp",
		)

	default:
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_HISTORICAL_INTELLIGENCE_REQUEST",
			"Historical Intelligence request is invalid",
		)
	}
}

func historicalIntelligenceUnavailable(
	ctx *fiber.Ctx,
) error {
	return response.Error(
		ctx,
		fiber.StatusServiceUnavailable,
		"HISTORICAL_INTELLIGENCE_SERVICE_UNAVAILABLE",
		"Historical Intelligence service is unavailable",
	)
}

func writeHistoricalIntelligenceError(
	ctx *fiber.Ctx,
	err error,
	defaultCode string,
	defaultMessage string,
) error {
	switch {
	case errors.Is(
		err,
		context.DeadlineExceeded,
	):
		return response.Error(
			ctx,
			fiber.StatusGatewayTimeout,
			"HISTORICAL_INTELLIGENCE_TIMEOUT",
			"Historical Intelligence request timed out",
		)

	case errors.Is(
		err,
		context.Canceled,
	):
		return response.Error(
			ctx,
			fiber.StatusRequestTimeout,
			"HISTORICAL_INTELLIGENCE_REQUEST_CANCELED",
			"Historical Intelligence request was canceled",
		)

	case errors.Is(
		err,
		pgx.ErrNoRows,
	),
		errors.Is(
			err,
			historicalaggregate.ErrResultNotFound,
		):
		return response.Error(
			ctx,
			fiber.StatusNotFound,
			"HISTORICAL_INTELLIGENCE_NOT_FOUND",
			"No matching Historical Intelligence aggregate was found",
		)

	case errors.Is(
		err,
		historicalaggregate.ErrScopeInvalid,
	):
		return historicalIntelligenceRequestError(
			ctx,
			errHistoricalScopeInvalid,
		)

	case errors.Is(
		err,
		historicalaggregate.ErrInvalidListLimit,
	):
		return historicalIntelligenceRequestError(
			ctx,
			errHistoricalLimitInvalid,
		)

	case errors.Is(
		err,
		historicalaggregate.
			ErrPostgresPoolRequired,
	),
		errors.Is(
			err,
			historicalaggregate.
				ErrPostgresExecutorRequired,
		):
		return historicalIntelligenceUnavailable(
			ctx,
		)

	default:
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			defaultCode,
			defaultMessage,
		)
	}
}
