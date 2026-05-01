package main

import (
	"context"
	"log"

	"credit-management/backend/internal/platform/config"
	"credit-management/backend/internal/platform/database"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := database.Open(ctx, database.Options{DSN: cfg.DatabaseDSN})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	if err := database.NewMigrator("./migrations").Apply(ctx, sqlDB); err != nil {
		log.Fatal(err)
	}
}
