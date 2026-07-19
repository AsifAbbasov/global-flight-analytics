package handlers

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/transponderalert"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

type TransponderEvidenceReader interface {
	GetLatest(
		ctx context.Context,
		icao24 string,
	) (transponderalert.LatestEvidence, error)
}

type TransponderEvidenceHandler struct {
	reader TransponderEvidenceReader
}

func NewTransponderEvidenceHandler(
	reader TransponderEvidenceReader,
) *TransponderEvidenceHandler {
	return &TransponderEvidenceHandler{
		reader: reader,
	}
}

func (handler *TransponderEvidenceHandler) GetLatest(
	ctx *fiber.Ctx,
) error {
	if handler == nil || handler.reader == nil {
		return response.Error(
			ctx,
			fiber.StatusServiceUnavailable,
			"TRANSPONDER_EVIDENCE_SERVICE_UNAVAILABLE",
			"Transponder evidence service is unavailable",
		)
	}

	result, err := handler.reader.GetLatest(
		ctx.Context(),
		ctx.Params("icao24"),
	)
	if err != nil {
		return writeTransponderEvidenceError(
			ctx,
			err,
		)
	}

	return response.OK(
		ctx,
		dto.ToTransponderEvidenceResponse(
			result,
		),
	)
}

func writeTransponderEvidenceError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(
		err,
		transponderalert.ErrICAO24Invalid,
	):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_TRANSPONDER_EVIDENCE_ICAO24",
			"ICAO24 must contain exactly six hexadecimal characters",
		)

	case errors.Is(
		err,
		flightstate.ErrNotFound,
	):
		return response.Error(
			ctx,
			fiber.StatusNotFound,
			"TRANSPONDER_EVIDENCE_SOURCE_NOT_FOUND",
			"No persisted flight state was found for the aircraft",
		)

	case errors.Is(
		err,
		transponderalert.ErrEvidenceNotFound,
	):
		return response.Error(
			ctx,
			fiber.StatusNotFound,
			"TRANSPONDER_EVIDENCE_NOT_FOUND",
			"No observed special transponder code evidence was found in the latest persisted flight state",
		)

	case errors.Is(
		err,
		context.DeadlineExceeded,
	):
		return response.Error(
			ctx,
			fiber.StatusGatewayTimeout,
			"TRANSPONDER_EVIDENCE_TIMEOUT",
			"Transponder evidence request timed out",
		)

	case errors.Is(
		err,
		context.Canceled,
	):
		return response.Error(
			ctx,
			fiber.StatusRequestTimeout,
			"TRANSPONDER_EVIDENCE_REQUEST_CANCELED",
			"Transponder evidence request was canceled",
		)

	default:
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"TRANSPONDER_EVIDENCE_LOAD_FAILED",
			"Failed to load transponder evidence",
		)
	}
}
