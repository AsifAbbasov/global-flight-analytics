package handlers

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/gofiber/fiber/v2"
)

type flightStateAltitudeHTTPTestRepository struct {
	list   []flightstate.FlightState
	latest flightstate.FlightState
}

func (
	repository *flightStateAltitudeHTTPTestRepository,
) ListByFlightID(
	_ context.Context,
	_ string,
) ([]flightstate.FlightState, error) {
	return repository.list,
		nil
}

func (
	repository *flightStateAltitudeHTTPTestRepository,
) GetLatestByICAO24(
	_ context.Context,
	_ string,
) (flightstate.FlightState, error) {
	return repository.latest,
		nil
}

func TestToFlightStateItemPublishesObservedZeroAltitude(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusObserved

	result := toFlightStateItem(
		item,
	)

	assertAltitudeHTTPValue(
		t,
		"barometric altitude",
		result.BarometricAltitudeM,
		result.BarometricAltitudeStatus,
		true,
		0,
		flightstate.AltitudeStatusObserved,
	)
}

func TestToFlightStateItemPublishesGroundAltitude(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusGround
	item.OnGround = true

	result := toFlightStateItem(
		item,
	)

	assertAltitudeHTTPValue(
		t,
		"barometric altitude",
		result.BarometricAltitudeM,
		result.BarometricAltitudeStatus,
		true,
		0,
		flightstate.AltitudeStatusGround,
	)
}

func TestToFlightStateItemRejectsGroundStatusWithoutOnGround(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusGround
	item.OnGround = false

	result := toFlightStateItem(
		item,
	)

	assertAltitudeHTTPValue(
		t,
		"barometric altitude",
		result.BarometricAltitudeM,
		result.BarometricAltitudeStatus,
		false,
		0,
		flightstate.AltitudeStatusInvalid,
	)
}

func TestToFlightStateItemPublishesUnknownAltitudeAsNull(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusUnknown

	result := toFlightStateItem(
		item,
	)

	assertAltitudeHTTPValue(
		t,
		"barometric altitude",
		result.BarometricAltitudeM,
		result.BarometricAltitudeStatus,
		false,
		0,
		flightstate.AltitudeStatusUnknown,
	)
}

func TestToFlightStateItemPublishesUnavailableAltitudeAsNull(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.GeometricAltitudeM = 0
	item.GeometricAltitudeStatus = flightstate.AltitudeStatusUnavailable

	result := toFlightStateItem(
		item,
	)

	assertAltitudeHTTPValue(
		t,
		"geometric altitude",
		result.GeometricAltitudeM,
		result.GeometricAltitudeStatus,
		false,
		0,
		flightstate.AltitudeStatusUnavailable,
	)
}

func TestToFlightStateItemPublishesInvalidAltitudeAsNull(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusInvalid

	result := toFlightStateItem(
		item,
	)

	assertAltitudeHTTPValue(
		t,
		"barometric altitude",
		result.BarometricAltitudeM,
		result.BarometricAltitudeStatus,
		false,
		0,
		flightstate.AltitudeStatusInvalid,
	)
}

func TestToFlightStateItemCanonicalizesLegacyAltitudeStatus(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.BarometricAltitudeM = 1000
	item.BarometricAltitudeStatus = ""
	item.GeometricAltitudeM = 0
	item.GeometricAltitudeStatus = ""

	result := toFlightStateItem(
		item,
	)

	assertAltitudeHTTPValue(
		t,
		"legacy barometric altitude",
		result.BarometricAltitudeM,
		result.BarometricAltitudeStatus,
		true,
		1000,
		flightstate.AltitudeStatusObserved,
	)

	assertAltitudeHTTPValue(
		t,
		"legacy geometric altitude",
		result.GeometricAltitudeM,
		result.GeometricAltitudeStatus,
		false,
		0,
		flightstate.AltitudeStatusUnavailable,
	)
}

func TestToFlightStateItemProtectsPublicContractFromUnsupportedStatus(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.BarometricAltitudeStatus = flightstate.AltitudeStatus(
		"unsupported",
	)

	result := toFlightStateItem(
		item,
	)

	assertAltitudeHTTPValue(
		t,
		"barometric altitude",
		result.BarometricAltitudeM,
		result.BarometricAltitudeStatus,
		false,
		0,
		flightstate.AltitudeStatusInvalid,
	)
}

func TestToFlightStateItemProtectsPublicContractFromNonFiniteObservedAltitude(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.BarometricAltitudeM = math.NaN()
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusObserved

	result := toFlightStateItem(
		item,
	)

	assertAltitudeHTTPValue(
		t,
		"barometric altitude",
		result.BarometricAltitudeM,
		result.BarometricAltitudeStatus,
		false,
		0,
		flightstate.AltitudeStatusInvalid,
	)
}

