package gapdetector

import (
	"math"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "zero values explicitly disable thresholds",
			config: Config{
				MaxTimeGap:        0,
				MaxGroundSpeedMPS: 0,
			},
			wantError: false,
		},
		{
			name: "positive values enable thresholds",
			config: Config{
				MaxTimeGap:        90 * time.Second,
				MaxGroundSpeedMPS: 420,
			},
			wantError: false,
		},
		{
			name: "negative max time gap is invalid",
			config: Config{
				MaxTimeGap:        -time.Second,
				MaxGroundSpeedMPS: 420,
			},
			wantError: true,
		},
		{
			name: "negative max ground speed is invalid",
			config: Config{
				MaxTimeGap:        90 * time.Second,
				MaxGroundSpeedMPS: -1,
			},
			wantError: true,
		},
		{
			name: "not a number max ground speed is invalid",
			config: Config{
				MaxTimeGap:        90 * time.Second,
				MaxGroundSpeedMPS: math.NaN(),
			},
			wantError: true,
		},
		{
			name: "positive infinity max ground speed is invalid",
			config: Config{
				MaxTimeGap:        90 * time.Second,
				MaxGroundSpeedMPS: math.Inf(1),
			},
			wantError: true,
		},
		{
			name: "negative infinity max ground speed is invalid",
			config: Config{
				MaxTimeGap:        90 * time.Second,
				MaxGroundSpeedMPS: math.Inf(-1),
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				err := test.config.Validate()

				if test.wantError && err == nil {
					t.Fatal(
						"expected validation error, got nil",
					)
				}

				if !test.wantError && err != nil {
					t.Fatalf(
						"expected valid configuration, got error: %v",
						err,
					)
				}
			},
		)
	}
}

func TestConfigValidateRejectsNegativeTimeGapWithoutChangingConfig(
	t *testing.T,
) {
	config := Config{
		MaxTimeGap:        -5 * time.Second,
		MaxGroundSpeedMPS: 420,
	}

	err := config.Validate()

	if err == nil {
		t.Fatal(
			"expected validation error, got nil",
		)
	}

	if config.MaxTimeGap != -5*time.Second {
		t.Fatalf(
			"expected max time gap to remain unchanged, got %s",
			config.MaxTimeGap,
		)
	}

	if config.MaxGroundSpeedMPS != 420 {
		t.Fatalf(
			"expected max ground speed to remain unchanged, got %f",
			config.MaxGroundSpeedMPS,
		)
	}
}

func TestConfigValidateRejectsNegativeGroundSpeedWithoutChangingConfig(
	t *testing.T,
) {
	config := Config{
		MaxTimeGap:        90 * time.Second,
		MaxGroundSpeedMPS: -100,
	}

	err := config.Validate()

	if err == nil {
		t.Fatal(
			"expected validation error, got nil",
		)
	}

	if config.MaxTimeGap != 90*time.Second {
		t.Fatalf(
			"expected max time gap to remain unchanged, got %s",
			config.MaxTimeGap,
		)
	}

	if config.MaxGroundSpeedMPS != -100 {
		t.Fatalf(
			"expected max ground speed to remain unchanged, got %f",
			config.MaxGroundSpeedMPS,
		)
	}
}
