package ourairports

import "errors"

var ErrCountryCodesRequired = errors.New(
	"OurAirports country codes are required",
)
