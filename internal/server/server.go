package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/redis"
	"github.com/Rafin000/call-recording-service-v2/internal/server/middlewares"
	"github.com/Rafin000/call-recording-service-v2/internal/server/routes"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	Router     *gin.Engine
	httpServer *http.Server
	DB         *mongo.Database
	Config     *common.AppConfig
	Redis      *redis.Client
}

func NewServer(ctx context.Context) (*Server, error) {
	cfg, err := common.LoadConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	setupSlogger(cfg.App)

	// db, err := setupPostgres(ctx, cfg.DB)
	// if err != nil {
	// 	return nil, err
	// }

	// Setup Redis connection
	// redisClient, err := setupRedis(cfg.Redis)
	// if err != nil {
	// 	return nil, err
	// }
	redisClient, err := redis.NewRedis(cfg.Redis)
	if err != nil {
		return nil, err
	}

	// Setup MongoDB connection
	mongoDB, err := setupMongoDB(ctx, cfg.MongoDB)
	if err != nil {
		return nil, err
	}

	router := setupRouter(cfg.App)

	s := &Server{
		Router: router,
		DB:     mongoDB,
		Config: cfg,
		Redis:  redisClient,
		httpServer: &http.Server{
			Addr:    cfg.App.ServerAddress,
			Handler: router,
		},
	}

	s.setupRoutes()
	s.setupMiddlewares()

	return s, nil
}

// Start begins listening for HTTP requests on the configured address.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// setupRoutes initializes all API routes for the server.
func (s *Server) setupRoutes() {
	s.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := s.Router.Group("/api/v1")
	routes.InitRoutes(apiGroup, s.DB, s.Config)
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

// setupPostgres establishes a connection to the PostgreSQL database and runs migrations.
// It returns a database connection pool (*sql.DB) on success.
// func setupPostgres(ctx context.Context, dbConfig common.DBConfig) (*sql.DB, error) {
// 	db, err := postgres.NewConnection(ctx, dbConfig)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to database: %w", err)
// 	}

// 	// if err := postgres.RunMigrations(ctx, db); err != nil {
// 	// 	slog.Warn("failed to run migrations", "err", err)
// 	// 	return nil, fmt.Errorf("failed to run migrations: %w", err)
// 	// }

// 	return db, nil
// }

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
// It uses the provided context for timeout control.
// func (s *Server) Shutdown(ctx context.Context) error {
// 	if err := s.DB.Close(); err != nil {
// 		slog.Error("failed to close database connection", "error", err)
// 	}

// 	if err := s.httpServer.Shutdown(ctx); err != nil {
// 		return fmt.Errorf("server shutdown failed: %w", err)
// 	}

//		return nil
//	}
//

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

func setupRedis(redisConfig common.RedisConfig) (*redis.Client, error) {
	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	}

	client := redis.NewClient(options)

	// Test the connection
	_, err := client.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
