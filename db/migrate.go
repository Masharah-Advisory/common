package db

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, models ...interface{}) error {
	log.Println("[COMMON] Running migration...")

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("[COMMON] Migration completed")
	return nil
}
