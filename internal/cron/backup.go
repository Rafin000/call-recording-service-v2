package cron

import (
	"log/slog"

	"github.com/Rafin000/call-recording-service-v2/internal/tasks"
	"github.com/go-co-op/gocron"
)

// RegisterReportJobs schedules reporting-related jobs.
func RegisterBackupJobs(scheduler *gocron.Scheduler) {
	_, err := scheduler.Every(15).Seconds().Do(
		tasks.BackupTask(),
	)
	if err != nil {
		slog.Error("failed to schedule report job", "error", err)
	}
}
