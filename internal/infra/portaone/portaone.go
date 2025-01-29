package portaone

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/Rafin000/call-recording-service-v2/internal/infra/redis"

	"github.com/go-resty/resty/v2"
)

const (
	portaOneSessionKey = "portaone_session_id"
	sessionTimeout     = 25 * time.Minute
)

type PortaOneClient interface {
	GetSessionID(ctx context.Context) (string, error)
}

// portaOneClient handles PortaOne API interactions
type portaOneClient struct {
	config     common.PortaOneConfig
	redis      redis.RedisClient
	httpClient *resty.Client
}

// // NewPortaOneClient creates a new PortaOne client
// func NewPortaOneClient(config common.PortaOneConfig, redisClient redis.RedisClient) PortaOneClient {
// 	return &portaOneClient{
// 		config:     config,
// 		redis:      redisClient,
// 		httpClient: resty.New(),
// 	}
// }

// NewPortaOneClient creates a new PortaOne client
func NewPortaOneClient(config common.PortaOneConfig, redisClient redis.RedisClient) (PortaOneClient, error) {
	// Check if the Redis client is valid
	if redisClient == nil {
		slog.Error("Invalid Redis client", "error", "client is nil")
		return nil, fmt.Errorf("invalid Redis client: client is nil")
	}

	// Create and return the new PortaOne client
	client := &portaOneClient{
		config:     config,
		redis:      redisClient,
		httpClient: resty.New(),
	}

	slog.Info("PortaOne client created", "config", config)

	return client, nil
}

// GetSessionID retrieves or creates a new PortaOne session
func (c *portaOneClient) GetSessionID(ctx context.Context) (string, error) {
	// Try to get existing session from Redis
	sessionID, err := c.redis.Get(ctx, portaOneSessionKey)
	if err == nil && sessionID != "" {
		slog.Info("Session found in Redis", "sessionID", sessionID)
		return sessionID, nil
	}

	slog.Info("No session found in Redis, creating new session")
	// Create new session if none exists
	return c.login(ctx)
}

// login performs the PortaOne login and stores the session
func (c *portaOneClient) login(ctx context.Context) (string, error) {
	loginURL := "https://pbwebsrv.intercloud.com.bd/rest/Session/login"
	loginPayload := map[string]interface{}{
		"params": map[string]string{
			"login":    c.config.Username,
			"password": c.config.Password,
		},
	}

	slog.Debug("Making login request", "url", loginURL, "payload", loginPayload)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(loginPayload).
		Post(loginURL)

	slog.Debug("Received PortaOne response", "statusCode", resp.StatusCode(), "body", string(resp.Body()))

	if err != nil {
		slog.Error("PortaOne request failed", "error", err)
		return "", fmt.Errorf("PortaOne request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		slog.Error("PortaOne returned non-200 status code", "status_code", resp.StatusCode())
		return "", fmt.Errorf("PortaOne returned non-200 status code: %d", resp.StatusCode())
	}

	var result map[string]interface{}
	if err = json.Unmarshal(resp.Body(), &result); err != nil {
		slog.Error("Failed to parse PortaOne response", "error", err)
		return "", fmt.Errorf("failed to parse PortaOne response: %w", err)
	}

	sessionID, ok := result["session_id"].(string)
	if !ok {
		slog.Error("session_id not found in PortaOne response")
		return "", fmt.Errorf("session_id not found in PortaOne response")
	}

	// Save session to Redis
	if err = c.redis.Set(ctx, portaOneSessionKey, sessionID, sessionTimeout); err != nil {
		slog.Error("Failed to save session to Redis", "error", err)
		return "", fmt.Errorf("failed to save session to Redis: %w", err)
	}

	slog.Info("Successfully created session", "sessionID", sessionID)
	return sessionID, nil
}
