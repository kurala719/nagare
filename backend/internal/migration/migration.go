package migration

import (
	"os"
	"strings"

	"nagare/internal/core/domain"
	"nagare/internal/core/service"
	"nagare/internal/database"
)

// InitDBTables initializes the database tables and performs necessary migrations.
func InitDBTables() error {
	if err := preSchemaUpdates(); err != nil {
		return err
	}
	if err := database.DB.AutoMigrate(
		&domain.User{},
		&domain.Monitor{},
		&domain.Alarm{},
		&domain.Group{},
		&domain.Host{},
		&domain.Item{},
		&domain.ItemHistory{},
		&domain.HostHistory{},
		&domain.NetworkStatusHistory{},
		&domain.Alert{},
		&domain.Media{},
		&domain.Action{},
		&domain.Trigger{},
		&domain.LogEntry{},
		&domain.AuditLog{},
		&domain.Chat{},
		&domain.Provider{},
		&domain.RegisterApplication{},
		&domain.PasswordResetApplication{},
		&domain.QQWhitelist{},
		&domain.Report{},
		&domain.ReportConfig{},
		&domain.KnowledgeBase{},
		&domain.SiteMessage{},
		&domain.AnsiblePlaybook{},
		&domain.AnsibleJob{},
		&domain.RetentionPolicy{},
	); err != nil {
		return err
	}
	if err := applySchemaUpdates(); err != nil {
		return err
	}
	return ensureDefaultMonitor()
}

func ensureDefaultMonitor() error {
	var count int64
	if err := database.DB.Model(&domain.Monitor{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		defaultMonitor := domain.Monitor{
			Name:        "Nagare Internal",
			URL:         "localhost",
			Type:        4, // SNMP
			Enabled:     1,
			Status:      1,
			Description: "System default internal monitoring engine",
		}
		return database.DB.Create(&defaultMonitor).Error
	}
	return nil
}

func preSchemaUpdates() error {
	// Deduplicate users before adding unique index if table exists
	if database.DB.Migrator().HasTable("users") {
		// This query finds duplicates and deletes all but the one with the smallest ID
		err := database.DB.Exec(`
			DELETE u1 FROM users u1
			INNER JOIN users u2 
			WHERE u1.id > u2.id AND u1.username = u2.username
		`).Error
		if err != nil {
			// Log error but continue - migration might fail later if duplicates persist
			// but we don't want to stop the whole process if this specific MySQL syntax fails
			// (though it's standard for MySQL/MariaDB)
		}
	}

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

	if database.DB.Migrator().HasTable(&domain.Trigger{}) && database.DB.Migrator().HasColumn(&domain.Trigger{}, "log_level") {
		if err := database.DB.Exec("UPDATE triggers SET log_level = CASE WHEN CAST(log_level AS CHAR) IN ('info','warn','warning','error') THEN CASE CAST(log_level AS CHAR) WHEN 'info' THEN 0 WHEN 'warn' THEN 1 WHEN 'warning' THEN 1 WHEN 'error' THEN 2 END WHEN CAST(log_level AS CHAR) REGEXP '^[0-9]+$' THEN CAST(log_level AS UNSIGNED) ELSE NULL END").Error; err != nil {
			return err
		}
	}
	if database.DB.Migrator().HasTable(&domain.LogEntry{}) && database.DB.Migrator().HasColumn(&domain.LogEntry{}, "level") {
		if err := database.DB.Exec("UPDATE log_entries SET level = CASE WHEN CAST(level AS CHAR) IN ('info','warn','warning','error') THEN CASE CAST(level AS CHAR) WHEN 'info' THEN 0 WHEN 'warn' THEN 1 WHEN 'warning' THEN 1 WHEN 'error' THEN 2 END WHEN CAST(level AS CHAR) REGEXP '^[0-9]+$' THEN CAST(level AS UNSIGNED) ELSE NULL END").Error; err != nil {
			return err
		}
	}
	// Migrate monitor type from string to int: 'zabbix' -> 1, 'prometheus' -> 2, others -> 3
	if database.DB.Migrator().HasTable(&domain.Monitor{}) && database.DB.Migrator().HasColumn(&domain.Monitor{}, "type") {
		columnType, err := database.DB.Migrator().ColumnTypes(&domain.Monitor{})
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
	if database.DB.Migrator().HasColumn(&domain.Action{}, "severity_min") {
		if err := database.DB.Migrator().DropColumn(&domain.Action{}, "severity_min"); err != nil {
			return err
		}
	}
	if database.DB.Migrator().HasTable(&domain.LogEntry{}) {
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
