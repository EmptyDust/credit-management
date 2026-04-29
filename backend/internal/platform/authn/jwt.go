package authn

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type ClaimsInput struct {
	UserID       string
	Username     string
	UserType     string
	SessionID    string
	TokenVersion int64
}

type Claims struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	UserType     string `json:"user_type"`
	SessionID    string `json:"session_id"`
	TokenVersion int64  `json:"token_version,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

var (
	managerMu sync.RWMutex
	manager   = NewManager(Config{
		AccessSecret:  "credit-management-dev-access-secret",
		RefreshSecret: "credit-management-dev-refresh-secret",
		AccessTTL:     24 * time.Hour,
		RefreshTTL:    7 * 24 * time.Hour,
	})
)

func NewManager(cfg Config) *Manager {
	if cfg.AccessSecret == "" {
		cfg.AccessSecret = "credit-management-dev-access-secret"
	}
	if cfg.RefreshSecret == "" {
		cfg.RefreshSecret = "credit-management-dev-refresh-secret"
	}
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 24 * time.Hour
	}
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = 7 * 24 * time.Hour
	}

	return &Manager{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}

func Configure(cfg Config) {
	managerMu.Lock()
	defer managerMu.Unlock()
	manager = NewManager(cfg)
}

func AccessTTL() time.Duration {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return manager.accessTTL
}

func RefreshTTL() time.Duration {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return manager.refreshTTL
}

func IssueAccessToken(input ClaimsInput) (string, error) {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return manager.issueToken(input, manager.accessSecret, manager.accessTTL, false)
}

func IssueRefreshToken(input ClaimsInput) (string, error) {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return manager.issueToken(input, manager.refreshSecret, manager.refreshTTL, true)
}

func ParseAccessToken(token string) (Claims, error) {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return manager.parseToken(token, manager.accessSecret)
}

func ParseRefreshToken(token string) (Claims, error) {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return manager.parseToken(token, manager.refreshSecret)
}

func (m *Manager) issueToken(input ClaimsInput, secret []byte, ttl time.Duration, includeTokenVersion bool) (string, error) {
	if input.UserID == "" || input.Username == "" || input.UserType == "" {
		return "", errors.New("user_id, username, and user_type are required")
	}

	now := time.Now().UTC()
	claims := Claims{
		UserID:    input.UserID,
		Username:  input.Username,
		UserType:  input.UserType,
		SessionID: input.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	if includeTokenVersion {
		claims.TokenVersion = input.TokenVersion
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (m *Manager) parseToken(token string, secret []byte) (Claims, error) {
	var claims Claims

	parsed, err := jwt.ParseWithClaims(token, &claims, func(parsed *jwt.Token) (any, error) {
		if parsed.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", parsed.Method.Alg())
		}

		return secret, nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !parsed.Valid {
		return Claims{}, errors.New("token is invalid")
	}

	return claims, nil
}
