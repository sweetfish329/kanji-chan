package database

import (
	"fmt"
	"log"
	"os"

	"github.com/sweetfish329/kanji-chan/backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB データベース接続の初期化とマイグレーションの実行
func InitDB() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "kanji_user"
	}
	if password == "" {
		password = "kanji_password"
	}
	if dbname == "" {
		dbname = "kanji_db"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自動マイグレーションの実行
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&model.User{},
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
