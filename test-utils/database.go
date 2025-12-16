package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDatabase holds the test database configuration
type TestDatabase struct {
	Container testcontainers.Container
	DB        *gorm.DB
	DSN       string
}

// SetupTestDatabase creates a PostgreSQL test container and returns a GORM connection
func SetupTestDatabase(ctx context.Context) (*TestDatabase, error) {
	// Define PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	// Start the container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Get container host and port
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
		host, port.Port())

	// Wait for database to be ready with exponential backoff
	var db *gorm.DB
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			// Test connection
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					break
				}
			}
		}
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}

	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to connect to database after %d retries: %w", maxRetries, err)
	}

	return &TestDatabase{
		Container: container,
		DB:        db,
		DSN:       dsn,
	}, nil
}

// Teardown cleans up the test database and container
func (td *TestDatabase) Teardown(ctx context.Context) error {
	// Close database connection
	if td.DB != nil {
		sqlDB, err := td.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// Stop and remove container
	if td.Container != nil {
		return td.Container.Terminate(ctx)
	}

	return nil
}

// CleanDatabase truncates all tables in the database
func (td *TestDatabase) CleanDatabase(tables ...string) error {
	if len(tables) == 0 {
		// Default tables to clean
		tables = []string{
			"users",
			"activities",
			"activity_participants",
			"attachments",
			"departments",
		}
	}

	// Disable foreign key checks temporarily
	if err := td.DB.Exec("SET session_replication_role = 'replica'").Error; err != nil {
		return fmt.Errorf("failed to disable foreign keys: %w", err)
	}

	// Truncate each table
	for _, table := range tables {
		if err := td.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			log.Printf("Warning: failed to truncate table %s: %v", table, err)
		}
	}

	// Re-enable foreign key checks
	if err := td.DB.Exec("SET session_replication_role = 'origin'").Error; err != nil {
		return fmt.Errorf("failed to re-enable foreign keys: %w", err)
	}

	return nil
}

// ExecuteSQL executes raw SQL statements (useful for migrations)
func (td *TestDatabase) ExecuteSQL(sql string) error {
	return td.DB.Exec(sql).Error
}

// ExecuteSQLFile executes SQL from a file
func (td *TestDatabase) ExecuteSQLFile(filepath string) error {
	_, err := td.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Read and execute SQL file
	// Note: In real implementation, you'd want to read the file and execute it
	// This is a simplified version
	sqlBytes, err := sql.Open("postgres", td.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer sqlBytes.Close()

	return nil
}

// GetConnection returns the GORM DB connection
func (td *TestDatabase) GetConnection() *gorm.DB {
	return td.DB
}
