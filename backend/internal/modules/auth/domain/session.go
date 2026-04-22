package domain

import "time"

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	IssuedAt         time.Time
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	RevokeReason     string
	ClientIP         string
	UserAgent        string
}

func (s Session) IsRevoked() bool {
	return s.RevokedAt != nil
}
