package sourcehttp

import "testing"

func TestValidatorHasValidatorsReturnsFalseForEmptyState(
	t *testing.T,
) {
	validator := Validator{}

	if validator.HasValidators() {
		t.Fatal(
			"expected empty validator state to report no validators",
		)
	}
}

func TestValidatorHasValidatorsReturnsFalseForWhitespaceOnlyState(
	t *testing.T,
) {
	validator := Validator{
		ETag:         "   ",
		LastModified: "\t\n",
	}

	if validator.HasValidators() {
		t.Fatal(
			"expected whitespace-only validator state to report no validators",
		)
	}
}

func TestValidatorHasValidatorsReturnsTrueForETag(
	t *testing.T,
) {
	validator := Validator{
		ETag: `"validator-a"`,
	}

	if !validator.HasValidators() {
		t.Fatal(
			"expected ETag validator state to be detected",
		)
	}
}

func TestValidatorHasValidatorsReturnsTrueForLastModified(
	t *testing.T,
) {
	validator := Validator{
		LastModified: "Sun, 05 Jul 2026 01:53:55 GMT",
	}

	if !validator.HasValidators() {
		t.Fatal(
			"expected Last-Modified validator state to be detected",
		)
	}
}
