package utils

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"

// 	"github.com/Rafin000/call-recording-service-v2/internal/common"
// 	"github.com/go-resty/resty/v2"
// )

// // PortaOneAuth handles PortaOne authentication
// type PortaOneAuth struct {
// 	config     *common.AppConfig
// 	redisUtils *RedisUtils
// }

// // NewPortaOneAuth creates a new PortaOneAuth instance
// func NewPortaOneAuth(config *common.AppConfig, redisUtils *RedisUtils) *PortaOneAuth {
// 	return &PortaOneAuth{
// 		config:     config,
// 		redisUtils: redisUtils,
// 	}
// }

// // SignInToPortaOne handles signing in to PortaOne and returns the session ID
// func (p *PortaOneAuth) SignInToPortaOne(ctx context.Context) (string, error) {
// 	// Check if we already have a session ID in Redis
// 	sessionID, err := p.redisUtils.GetSessionID(ctx)
// 	if err == nil && sessionID != "" {
// 		return sessionID, nil
// 	}

// 	// If no session found in Redis, we need to login
// 	loginURL := "https://pbwebsrv.intercloud.com.bd/rest/Session/login"
// 	loginPayload := map[string]interface{}{
// 		"params": map[string]string{
// 			"login":    p.config.PortaOne.Username,
// 			"password": p.config.PortaOne.Password,
// 		},
// 	}

// 	client := resty.New()
// 	resp, err := client.R().
// 		SetBody(loginPayload).
// 		Post(loginURL)

// 	if err != nil {
// 		return "", fmt.Errorf("error making request to PortaOne: %w", err)
// 	}

// 	if resp.StatusCode() != 200 {
// 		return "", fmt.Errorf("received non-OK status code: %d", resp.StatusCode())
// 	}

// 	var result map[string]interface{}
// 	if err = json.Unmarshal(resp.Body(), &result); err != nil {
// 		return "", fmt.Errorf("error unmarshaling response: %w", err)
// 	}

// 	sessionID, ok := result["session_id"].(string)
// 	if !ok {
// 		return "", fmt.Errorf("no session_id found in response")
// 	}

// 	if err = p.redisUtils.SaveSessionID(ctx, sessionID); err != nil {
// 		return "", fmt.Errorf("error saving session ID to Redis: %w", err)
// 	}

// 	return sessionID, nil
// }
