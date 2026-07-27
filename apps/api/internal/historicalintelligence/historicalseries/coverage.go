package historicalseries

import (
	"math"
)

func BindDatasetCoverage(
	values []BucketValue,
	evidence DatasetCoverage,
) ([]BucketValue, error) {
	if evidence.MatchedCount < 0 {
		return nil, ErrDatasetCoverageInvalid
	}
	if evidence.State != DatasetReadComplete &&
		evidence.State != DatasetReadIncomplete {
		return nil, ErrDatasetCoverageInvalid
	}
	if evidence.State == DatasetReadIncomplete &&
		evidence.MatchedCount == 0 {
		return nil, ErrDatasetCoverageInvalid
	}

	bound := append([]BucketValue(nil), values...)
	for index := range bound {
		if bound[index].SampleCount < 0 {
			return nil, ErrBucketSampleCountInvalid
		}

		loadedCount := int64(bound[index].SampleCount)
		switch evidence.State {
		case DatasetReadComplete:
			bound[index].Coverage = CoverageEvidence{
				State:        CoverageStateComplete,
				LoadedCount:  loadedCount,
				MatchedCount: loadedCount,
				Ratio:        1,
			}

		case DatasetReadIncomplete:
			if loadedCount == 0 {
				if bound[index].Value != 0 {
					return nil, ErrUnavailableBucketHasData
				}
				bound[index].Coverage = CoverageEvidence{
					State:        CoverageStateUnavailable,
					LoadedCount:  0,
					MatchedCount: evidence.MatchedCount,
					Ratio:        0,
				}
				continue
			}

			denominator := evidence.MatchedCount
			if denominator <= loadedCount {
				if loadedCount == math.MaxInt64 {
					return nil, ErrDatasetCoverageInvalid
				}
				denominator = loadedCount + 1
			}

			bound[index].Coverage = CoverageEvidence{
				State:        CoverageStatePartial,
				LoadedCount:  loadedCount,
				MatchedCount: denominator,
				Ratio: float64(loadedCount) /
					float64(denominator),
			}
		}
	}

	return bound, nil
}

func validateCoverageEvidence(
	value BucketValue,
) error {
	if value.SampleCount < 0 {
		return ErrBucketSampleCountInvalid
	}
	if math.IsNaN(value.Value) ||
		math.IsInf(value.Value, 0) ||
		value.Value < 0 {
		return ErrBucketValueInvalid
	}
	if value.Coverage.LoadedCount !=
		int64(value.SampleCount) {
		return ErrCoverageEvidenceInvalid
	}
	if value.Coverage.MatchedCount < 0 ||
		math.IsNaN(value.Coverage.Ratio) ||
		math.IsInf(value.Coverage.Ratio, 0) ||
		value.Coverage.Ratio < 0 ||
		value.Coverage.Ratio > 1 {
		return ErrCoverageEvidenceInvalid
	}

	switch value.Coverage.State {
	case CoverageStateUnavailable:
		if value.Coverage.LoadedCount != 0 ||
			value.Coverage.Ratio != 0 ||
			value.Value != 0 ||
			value.SampleCount != 0 {
			return ErrUnavailableBucketHasData
		}

	case CoverageStatePartial:
		if value.Coverage.LoadedCount <= 0 ||
			value.Coverage.MatchedCount <=
				value.Coverage.LoadedCount ||
			value.Coverage.Ratio <= 0 ||
			value.Coverage.Ratio >= 1 {
			return ErrCoverageEvidenceInvalid
		}

	case CoverageStateComplete:
		if value.Coverage.Ratio != 1 ||
			value.Coverage.MatchedCount !=
				value.Coverage.LoadedCount {
			return ErrCoverageEvidenceInvalid
		}

	default:
		return ErrCoverageEvidenceInvalid
	}

	return nil
}
