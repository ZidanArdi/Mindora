package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"mindora/backend/internal/config"
)

// Database encapsulates the GORM DB instance and its underlying SQL connection pool.
type Database struct {
	DB    *gorm.DB
	sqlDB *sql.DB
}

// Connect establishes a centralized GORM PostgreSQL connection based on the application configuration.
func Connect(cfg *config.Config) (*Database, error) {
	var gormLogLevel logger.LogLevel
	if cfg.Env == "production" {
		gormLogLevel = logger.Warn
	} else {
		gormLogLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	}

	dsn := cfg.Database.DSN()
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sql.DB: %w", err)
	}

	// Apply connection pool settings
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("database ping check failed: %w", err)
	}

	log.Printf("PostgreSQL database connection successfully verified on port %s", cfg.Database.Port)

	return &Database{
		DB:    db,
		sqlDB: sqlDB,
	}, nil
}

// Ping verifies that the database connection is alive.
func (d *Database) Ping(ctx context.Context) error {
	if d == nil || d.sqlDB == nil {
		return fmt.Errorf("database instance is uninitialized")
	}
	return d.sqlDB.PingContext(ctx)
}

// Close gracefully closes the database connection pool.
func (d *Database) Close() error {
	if d != nil && d.sqlDB != nil {
		log.Println("Closing database connection pool...")
		return d.sqlDB.Close()
	}
	return nil
}

// SQLDB returns the underlying *sql.DB instance.
func (d *Database) SQLDB() *sql.DB {
	if d == nil {
		return nil
	}
	return d.sqlDB
}
