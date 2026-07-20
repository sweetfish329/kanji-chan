package database

import (
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB データベース接続の初期化とマイグレーションの実行
func InitDB() (*gorm.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		if _, err := os.Stat("/data"); err == nil {
			dbPath = "/data/kanji.db"
		} else {
			dbPath = "kanji.db"
		}
	}

	// 外部キー制約を有効にするために _pragma=foreign_keys(1) を付与
	dsn := fmt.Sprintf("%s?_pragma=foreign_keys(1)", dbPath)
	log.Printf("Connecting to SQLite database at: %s", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自動マイグレーションの実行
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&model.User{},
		&model.ApiKey{},
		&model.Event{},
		&model.EventCandidate{},
		&model.Response{},
		&model.CandidateAnswer{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	DB = db
	log.Println("Database connection established and migrated successfully.")
	return db, nil
}
