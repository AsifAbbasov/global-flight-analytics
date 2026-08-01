package transponderalert

import "errors"

var ErrLatestEvidenceContextRequired = errors.New(
	"transponder latest evidence context is required",
)
