package main

import (
	"os"
	"path/filepath"
	"strings"

	"nagare/cmd/server/router"
	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/service"
	"nagare/pkg/queue"

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

	if err := repository.InitConfig(configPath); err != nil {
		return err
	}

	if err := database.InitDBFromConfig(); err != nil {
		return err
	}
	defer database.CloseDB(database.DB)

	service.LogSystem("info", "starting application", map[string]interface{}{"system_name": viper.GetString("system.system_name")}, nil, "")

	service.LogSystem("info", "database connection established", nil, nil, "")

	if err := initDBTables(); err != nil {
		return err
	}
	if err := service.RecomputeActionAndTriggerStatuses(); err != nil {
		service.LogSystem("warn", "startup status recompute failed", map[string]interface{}{"error": err.Error()}, nil, "")
	}

	// Initialize task queue
	redisAddr := viper.GetString("redis.addr")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	tq, err := queue.New(redisAddr)
	if err != nil {
		service.LogSystem("warn", "failed to initialize task queue", map[string]interface{}{"error": err.Error(), "redis_addr": redisAddr}, nil, "")
	} else {
		service.SetTaskQueue(tq)
		defer tq.Close()
		// Start task workers
		go service.StartTaskWorkers()
		service.LogSystem("info", "task queue initialized", map[string]interface{}{"redis_addr": redisAddr}, nil, "")
	}

	service.StartAutoSync()
	service.StartStatusChecks()

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

func initDBTables() error {
	if err := preSchemaUpdates(); err != nil {
		return err
	}
	if err := database.DB.AutoMigrate(
		&model.User{},
		&model.Monitor{},
		&model.Group{},
		&model.Host{},
		&model.Item{},
		&model.ItemHistory{},
		&model.HostHistory{},
		&model.NetworkStatusHistory{},
		&model.Alert{},
		&model.MediaType{},
		&model.Media{},
		&model.Action{},
		&model.Trigger{},
		&model.LogEntry{},
		&model.Chat{},
		&model.Provider{},
		&model.UserInformation{},
		&model.RegisterApplication{},
	); err != nil {
		return err
	}
	return applySchemaUpdates()
}

func preSchemaUpdates() error {
	if database.DB.Migrator().HasTable("sites") && !database.DB.Migrator().HasTable("groups") {
		if err := database.DB.Migrator().RenameTable("sites", "groups"); err != nil {
			return err
		}
	}
	if database.DB.Migrator().HasTable("hosts") {
		if database.DB.Migrator().HasColumn("hosts", "site_id") && !database.DB.Migrator().HasColumn("hosts", "group_id") {
			if err := database.DB.Migrator().RenameColumn("hosts", "site_id", "group_id"); err != nil {
				return err
			}
		}
	}
	if database.DB.Migrator().HasTable("triggers") {
		if database.DB.Migrator().HasColumn("triggers", "alert_site_id") && !database.DB.Migrator().HasColumn("triggers", "alert_group_id") {
			if err := database.DB.Migrator().RenameColumn("triggers", "alert_site_id", "alert_group_id"); err != nil {
				return err
			}
		}
	}

	if database.DB.Migrator().HasTable(&model.Trigger{}) && database.DB.Migrator().HasColumn(&model.Trigger{}, "log_level") {
		if err := database.DB.Exec("UPDATE triggers SET log_level = CASE WHEN CAST(log_level AS CHAR) IN ('info','warn','warning','error') THEN CASE CAST(log_level AS CHAR) WHEN 'info' THEN 0 WHEN 'warn' THEN 1 WHEN 'warning' THEN 1 WHEN 'error' THEN 2 END WHEN CAST(log_level AS CHAR) REGEXP '^[0-9]+$' THEN CAST(log_level AS UNSIGNED) ELSE NULL END").Error; err != nil {
			return err
		}
	}
	if database.DB.Migrator().HasTable(&model.LogEntry{}) && database.DB.Migrator().HasColumn(&model.LogEntry{}, "level") {
		if err := database.DB.Exec("UPDATE log_entries SET level = CASE WHEN CAST(level AS CHAR) IN ('info','warn','warning','error') THEN CASE CAST(level AS CHAR) WHEN 'info' THEN 0 WHEN 'warn' THEN 1 WHEN 'warning' THEN 1 WHEN 'error' THEN 2 END WHEN CAST(level AS CHAR) REGEXP '^[0-9]+$' THEN CAST(level AS UNSIGNED) ELSE NULL END").Error; err != nil {
			return err
		}
	}
	// Migrate monitor type from string to int: 'zabbix' -> 1, 'prometheus' -> 2, others -> 3
	if database.DB.Migrator().HasTable(&model.Monitor{}) && database.DB.Migrator().HasColumn(&model.Monitor{}, "type") {
		columnType, err := database.DB.Migrator().ColumnTypes(&model.Monitor{})
		if err != nil {
			return err
		}
		var isString bool
		for _, c := range columnType {
			if c.Name() == "type" {
				// Check if it's still a string type (VARCHAR, CHAR, TEXT, etc.)
				typeName := strings.ToUpper(c.DatabaseTypeName())
				isString = strings.Contains(typeName, "VARCHAR") || strings.Contains(typeName, "CHAR") || strings.Contains(typeName, "TEXT")
				break
			}
		}
		if isString {
			if err := database.DB.Exec("UPDATE monitors SET type = CASE WHEN LOWER(type) = 'zabbix' THEN '1' WHEN LOWER(type) = 'prometheus' THEN '2' ELSE '3' END WHERE type IS NOT NULL").Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func applySchemaUpdates() error {
	// Explicitly remove deprecated columns that AutoMigrate does not drop.
	if database.DB.Migrator().HasColumn(&model.Action{}, "severity_min") {
		if err := database.DB.Migrator().DropColumn(&model.Action{}, "severity_min"); err != nil {
			return err
		}
	}
	if database.DB.Migrator().HasTable(&model.LogEntry{}) {
		if err := database.DB.Exec("ALTER TABLE log_entries MODIFY COLUMN level INT").Error; err != nil {
			return err
		}
	}
	if os.Getenv("NAGARE_MEDIA_BACKFILL") == "1" {
		updated, skipped, err := service.BackfillMediaParamsAndTargetsServ()
		if err != nil {
			return err
		}
		service.LogSystem("info", "media backfill completed", map[string]interface{}{"updated": updated, "skipped": skipped}, nil, "")
	}
	return nil
}
