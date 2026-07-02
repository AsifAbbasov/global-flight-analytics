package handlers

import (
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	trafficquery "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/query"
	"github.com/gofiber/fiber/v2"
)

type TrajectoryHandler struct {
	service *trafficquery.Service
}

func NewTrajectoryHandler(service *trafficquery.Service) *TrajectoryHandler {
	return &TrajectoryHandler{
		service: service,
	}
}

func (handler *TrajectoryHandler) GetLatestByICAO24(c *fiber.Ctx) error {
	if handler.service == nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "TRAJECTORY_SERVICE_UNAVAILABLE", "Trajectory service is unavailable")
	}

	icao24 := c.Params("icao24")

	item, err := handler.service.GetLatestTrajectoryByICAO24(c.Context(), icao24)
	if err != nil {
		if errors.Is(err, trafficquery.ErrInvalidICAO24) {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_ICAO24", "Invalid ICAO24")
		}

		if errors.Is(err, trafficquery.ErrTrajectoryRepositoryRequired) {
			return response.Error(c, fiber.StatusServiceUnavailable, "TRAJECTORY_SERVICE_UNAVAILABLE", "Trajectory service is unavailable")
		}

		if errors.Is(err, postgres.ErrTrajectoryNotFound) {
			return response.Error(c, fiber.StatusNotFound, "TRAJECTORY_NOT_FOUND", "Trajectory not found")
		}

		return response.Error(c, fiber.StatusInternalServerError, "TRAJECTORY_LOAD_FAILED", "Failed to load trajectory")
	}

	return response.OK(c, dto.ToTrajectory(item))
}

func (handler *TrajectoryHandler) GetByID(c *fiber.Ctx) error {
	if handler.service == nil {
		return response.Error(c, fiber.StatusServiceUnavailable, "TRAJECTORY_SERVICE_UNAVAILABLE", "Trajectory service is unavailable")
	}

	trajectoryID := c.Params("id")

	item, err := handler.service.GetTrajectoryByID(c.Context(), trajectoryID)
	if err != nil {
		if errors.Is(err, trafficquery.ErrInvalidTrajectoryID) {
			return response.Error(c, fiber.StatusBadRequest, "INVALID_TRAJECTORY_ID", "Invalid trajectory id")
		}

		if errors.Is(err, trafficquery.ErrTrajectoryRepositoryRequired) {
			return response.Error(c, fiber.StatusServiceUnavailable, "TRAJECTORY_SERVICE_UNAVAILABLE", "Trajectory service is unavailable")
		}

		if errors.Is(err, postgres.ErrTrajectoryNotFound) {
			return response.Error(c, fiber.StatusNotFound, "TRAJECTORY_NOT_FOUND", "Trajectory not found")
		}

		return response.Error(c, fiber.StatusInternalServerError, "TRAJECTORY_LOAD_FAILED", "Failed to load trajectory")
	}

	return response.OK(c, dto.ToTrajectory(item))
}
