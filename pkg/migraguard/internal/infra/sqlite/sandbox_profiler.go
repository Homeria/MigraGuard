package sqlite

import (
	"math"
	"math/rand"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// WorkloadProfiler is the interface for generating realistic traffic patterns.
type WorkloadProfiler interface {
	CalculateTPS(t time.Time, profile types.SQLiteHistoryProfile) float64
}

// DefaultWorkloadProfiler implements a rich daily/weekly pattern with sine waves and noise.
type DefaultWorkloadProfiler struct{}

func (p *DefaultWorkloadProfiler) CalculateTPS(t time.Time, profile types.SQLiteHistoryProfile) float64 {
	// A. Time-based normalization (0.0 - 1.0)
	hour := float64(t.Hour()) + float64(t.Minute())/60.0
	
	// 1. Daily Deterministic Fluctuation Factors (Seed based on Year-Month-Day)
	shift := 0.0
	volumeScale := 1.0
	skew := profile.AsymmetricSkew

	if profile.PeakShiftHours > 0 || profile.AsymmetricSkew != 0 {
		daySeed := int64(t.Year()*10000 + int(t.Month())*100 + t.Day())
		r := rand.New(rand.NewSource(daySeed))
		
		// Peak Shift Offset
		if profile.PeakShiftHours > 0 {
			shift = (r.Float64()*2.0 - 1.0) * profile.PeakShiftHours
		}
		
		// Daily Volume Scale: Modulates peak amplitude by +-15% per day
		volumeScale = 0.85 + r.Float64()*0.30 // [0.85, 1.15]
		
		// Daily Skewness Modulation: Modulates skewness by +-20% per day
		if profile.AsymmetricSkew != 0 {
			skew = profile.AsymmetricSkew * (0.8 + r.Float64()*0.4) // [0.8 * skew, 1.2 * skew]
		}
	}

	skewedHour := hour - shift
	for skewedHour < 0 {
		skewedHour += 24.0
	}
	for skewedHour >= 24.0 {
		skewedHour -= 24.0
	}

	// 2. Asymmetric Load Curve (Time Warping) with modulated daily skew
	warpedHour := skewedHour
	if skew != 0 {
		theta := skewedHour * math.Pi / 12.0
		thetaSkewed := theta + skew*math.Sin(theta)
		warpedHour = thetaSkewed * 12.0 / math.Pi
		for warpedHour < 0 {
			warpedHour += 24.0
		}
		for warpedHour >= 24.0 {
			warpedHour -= 24.0
		}
	}

	// 3. Daily Sine Wave (Peak at 14:00 and 20:00, modified by shifted & warped time)
	dailyPattern := 0.4*math.Sin((warpedHour-9)*math.Pi/12.0) + 0.3*math.Sin((warpedHour-18)*math.Pi/6.0) + 0.5
	
	// 4. Weekly Pattern (Weekend traffic is 40% lower)
	weeklyMult := 1.0
	if profile.WeeklyPattern && (t.Weekday() == time.Saturday || t.Weekday() == time.Sunday) {
		weeklyMult = 0.6
	}

	// 5. Special Events (Flash Sales, Maintenance)
	eventMult := 1.0
	for _, event := range profile.Events {
		eventStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).
			AddDate(0, 0, -profile.Days+event.StartDay).
			Add(time.Duration(event.StartHour) * time.Hour)
		eventEnd := eventStart.Add(time.Duration(event.DurationH) * time.Hour)

		if t.After(eventStart) && t.Before(eventEnd) {
			eventMult = event.Multiplier
			break
		}
	}

	// 6. Combined Calculation with daily volumeScale
	finalTPS := (profile.BaseTPS + dailyPattern*(profile.PeakTPS-profile.BaseTPS)) * weeklyMult * eventMult * volumeScale
	
	// 7. Gaussian-like Noise
	noise := (rand.Float64()*2 - 1) * profile.NoiseVariance * finalTPS
	
	return math.Max(1.0, finalTPS + noise)
}
