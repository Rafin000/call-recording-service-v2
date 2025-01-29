package cron

import (
	"log/slog"

	"github.com/go-co-op/gocron"
)

// RegisterCleanupJobs schedules cleanup-related jobs.
func RegisterCleanupJobs(scheduler *gocron.Scheduler) {
	_, err := scheduler.Every(10).Seconds().Do(func() {
		slog.Info("Running daily cleanup task")
		// Add your cleanup logic here (e.g., deleting old records)
	})
	if err != nil {
		slog.Error("failed to schedule cleanup job", "error", err)
	}
}
