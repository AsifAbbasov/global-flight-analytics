package handlers

import (
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

type RegionHandler struct {
	service *region.Service
}

func NewRegionHandler(service *region.Service) *RegionHandler {
	return &RegionHandler{
		service: service,
	}
}

func (h *RegionHandler) List(c *fiber.Ctx) error {
	return response.OK(c, toRegionItems(h.service.List()))
}

func (h *RegionHandler) GetByCode(c *fiber.Ctx) error {
	code := c.Params("code")

	item, err := h.service.GetByCode(code)
	if err != nil {
		if errors.Is(err, region.ErrRegionNotFound) {
			return response.Error(c, fiber.StatusNotFound, "REGION_NOT_FOUND", "Region not found")
		}

		return response.Error(c, fiber.StatusInternalServerError, "REGION_LOAD_FAILED", "Failed to load region")
	}

	return response.OK(c, toRegionItem(item))
}

func toRegionItems(items []region.Region) []dto.RegionItem {
	result := make([]dto.RegionItem, 0, len(items))

	for _, item := range items {
		result = append(result, toRegionItem(item))
	}

	return result
}

func toRegionItem(item region.Region) dto.RegionItem {
	return dto.RegionItem{
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		Bounds: dto.RegionBounds{
			MinLatitude:  item.Bounds.MinLatitude,
			MaxLatitude:  item.Bounds.MaxLatitude,
			MinLongitude: item.Bounds.MinLongitude,
			MaxLongitude: item.Bounds.MaxLongitude,
		},
	}
}
