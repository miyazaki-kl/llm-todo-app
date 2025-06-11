package db

import (
	"fmt"
	"log"
	"myapp/db/model"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// DatabaseConfig データベース設定
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetDefaultConfig デフォルトのデータベース設定を取得
func GetDefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "user"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "myapp"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// getEnv 環境変数を取得、デフォルト値を設定
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// BuildDSN データベース接続文字列を構築
func (config *DatabaseConfig) BuildDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
}

// Connect データベースに接続
func Connect() error {
	config := GetDefaultConfig()
	dsn := config.BuildDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("データベース接続に失敗しました: %w", err)
	}

	// 接続プールの設定
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("データベース接続プールの設定に失敗しました: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db
	log.Println("データベース接続が成功しました")
	return nil
}

// Migrate データベースマイグレーションを実行
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("データベース接続が初期化されていません")
	}

	err := DB.AutoMigrate(
		&model.Todo{},
	)
	if err != nil {
		return fmt.Errorf("マイグレーションに失敗しました: %w", err)
	}

	log.Println("データベースマイグレーションが完了しました")
	return nil
}

// Close データベース接続を閉じる
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetDB データベースインスタンスを取得
func GetDB() *gorm.DB {
	return DB
}
