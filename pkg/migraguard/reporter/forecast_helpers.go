package reporter

import "github.com/Homeria/MigraGuard/pkg/migraguard/types"

func forecastBestSlot(f *types.ForecastReport) (types.ForecastTimeSlot, bool) {
	if f == nil || f.BestHour < 0 {
		return types.ForecastTimeSlot{}, false
	}
	for _, slot := range f.Timeline {
		if slot.Hour == f.BestHour {
			return slot, true
		}
	}
	return types.ForecastTimeSlot{}, false
}

func findForecastForTable(forecasts []*types.ForecastReport, tableName string) *types.ForecastReport {
	for _, f := range forecasts {
		if f.TableName == tableName {
			return f
		}
	}
	return nil
}
