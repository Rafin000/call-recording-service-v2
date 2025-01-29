package portaone

import (
	"context"
	"encoding/json"
	"fmt"
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
	redis      *redis.RedisClient
	httpClient *resty.Client
}

// NewPortaOneClient creates a new PortaOne client
func NewPortaOneClient(config common.PortaOneConfig, redisClient *redis.RedisClient) PortaOneClient { // Return type should be PortaOneClient interface
	return &portaOneClient{
		config:     config,
		redis:      redisClient,
		httpClient: resty.New(),
	}
}

// GetSessionID retrieves or creates a new PortaOne session
func (c *portaOneClient) GetSessionID(ctx context.Context) (string, error) {
	// Try to get existing session from Redis
	sessionID, err := c.redis.Get(ctx, portaOneSessionKey)
	if err == nil && sessionID != "" {
		return sessionID, nil
	}

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

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(loginPayload).
		Post(loginURL)

	if err != nil {
		return "", fmt.Errorf("PortaOne request failed: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("PortaOne returned non-200 status code: %d", resp.StatusCode())
	}

	var result map[string]interface{}
	if err = json.Unmarshal(resp.Body(), &result); err != nil {
		return "", fmt.Errorf("failed to parse PortaOne response: %w", err)
	}

	sessionID, ok := result["session_id"].(string)
	if !ok {
		return "", fmt.Errorf("session_id not found in PortaOne response")
	}

	// Save session to Redis
	if err = c.redis.Set(ctx, portaOneSessionKey, sessionID, sessionTimeout); err != nil {
		return "", fmt.Errorf("failed to save session to Redis: %w", err)
	}

	return sessionID, nil
}
