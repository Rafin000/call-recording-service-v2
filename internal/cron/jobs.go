package cron

import (
	"log/slog"
	"time"

	"github.com/go-co-op/gocron"
)

// JobManager holds the scheduler instance.
type JobManager struct {
	Scheduler *gocron.Scheduler
}

// NewJobManager initializes a new JobManager.
func NewJobManager() *JobManager {
	scheduler := gocron.NewScheduler(time.UTC)
	return &JobManager{Scheduler: scheduler}
}

// RegisterJobs sets up all scheduled jobs.
func (jm *JobManager) RegisterJobs() {
	// Register individual job files here
	RegisterCleanupJobs(jm.Scheduler)
	RegisterNotificationJobs(jm.Scheduler)
	RegisterReportJobs(jm.Scheduler)

	// Start the scheduler asynchronously
	jm.Scheduler.StartAsync()
	slog.Info("gocron scheduler started")
}
