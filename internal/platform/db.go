package platform

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenDB connects to PostgreSQL with GORM.
func OpenDB(dsn, env string) (*gorm.DB, error) {
	level := logger.Warn
	if env == "development" {
		level = logger.Info
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
}
