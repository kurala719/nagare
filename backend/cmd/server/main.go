package main

import (
	"fmt"
	"os"
	"path/filepath"

	"nagare/cmd/server/router"
	"nagare/internal/database"
	"nagare/internal/migration"
	"nagare/internal/repository"
	"nagare/internal/service"

	"github.com/spf13/viper"
)

const (
	defaultConfigPath = "configs/nagare_config.json"
	envConfigPath     = "NAGARE_CONFIG_PATH"
)

func main() {
	if err := run(); err != nil {
		service.LogSystem("error", "application failed to start", map[string]interface{}{"error": err.Error()}, nil, "")
	}
}

func run() error {
	configPath := getConfigPath()
	fmt.Printf(">>> Loading config from: %s\n", configPath)

	if err := repository.InitConfig(configPath); err != nil {
		return err
	}

	fmt.Println(">>> Initializing Database...")
	if err := database.InitDBFromConfig(); err != nil {
		return err
	}
	defer database.CloseDB(database.DB)

	service.LogSystem("info", "starting application", map[string]interface{}{"system_name": viper.GetString("system.system_name")}, nil, "")

	service.LogSystem("info", "database connection established", nil, nil, "")

	fmt.Println(">>> Running Database Migrations...")
	if err := migration.InitDBTables(); err != nil {
		return err
	}
	
	fmt.Println(">>> Starting Background Services...")
	service.StartAutoSync()
	service.StartStatusChecks()
	service.InitQQWSServ()
	service.LogSystem("info", "background services started", nil, nil, "")

	fmt.Println(">>> Recomputing Action and Trigger statuses...")
	if err := service.RecomputeActionAndTriggerStatuses(); err != nil {
		service.LogSystem("warn", "startup status recompute failed", map[string]interface{}{"error": err.Error()}, nil, "")
	}

	// Initialize cron scheduler for automated reports
	fmt.Println(">>> Initializing Cron Scheduler...")
	if err := service.InitCronScheduler(); err != nil {
		service.LogSystem("warn", "failed to initialize cron scheduler", map[string]interface{}{"error": err.Error()}, nil, "")
	}

	service.LogSystem("info", "initializing router", nil, nil, "")
	fmt.Println(">>> Initializing Router and starting server...")
	router.InitRouter()
	return nil
}

func getConfigPath() string {
	// First, check environment variable
	if path := os.Getenv(envConfigPath); path != "" {
		return path
	}

	// Try current working directory (for development with go run)
	if _, err := os.Stat(defaultConfigPath); err == nil {
		return defaultConfigPath
	}

	// Try relative to executable (for production)
	execPath, err := os.Executable()
	if err != nil {
		return defaultConfigPath
	}

	return filepath.Join(filepath.Dir(execPath), defaultConfigPath)
}
