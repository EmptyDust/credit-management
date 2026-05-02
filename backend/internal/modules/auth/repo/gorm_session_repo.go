package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	authdomain "credit-management/backend/internal/modules/auth/domain"
	"credit-management/backend/internal/platform/authn"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	userRevokedAfterKeyPrefix = "auth:user-revoked-after:"
	userRevokedAfterTTLBuffer = 24 * time.Hour
)

var errSessionNotFound = errors.New("session not found")

type GormSessionRepo struct {
	db    *gorm.DB
	redis *goredis.Client
}

type sessionRow struct {
	ID               string         `gorm:"column:id"`
	UserID           string         `gorm:"column:user_id"`
	RefreshTokenHash string         `gorm:"column:refresh_token_hash"`
	IssuedAt         time.Time      `gorm:"column:issued_at"`
	ExpiresAt        time.Time      `gorm:"column:expires_at"`
	RevokedAt        sql.NullTime   `gorm:"column:revoked_at"`
	RevokeReason     sql.NullString `gorm:"column:revoke_reason"`
	ClientIP         sql.NullString `gorm:"column:client_ip"`
	UserAgent        sql.NullString `gorm:"column:user_agent"`
}

type userRevocation struct {
	RevokedAt time.Time
	Reason    string
}

func (sessionRow) TableName() string {
	return "auth_sessions"
}

func NewGorm(db *gorm.DB, redisClient *goredis.Client) *GormSessionRepo {
	return &GormSessionRepo{
		db:    db,
		redis: redisClient,
	}
}

func (r *GormSessionRepo) SaveSession(ctx context.Context, session authdomain.Session) error {
	row := sessionRow{
		ID:               session.ID,
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		IssuedAt:         session.IssuedAt.UTC(),
		ExpiresAt:        session.ExpiresAt.UTC(),
		ClientIP:         toNullString(session.ClientIP),
		UserAgent:        toNullString(session.UserAgent),
	}
	if session.RevokedAt != nil {
		row.RevokedAt = sql.NullTime{Time: session.RevokedAt.UTC(), Valid: true}
	}
	if session.RevokeReason != "" {
		row.RevokeReason = sql.NullString{String: session.RevokeReason, Valid: true}
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"user_id",
				"refresh_token_hash",
				"issued_at",
				"expires_at",
				"revoked_at",
				"revoke_reason",
				"client_ip",
				"user_agent",
			}),
		}).
		Create(&row).Error
}

func (r *GormSessionRepo) GetSession(ctx context.Context, sessionID string) (*authdomain.Session, error) {
	var row sessionRow
	if err := r.db.WithContext(ctx).First(&row, "id = ?", sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errSessionNotFound
		}
		return nil, err
	}

	session := row.toDomain()
	userRevoke, err := r.getUserRevocation(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if userRevoke != nil && !session.IssuedAt.After(userRevoke.RevokedAt) {
		if session.RevokedAt == nil || session.RevokedAt.Before(userRevoke.RevokedAt) {
			revokedAt := userRevoke.RevokedAt.UTC()
			session.RevokedAt = &revokedAt
			if userRevoke.Reason != "" {
				session.RevokeReason = userRevoke.Reason
			}
		}
	}

	return session, nil
}

func (r *GormSessionRepo) RevokeSession(ctx context.Context, sessionID string, reason string) error {
	now := time.Now().UTC()
	updates := map[string]any{
		"revoked_at":    now,
		"revoke_reason": nullableReason(reason),
	}

	return r.db.WithContext(ctx).
		Model(&sessionRow{}).
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Updates(updates).Error
}

func (r *GormSessionRepo) RevokeAllUserSessions(ctx context.Context, userID string, reason string) error {
	now := time.Now().UTC()

	if err := r.setUserRevocation(ctx, userID, userRevocation{
		RevokedAt: now,
		Reason:    reason,
	}); err != nil {
		return err
	}

	updates := map[string]any{
		"revoked_at":    now,
		"revoke_reason": nullableReason(reason),
	}

	return r.db.WithContext(ctx).
		Model(&sessionRow{}).
		Where("user_id = ? AND revoked_at IS NULL AND issued_at <= ?", userID, now).
		Updates(updates).Error
}

func (r *GormSessionRepo) getUserRevocation(ctx context.Context, userID string) (*userRevocation, error) {
	if r.redis == nil {
		return nil, nil
	}

	value, err := r.redis.Get(ctx, userRevocationKey(userID)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user session revocation: %w", err)
	}

	parts := strings.SplitN(value, "|", 2)
	if len(parts) == 0 || parts[0] == "" {
		return nil, nil
	}

	unixNano, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse user session revocation timestamp: %w", err)
	}

	state := &userRevocation{
		RevokedAt: time.Unix(0, unixNano).UTC(),
	}
	if len(parts) == 2 {
		state.Reason = parts[1]
	}

	return state, nil
}

func (r *GormSessionRepo) setUserRevocation(ctx context.Context, userID string, state userRevocation) error {
	if r.redis == nil {
		return nil
	}

	ttl := authn.RefreshTTL() + userRevokedAfterTTLBuffer
	if ttl <= 0 {
		ttl = userRevokedAfterTTLBuffer
	}

	value := strconv.FormatInt(state.RevokedAt.UTC().UnixNano(), 10) + "|" + state.Reason
	if err := r.redis.Set(ctx, userRevocationKey(userID), value, ttl).Err(); err != nil {
		return fmt.Errorf("set user session revocation: %w", err)
	}

	return nil
}

func (r sessionRow) toDomain() *authdomain.Session {
	session := &authdomain.Session{
		ID:               r.ID,
		UserID:           r.UserID,
		RefreshTokenHash: r.RefreshTokenHash,
		IssuedAt:         r.IssuedAt.UTC(),
		ExpiresAt:        r.ExpiresAt.UTC(),
		ClientIP:         r.ClientIP.String,
		UserAgent:        r.UserAgent.String,
	}

	if r.RevokedAt.Valid {
		revokedAt := r.RevokedAt.Time.UTC()
		session.RevokedAt = &revokedAt
	}
	if r.RevokeReason.Valid {
		session.RevokeReason = r.RevokeReason.String
	}

	return session
}

func userRevocationKey(userID string) string {
	return userRevokedAfterKeyPrefix + userID
}

func nullableReason(reason string) any {
	if reason == "" {
		return nil
	}
	return reason
}

func toNullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
