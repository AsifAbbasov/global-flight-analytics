package historicalaggregate

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregatecontract"
)

const Version = historicalaggregatecontract.Version

const (
	DefaultListLimit = historicalaggregatecontract.DefaultListLimit
	MaximumListLimit = historicalaggregatecontract.MaximumListLimit
)

type ResultKey = historicalaggregatecontract.ResultKey
type Record = historicalaggregatecontract.Record
type ListQuery = historicalaggregatecontract.ListQuery
type Page = historicalaggregatecontract.Page
type Reader = historicalaggregatecontract.Reader
type Store = historicalaggregatecontract.Store
