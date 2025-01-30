package utils

import (
	"errors"
	"time"

	"github.com/Rafin000/call-recording-service-v2/internal/common"
	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	Name      string  `json:"name"`
	ICustomer *string `json:"i_customer,omitempty"`
	jwt.StandardClaims
}

func DecodeAuthToken(token string, config common.AppConfig) (*JWTClaims, error) {
	claims := &JWTClaims{}
	tokenParsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.App.SECRET_KEY), nil
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

func GenerateAccessToken(payloads map[string]interface{}, config common.AppConfig) (string, error) {
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
	return token.SignedString([]byte(config.App.SECRET_KEY))
}

func GenerateRefreshToken(payloads map[string]interface{}, config common.AppConfig) (string, error) {
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
	return token.SignedString([]byte(config.App.SECRET_KEY))
}
