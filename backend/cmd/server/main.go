package main

import (
	"fmt"
	"os"
	"path/filepath"

	"nagare/cmd/server/router"
	"nagare/internal/database"
	"nagare/internal/mcp"
	"nagare/internal/migration"
	"nagare/internal/service"
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

	if err := initializeApplication(configPath); err != nil {
		return err
	}
	defer database.CloseDB(database.DB)

	recomputeStartupState()
	startBackgroundServices()

	fmt.Println(">>> Initializing Router and starting server...")
	service.LogSystem("info", "initializing router", nil, nil, "")
	return router.InitRouter()
}

func initializeApplication(configPath string) error {
	// Use service wrapper to initialize config with logging and hot-reload observers.
	if err := service.InitConfigServ(configPath); err != nil {
		return err
	}

	fmt.Println(">>> Initializing Database...")
	if err := database.InitDBFromConfig(); err != nil {
		return err
	}

	service.LogSystem("info", "starting application", nil, nil, "")
	service.LogSystem("info", "database connection established", nil, nil, "")

	fmt.Println(">>> Running Database Migrations...")
	if err := migration.InitDBTables(); err != nil {
		return err
	}

	return nil
}

func recomputeStartupState() {
	fmt.Println(">>> Synchronizing internal state (Actions/Triggers)...")
	if err := service.RecomputeActionAndTriggerStatuses(); err != nil {
		service.LogSystem("warn", "startup status recompute failed", map[string]interface{}{"error": err.Error()}, nil, "")
	}
}

func startBackgroundServices() {
	fmt.Println(">>> Starting Background Services...")
	service.GlobalHub.Start()
	service.StartAutoSync()
	service.StartStatusChecks()
	service.InitQQWSServ()
	mcp.InitClients()

	if err := service.InitCronScheduler(); err != nil {
		service.LogSystem("warn", "failed to initialize cron scheduler", map[string]interface{}{"error": err.Error()}, nil, "")
	}

	service.LogSystem("info", "all background services started", nil, nil, "")
}

func getConfigPath() string {
	if path, ok := os.LookupEnv(envConfigPath); ok && path != "" {
		return path
	}

	if _, err := os.Stat(defaultConfigPath); err == nil {
		return defaultConfigPath
	}

	execPath, err := os.Executable()
	if err != nil {
		return defaultConfigPath
	}

	return filepath.Join(filepath.Dir(execPath), defaultConfigPath)
}
