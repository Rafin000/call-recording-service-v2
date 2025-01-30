package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/cron"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/portaone"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/redis"
	"github.com/Rafin000/call-recording-service-v2/internal/server/middlewares"
	"github.com/Rafin000/call-recording-service-v2/internal/server/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	Router         *gin.Engine
	httpServer     *http.Server
	DB             *mongo.Database
	Config         *common.AppConfig
	Redis          *redis.RedisClient
	PortaOneClient *portaone.PortaOneClient
	JobManager     *cron.JobManager
}

func NewServer(ctx context.Context) (*Server, error) {
	cfg, err := common.LoadConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	setupSlogger(cfg.App)

	redisClient, err := redis.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		return nil, err
	}

	// Setup PortaOne client
	portaOneClient, err := portaone.NewPortaOneClient(cfg.PortaOne, redisClient)
	if err != nil {
		return nil, err
	}

	// Setup MongoDB connection
	mongoDB, err := setupMongoDB(ctx, cfg.MongoDB)
	if err != nil {
		return nil, err
	}

	router := setupRouter(cfg.App)

	// Initialize JobManager
	jobManager := cron.NewJobManager(ctx, mongoDB, portaOneClient, cfg.App)
	jobManager.RegisterJobs()

	s := &Server{
		Router:         router,
		DB:             mongoDB,
		Config:         cfg,
		Redis:          &redisClient,
		PortaOneClient: &portaOneClient,
		JobManager:     jobManager,
		httpServer: &http.Server{
			Addr:    cfg.App.ServerAddress,
			Handler: router,
		},
	}

	s.setupRoutes()
	s.setupMiddlewares()
	// s.setupCronJobs()

	return s, nil
}

// setupCronJobs defines and starts scheduled tasks.
// func (s *Server) setupCronJobs() {
// 	_, err := s.Scheduler.Every(30).Seconds().Do(func() {
// 		slog.Info("Running hourly task")
// 		// Add your logic here (e.g., deleting old records, refreshing tokens)
// 	})
// 	if err != nil {
// 		slog.Error("failed to schedule job", "error", err)
// 	}

// 	s.Scheduler.StartAsync()
// 	slog.Info("gocron scheduler started")
// }

// Start begins listening for HTTP requests on the configured address.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// setupRoutes initializes all API routes for the server.
func (s *Server) setupRoutes() {
	s.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := s.Router.Group("/api/v1")
	routes.InitRoutes(apiGroup, s.DB, s.Config, *s.PortaOneClient)
}

// setupMiddlewares adds all necessary middlewares to the Gin router.
func (s *Server) setupMiddlewares() {
	s.Router.Use(middlewares.InitMiddlewares()...)
}

// setupRouter initializes and configures the Gin router.
// It sets the Gin mode based on the application settings and disables trusted proxies.
func setupRouter(appSettings common.AppSettings) *gin.Engine {
	gin.SetMode(appSettings.GinMode)
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	return router
}

// setupSlogger configures the global logger based on the application environment.
// It uses a text handler for development and a JSON handler for other environments.
func setupSlogger(appSettings common.AppSettings) {
	var logLevel = new(slog.LevelVar)
	var handler slog.Handler

	if appSettings.Env == common.AppEnvDev {
		handler = slog.NewTextHandler(os.Stderr, common.GetTextHandlerOptions(logLevel))
	} else {
		handler = slog.NewJSONHandler(os.Stderr, common.GetJSONHandlerOptions(logLevel))
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	if appSettings.GinMode == gin.DebugMode {
		logLevel.Set(slog.LevelDebug)
	}

}

// setupMongoDB establishes a connection to MongoDB and returns the *mongo.Database instance.
func setupMongoDB(ctx context.Context, mongoDBConfig common.MongoDBConfig) (*mongo.Database, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoDBConfig.URI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ensure the connection is established
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client.Database(mongoDBConfig.Database), nil
}

// Shutdown gracefully stops the server, closing the database connection and stopping the HTTP server.
// For MongoDB
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.DB.Client().Disconnect(ctx); err != nil {
		slog.Error("failed to disconnect from MongoDB", "error", err)
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
