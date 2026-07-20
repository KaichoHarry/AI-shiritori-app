package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenIssuer struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewTokenIssuer(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenIssuer {
	return &TokenIssuer{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

type claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func (t *TokenIssuer) IssueAccessToken(userID string) (string, error) {
	return t.issue(userID, t.accessSecret, t.accessTTL)
}

func (t *TokenIssuer) IssueRefreshToken(userID string) (string, error) {
	return t.issue(userID, t.refreshSecret, t.refreshTTL)
}

func (t *TokenIssuer) issue(userID string, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})
	return token.SignedString(secret)
}

func (t *TokenIssuer) ParseAccessToken(tokenString string) (string, error) {
	return t.parse(tokenString, t.accessSecret)
}

func (t *TokenIssuer) ParseRefreshToken(tokenString string) (string, error) {
	return t.parse(tokenString, t.refreshSecret)
}

func (t *TokenIssuer) parse(tokenString string, secret []byte) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}
	c, ok := token.Claims.(*claims)
	if !ok || c.UserID == "" {
		return "", ErrInvalidToken
	}
	return c.UserID, nil
}