func TestFlightStateHandlerLatestPublishesExplicitAltitudeJSONSemantics(
	t *testing.T,
) {
	item := altitudeHTTPTestState()
	item.BarometricAltitudeM = 0
	item.BarometricAltitudeStatus = flightstate.AltitudeStatusObserved
	item.GeometricAltitudeM = 0
	item.GeometricAltitudeStatus = flightstate.AltitudeStatusUnknown

	repository := &flightStateAltitudeHTTPTestRepository{
		latest: item,
	}

	app := fiber.New()

	handler := NewFlightStateHandler(
		flightstate.MustNewService(
			repository,
		),
	)

	app.Get(
		"/flight-states/:icao24/latest",
		handler.GetLatestByICAO24,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/flight-states/ABC123/latest",
		nil,
	)

	response, err := app.Test(
		request,
		-1,
	)
	if err != nil {
		t.Fatalf(
			"execute latest flight state request: %v",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			response.StatusCode,
		)
	}

	var envelope struct {
		Success bool                `json:"success"`
		Data    dto.FlightStateItem `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(
		&envelope,
	); err != nil {
		t.Fatalf(
			"decode latest flight state response: %v",
			err,
		)
	}

	if !envelope.Success {
		t.Fatal(
			"expected success response",
		)
	}

	assertAltitudeHTTPValue(
		t,
		"barometric altitude",
		envelope.Data.BarometricAltitudeM,
		envelope.Data.BarometricAltitudeStatus,
		true,
		0,
		flightstate.AltitudeStatusObserved,
	)

	assertAltitudeHTTPValue(
		t,
		"geometric altitude",
		envelope.Data.GeometricAltitudeM,
		envelope.Data.GeometricAltitudeStatus,
		false,
		0,
		flightstate.AltitudeStatusUnknown,
	)
}

func TestFlightStateHandlerListPublishesDistinctAltitudeJSONSemantics(
	t *testing.T,
) {
	baseTime := altitudeHTTPTestState().ObservedAt

	observedZero := altitudeHTTPTestState()
	observedZero.ObservedAt = baseTime
	observedZero.BarometricAltitudeM = 0
	observedZero.BarometricAltitudeStatus = flightstate.AltitudeStatusObserved

	ground := altitudeHTTPTestState()
	ground.ObservedAt = baseTime.Add(
		time.Second,
	)
	ground.BarometricAltitudeM = 0
	ground.BarometricAltitudeStatus = flightstate.AltitudeStatusGround
	ground.OnGround = true

	unknown := altitudeHTTPTestState()
	unknown.ObservedAt = baseTime.Add(
		2 * time.Second,
	)
	unknown.BarometricAltitudeM = 0
	unknown.BarometricAltitudeStatus = flightstate.AltitudeStatusUnknown

	repository := &flightStateAltitudeHTTPTestRepository{
		list: []flightstate.FlightState{
			observedZero,
			ground,
			unknown,
		},
	}

	app := fiber.New()

	handler := NewFlightStateHandler(
		flightstate.MustNewService(
			repository,
		),
	)

	app.Get(
		"/flights/:flightID/states",
		handler.ListByFlightID,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/flights/flight-1/states",
		nil,
	)

	response, err := app.Test(
		request,
		-1,
	)
	if err != nil {
		t.Fatalf(
			"execute flight state list request: %v",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			response.StatusCode,
		)
	}

	var envelope struct {
		Success bool                  `json:"success"`
		Data    []dto.FlightStateItem `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(
		&envelope,
	); err != nil {
		t.Fatalf(
			"decode flight state list response: %v",
			err,
		)
	}

	if !envelope.Success {
		t.Fatal(
			"expected success response",
		)
	}

	if len(envelope.Data) != 3 {
		t.Fatalf(
			"expected 3 flight states, got %d",
			len(envelope.Data),
		)
	}

	assertAltitudeHTTPValue(
		t,
		"observed zero altitude",
		envelope.Data[0].BarometricAltitudeM,
		envelope.Data[0].BarometricAltitudeStatus,
		true,
		0,
		flightstate.AltitudeStatusObserved,
	)

	assertAltitudeHTTPValue(
		t,
		"ground altitude",
		envelope.Data[1].BarometricAltitudeM,
		envelope.Data[1].BarometricAltitudeStatus,
		true,
		0,
		flightstate.AltitudeStatusGround,
	)

	assertAltitudeHTTPValue(
		t,
		"unknown altitude",
		envelope.Data[2].BarometricAltitudeM,
		envelope.Data[2].BarometricAltitudeStatus,
		false,
		0,
		flightstate.AltitudeStatusUnknown,
	)
}

func altitudeHTTPTestState() flightstate.FlightState {
	return flightstate.FlightState{
		ID:                       "state-1",
		FlightID:                 "flight-1",
		AircraftID:               "aircraft-1",
		ICAO24:                   "ABC123",
		Callsign:                 "AHY101",
		Latitude:                 40.4093,
		Longitude:                49.8671,
		BarometricAltitudeM:      1000,
		BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
		GeometricAltitudeM:       1100,
		GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
		VelocityMPS:              220,
		HeadingDegrees:           90,
		VerticalRateMPS:          0,
		OnGround:                 false,
		OriginCountry:            "Azerbaijan",
		ObservedAt: time.Date(
			2026,
			time.July,
			10,
			0,
			0,
			0,
			0,
			time.UTC,
		),
		SourceName: "airplanes.live",
	}
}

func assertAltitudeHTTPValue(
	t *testing.T,
	field string,
	actualValue *float64,
	actualStatus flightstate.AltitudeStatus,
	expectValue bool,
	expectedValue float64,
	expectedStatus flightstate.AltitudeStatus,
) {
	t.Helper()

	if actualStatus != expectedStatus {
		t.Fatalf(
			"%s expected status %q, got %q",
			field,
			expectedStatus,
			actualStatus,
		)
	}

	if !expectValue {
		if actualValue != nil {
			t.Fatalf(
				"%s expected null value, got %v",
				field,
				*actualValue,
			)
		}

		return
	}

	if actualValue == nil {
		t.Fatalf(
			"%s expected numeric value, got null",
			field,
		)
	}

	if *actualValue != expectedValue {
		t.Fatalf(
			"%s expected value %v, got %v",
			field,
			expectedValue,
			*actualValue,
		)
	}
}
