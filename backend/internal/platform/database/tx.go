package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type TxFunc func(tx *gorm.DB) error

func WithTransaction(ctx context.Context, db *gorm.DB, fn TxFunc) error {
	if db == nil {
		return errors.New("gorm db is required")
	}
	if fn == nil {
		return errors.New("transaction callback is required")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
