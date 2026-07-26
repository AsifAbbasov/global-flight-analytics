package featurepipeline

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractorcomposition"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/featurestore"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/validator"
)

const Version = "flight-feature-processing-pipeline-v4"

type Stage string

const (
	StageExtraction Stage = "extraction"
	StageValidation Stage = "validation"
	StageStorage    Stage = "storage"
)

type FeatureExtractor interface {
	Extract(
		ctx context.Context,
		request extractor.Request,
	) (flightfeatures.FlightFeatures, error)
}

type FeatureValidator interface {
	Validate(
		ctx context.Context,
		features flightfeatures.FlightFeatures,
	) (
		flightfeatures.FlightFeatures,
		validator.Report,
		error,
	)
}

type FeatureWriter interface {
	Put(
		ctx context.Context,
		features flightfeatures.FlightFeatures,
	) (featurestore.Record, error)
}

type Config struct {
	Extractor         FeatureExtractor
	Validator         FeatureValidator
	Writer            FeatureWriter
	ProcessingVersion flightfeatures.ProcessingVersion
}

type Result struct {
	PipelineVersion  string
	ValidationReport validator.Report
	Record           featurestore.Record
}

func (result Result) Clone() Result {
	cloned := result
	cloned.ValidationReport =
		result.ValidationReport.Clone()
	cloned.Record = result.Record.Clone()

	return cloned
}

func (result Result) Features() flightfeatures.FlightFeatures {
	return result.Record.Features.Clone()
}

type InMemoryConfig struct {
	Extractor       extractorcomposition.Config
	ValidatorPolicy *validator.Policy
	Now             func() time.Time
}

type Versions struct {
	Composition         string
	Pipeline            string
	ExtractorComponents extractorcomposition.Versions
	Validator           string
	Store               string
}

type InMemoryComposition struct {
	Pipeline             *Pipeline
	Store                *featurestore.MemoryStore
	ExtractorComposition *extractorcomposition.Composition
	Validator            *validator.Validator
	Versions             Versions
}
