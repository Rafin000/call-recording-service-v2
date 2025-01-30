package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/domain"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/portaone"
	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/mongo"
)

// JobManager holds the scheduler instance.
type JobManager struct {
	Scheduler      *gocron.Scheduler
	UserRepo       domain.UserRepository
	XDRRepo        domain.XDRRepository
	portaOneClient portaone.PortaOneClient
	config         common.AppSettings
	C              context.Context
}

// NewJobManager initializes a new JobManager.
func NewJobManager(c context.Context, mongoDB *mongo.Database, portaOneClient portaone.PortaOneClient, cfg common.AppSettings) *JobManager {
	userRepo := domain.NewUserRepository(mongoDB)
	XDRRepo := domain.NewXDRRepository(mongoDB)
	scheduler := gocron.NewScheduler(time.UTC)
	return &JobManager{Scheduler: scheduler, UserRepo: userRepo, XDRRepo: XDRRepo, C: c, portaOneClient: portaOneClient, config: cfg}
}

// RegisterJobs sets up all scheduled jobs.
func (jm *JobManager) RegisterJobs() {

	RegisterBackupJobs(jm.Scheduler, jm.UserRepo, jm.XDRRepo, jm.C, jm.config, jm.portaOneClient)

	jm.Scheduler.StartAsync()
	slog.Info("gocron scheduler started")
}
