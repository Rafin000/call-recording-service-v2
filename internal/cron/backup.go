package cron

import (
	"context"
	"log/slog"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/portaone"
	"github.com/Rafin000/call-recording-service-v2/internal/tasks"
	"github.com/go-co-op/gocron"
)

// RegisterReportJobs schedules reporting-related jobs.
func RegisterBackupJobs(scheduler *gocron.Scheduler, userRepo domain.UserRepository, XDRRepo domain.XDRRepository, c context.Context, cfg common.AppSettings, portaOneClient portaone.PortaOneClient) {
	_, err := scheduler.Every(5).Minutes().Do(func() {
		tasks.BackupTask(userRepo, XDRRepo, c, cfg, portaOneClient)
	})
	if err != nil {
		slog.Error("failed to schedule report job", "error", err)
	}
}
