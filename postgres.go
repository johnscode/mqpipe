package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"time"
)

func setupPostgres(logger *zerolog.Logger) *Repository {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbName := os.Getenv("POSTGRES_DB")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	logger.Info().Msg(fmt.Sprintf("Connecting to PostgreSQL at %s", dsn))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&IoTDeviceMessage{}, &DeviceModel{}, &TempRHDevice{})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to migrate models")
	}

	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	repo := NewRepository(db, logger)
	return repo
}
