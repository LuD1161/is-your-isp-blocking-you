package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initialiseDB(storeInDB, scanID string) (*gorm.DB, error) {
	switch storeInDB {
	case "postgres":
		postgresDSN := os.Getenv("POSTGRES_DSN")
		if postgresDSN == "" {
			return nil, errors.New("POSTGRES_DSN not specified")
		}
		db, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
		if err != nil {
			log.Error().Msgf("Can't open DB connection : %s", err.Error())
			return nil, err
		}
		return db, err
	case "sqlite":
		dbFileName := fmt.Sprintf("is_your_isp_blocking_you-%s.db", scanID)
		db, err := gorm.Open(sqlite.Open(dbFileName), &gorm.Config{})
		if err != nil {
			log.Error().Msgf("Can't open DB connection : %s", err.Error())
			return db, err
		}
		return db, err
	case "mysql":
	default:
		if storeInDB != "" {
			fmt.Println("mysql is WIP. Please try with 'postgres' or default is 'sqlite'.")
		}
		return nil, nil

	}
	return nil, nil
}

func saveToDB(db *gorm.DB, results []Record, scanStats ScanStats) error {
	// Perform the
	dbn, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer dbn.Close()
	db.AutoMigrate(Record{}, ScanStats{})
	if err := db.CreateInBatches(results, 1000).Error; err != nil {
		log.Error().Stack().Err(err).Msgf("Error saving results in DB [CreateInBatches] : %s", err.Error())
		return err
	}
	// Create scan stats
	return db.Create(&scanStats).Error
}
