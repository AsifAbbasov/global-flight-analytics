package historicalaggregate

func listCursorFromRecord(
	record Record,
) *ListCursor {
	return &ListCursor{
		WindowEnd: record.Key.Window.
			EndTime.UTC(),
		WindowStart: record.Key.Window.
			StartTime.UTC(),
		AsOfTime: record.Key.Window.
			AsOfTime.UTC(),
		ID: record.ID,
	}
}
