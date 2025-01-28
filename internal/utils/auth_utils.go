package utils

import (
	"bytes"
	"errors"
	"time"

	"context"
	"encoding/json"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
)

type JWTClaims struct {
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	Name      string  `json:"name"`
	ICustomer *string `json:"i_customer,omitempty"`
	jwt.StandardClaims
}

var (
	SecretKey        = "your-secret-key" // Replace with actual secret key
	PortaOneUsername = "your-username"   // Replace with actual PortaOne username
	PortaOnePassword = "your-password"   // Replace with actual PortaOne password
	RedisClient      *redis.Client       // Initialize Redis client elsewhere in your app
)

func DecodeAuthToken(token string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	tokenParsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, errors.New("invalid")
		}
		if vErr, ok := err.(*jwt.ValidationError); ok && vErr.Errors == jwt.ValidationErrorExpired {
			return nil, errors.New("expired")
		}
		return nil, errors.New("invalid")
	}
	if !tokenParsed.Valid {
		return nil, errors.New("invalid")
	}
	return claims, nil
}

func GenerateAccessToken(payloads map[string]interface{}) (string, error) {
	claims := &JWTClaims{
		Email:     payloads["email"].(string),
		Role:      payloads["role"].(string),
		Name:      payloads["name"].(string),
		ICustomer: payloads["i_customer"].(*string),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(SecretKey))
}

func GenerateRefreshToken(payloads map[string]interface{}) (string, error) {
	claims := &JWTClaims{
		Email:     payloads["email"].(string),
		Role:      payloads["role"].(string),
		Name:      payloads["name"].(string),
		ICustomer: payloads["i_customer"].(*string),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(SecretKey))
}

func SignInToPortaOne(ctx context.Context) (string, error) {
	loginURL := "https://pbwebsrv.intercloud.com.bd/rest/Session/login"
	loginPayload := map[string]interface{}{
		"params": map[string]string{
			"login":    PortaOneUsername,
			"password": PortaOnePassword,
		},
	}

	sessionID, err := RedisClient.Get(ctx, "portaone_session_id").Result()
	if err == nil {
		return sessionID, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	payloadBytes, err := json.Marshal(loginPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", loginURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to login to PortaOne")
	}

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", err
	}

	sessionID, ok := responseData["session_id"].(string)
	if !ok {
		return "", errors.New("session_id not found in response")
	}

	saveSessionID(ctx, sessionID)
	return sessionID, nil
}

func saveSessionID(ctx context.Context, sessionID string) {
	RedisClient.Set(ctx, "portaone_session_id", sessionID, 25*time.Minute)
}
