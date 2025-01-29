package cron

import (
	"log/slog"

	"github.com/go-co-op/gocron"
)

// RegisterNotificationJobs schedules notification-related jobs.
func RegisterNotificationJobs(scheduler *gocron.Scheduler) {
	_, err := scheduler.Every(30).Seconds().Do(func() {
		slog.Info("Checking for pending notifications")
		// Add your notification logic
	})
	if err != nil {
		slog.Error("failed to schedule notification job", "error", err)
	}
}
