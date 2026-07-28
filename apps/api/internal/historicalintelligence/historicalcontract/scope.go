package historicalcontract

// Equal compares the complete domain identity of two Historical Intelligence
// scopes. New scope fields must be considered here deliberately instead of
// changing comparison semantics implicitly through reflection.
func (scope Scope) Equal(other Scope) bool {
	return scope.Type == other.Type &&
		scope.RegionCode == other.RegionCode &&
		scope.AirportICAOCode == other.AirportICAOCode &&
		scope.OriginICAOCode == other.OriginICAOCode &&
		scope.DestinationICAOCode ==
			other.DestinationICAOCode
}
