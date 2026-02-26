package database

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// DBConfig holds database configuration
type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	DBName   string
}

// CloseDB closes the database connection
func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// InitDB initializes the database connection with the given configuration
func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	return nil
}

func InitDBFromConfig() error {
	user := viper.GetString("database.username")
	password := viper.GetString("database.password")
	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	dbname := viper.GetString("database.database_name")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
	if err := InitDB(dsn); err != nil {
		return err
	}
	return applyPoolSettings()
}

// ReapplyPoolSettings updates the pool settings from configuration
func ReapplyPoolSettings() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return applyPoolSettings()
}

func applyPoolSettings() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	maxOpen := viper.GetInt("database.pool.max_open")
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := viper.GetInt("database.pool.max_idle")
	if maxIdle <= 0 {
		maxIdle = 10
	}
	maxLifetime := viper.GetDuration("database.pool.max_lifetime")
	if maxLifetime <= 0 {
		maxLifetime = time.Hour
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(maxLifetime)
	return nil
}
