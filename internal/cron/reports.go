package cron

import (
	"log/slog"

	"github.com/go-co-op/gocron"
)

// RegisterReportJobs schedules reporting-related jobs.
func RegisterReportJobs(scheduler *gocron.Scheduler) {
	_, err := scheduler.Every(15).Seconds().Do(func() {
		slog.Info("Generating weekly report")
		// Add your reporting logic
	})
	if err != nil {
		slog.Error("failed to schedule report job", "error", err)
	}
}
