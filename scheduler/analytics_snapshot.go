package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/reganputra/skripsi-backend/service"
)

// StartAnalyticsSnapshotScheduler runs a daily job that computes and persists
// the previous calendar day's engagement snapshot into daily_analytics_snapshots.
//
// The job fires at 00:05 UTC each day (5 minutes after midnight gives all
// timestamp-based inserts from the previous day time to flush). After the first
// trigger, it ticks every 24 hours.
//
// The scheduler respects the context and exits cleanly when it is cancelled
// (e.g. during graceful shutdown).
func StartAnalyticsSnapshotScheduler(ctx context.Context, svc service.AnalyticsService) {
	// Calculate time until 00:05 UTC tomorrow.
	now := time.Now().UTC()
	nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, time.UTC)
	initialDelay := time.Until(nextRun)

	log.Printf("[analytics-scheduler] first snapshot run at %s (in %s)",
		nextRun.Format(time.RFC3339), initialDelay.Round(time.Minute))

	timer := time.NewTimer(initialDelay)
	ticker := (*time.Ticker)(nil)

	go func() {
		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				if ticker != nil {
					ticker.Stop()
				}
				log.Println("[analytics-scheduler] stopped")
				return

			case <-timer.C:
				// Switch to a 24-hour ticker after the first fire.
				ticker = time.NewTicker(24 * time.Hour)
				runSnapshot(svc)

			case t := <-func() <-chan time.Time {
				if ticker != nil {
					return ticker.C
				}
				return nil
			}():
				_ = t
				runSnapshot(svc)
			}
		}
	}()
}

func runSnapshot(svc service.AnalyticsService) {
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	log.Printf("[analytics-scheduler] computing snapshot for %s", yesterday.Format("2006-01-02"))
	if err := svc.BuildSnapshotForDate(yesterday); err != nil {
		log.Printf("[analytics-scheduler] error: %v", err)
	} else {
		log.Printf("[analytics-scheduler] snapshot saved for %s", yesterday.Format("2006-01-02"))
	}
}
