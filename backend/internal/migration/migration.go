package migration

import (
	"fmt"
	"os"
	"strings"

	"nagare/internal/database"
	"nagare/internal/model"
	"nagare/internal/service"
)

// InitDBTables initializes the database tables and performs necessary migrations.
func InitDBTables() error {
	if err := preSchemaUpdates(); err != nil {
		return err
	}
	// Create users table manually only if it doesn't exist
	if !database.DB.Migrator().HasTable("users") {
		if err := database.DB.Exec(`CREATE TABLE users (
			id bigint unsigned NOT NULL auto_increment,
			created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at timestamp NULL,
			username varchar(100) NOT NULL,
			password longtext NOT NULL,
			privileges int default 1,
			status int default 1,
			email varchar(255),
			phone varchar(20),
			avatar varchar(255),
			address varchar(255),
			introduction text,
			nickname varchar(100),
			qq varchar(20),
			PRIMARY KEY (id),
			KEY idx_users_deleted_at (deleted_at),
			UNIQUE KEY idx_username (username(100))
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`).Error; err != nil {
			return fmt.Errorf("failed to create users table: %w", err)
		}
	}

	// Now run AutoMigrate on all models.
	// Since we enabled Username in the struct, GORM will handle it.
	if err := database.DB.AutoMigrate(
		&model.User{},
		&model.Monitor{},
		&model.Alarm{},
		&model.Group{},
		&model.Host{},
		&model.Item{},
		&model.ItemHistory{},
		&model.HostHistory{},
		&model.NetworkStatusHistory{},
		&model.Alert{},
		&model.Media{},
		&model.Action{},
		&model.Trigger{},
		&model.LogEntry{},
		&model.AuditLog{},
		&model.Chat{},
		&model.Provider{},
		&model.RegisterApplication{},
		&model.PasswordResetApplication{},
		&model.QQWhitelist{},
		&model.Report{},
		&model.ReportConfig{},
		&model.KnowledgeBase{},
		&model.SiteMessage{},
		&model.AnsiblePlaybook{},
		&model.AnsibleJob{},
		&model.RetentionPolicy{},
		&model.PacketAnalysis{},
	); err != nil {
		return err
	}
	if err := applySchemaUpdates(); err != nil {
		return err
	}
	if err := ensureDefaultAdmin(); err != nil {
		return err
	}
	if err := ensureDefaultMonitor(); err != nil {
		return err
	}
	if err := ensureDefaultReportConfig(); err != nil {
		return err
	}
	return ensureDefaultRetentionPolicies()
}

func ensureDefaultRetentionPolicies() error {
	// Cleanup malformed entries first
	if err := database.DB.Where("data_type = '' OR data_type IS NULL").Delete(&model.RetentionPolicy{}).Error; err != nil {
		return err
	}

	var count int64
	if err := database.DB.Model(&model.RetentionPolicy{}).Count(&count).Error; err != nil {
		return err
	}

	defaults := []model.RetentionPolicy{
		{DataType: "logs", RetentionDays: 30, Description: "System and service logs"},
		{DataType: "alerts", RetentionDays: 90, Description: "Alert history"},
		{DataType: "audit_logs", RetentionDays: 180, Description: "User operational logs"},
		{DataType: "item_history", RetentionDays: 30, Description: "Metric history data"},
		{DataType: "host_history", RetentionDays: 30, Description: "Host status history"},
		{DataType: "network_history", RetentionDays: 90, Description: "Network health score history"},
		{DataType: "chat", RetentionDays: 30, Description: "AI chat messages"},
		{DataType: "ansible_jobs", RetentionDays: 30, Description: "Ansible execution logs"},
		{DataType: "reports", RetentionDays: 30, Description: "Generated PDF reports"},
		{DataType: "site_messages", RetentionDays: 30, Description: "User notifications"},
	}

	enabled := 1
	for _, d := range defaults {
		var existing model.RetentionPolicy
		err := database.DB.Where("data_type = ?", d.DataType).First(&existing).Error
		if err != nil { // Not found or error
			d.Enabled = &enabled
			if err := database.DB.Create(&d).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func ensureDefaultAdmin() error {
	var count int64
	if err := database.DB.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		admin := model.User{
			Username:   "admin",
			Password:   "password", // Initial default password
			Privileges: 3,          // Superadmin
			Status:     1,          // Active
			Nickname:   "System Administrator",
		}
		return database.DB.Create(&admin).Error
	}
	return nil
}

func ensureDefaultMonitor() error {
	var count int64
	if err := database.DB.Model(&model.Monitor{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		defaultMonitor := model.Monitor{
			Name:        "Nagare Internal",
			URL:         "localhost",
			Type:        1, // SNMP
			Enabled:     1,
			Status:      1,
			Description: "System default internal monitoring engine",
		}
		// Set ID to 1 explicitly
		defaultMonitor.ID = 1
		return database.DB.Create(&defaultMonitor).Error
	}

	// Enforce Type 1 (SNMP) for ID 1 if it exists
	return database.DB.Model(&model.Monitor{}).Where("id = ?", 1).Update("type", 1).Error
}

func ensureDefaultReportConfig() error {
	var count int64
	if err := database.DB.Model(&model.ReportConfig{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		config := model.ReportConfig{
			AutoGenerateWeekly:  0,
			WeeklyGenerateDay:   "Monday",
			WeeklyGenerateTime:  "09:00",
			AutoGenerateMonthly: 0,
			MonthlyGenerateDate: 1,
			MonthlyGenerateTime: "09:00",
			IncludeAlerts:       1,
			IncludeMetrics:      1,
			TopHostsCount:       5,
			EnableLLMSummary:    1,
		}
		return database.DB.Create(&config).Error
	}
	return nil
}

func preSchemaUpdates() error {
	// Deduplicate users before adding unique index if table exists
	if database.DB.Migrator().HasTable("users") {
		// This query finds duplicates and deletes all but the one with the smallest ID
		_ = database.DB.Exec(`
			DELETE u1 FROM users u1
			INNER JOIN users u2 
			WHERE u1.id > u2.id AND u1.username = u2.username
		`)
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
	// Migrate monitor type from string to int: 'snmp' -> 1, 'zabbix' -> 2, 'prometheus' -> 3 (now 'other')
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
			if err := database.DB.Exec("UPDATE monitors SET type = CASE WHEN LOWER(type) = 'snmp' THEN '1' WHEN LOWER(type) = 'zabbix' THEN '2' WHEN LOWER(type) = 'prometheus' THEN '3' ELSE '3' END WHERE type IS NOT NULL").Error; err != nil {
				return err
			}
		}
		// Migrate numeric types only if legacy markers are present to avoid re-mapping on every startup.
		// Legacy markers: type=4 (old SNMP) or type=1 on non-internal monitors.
		var legacyCount int64
		if err := database.DB.Model(&model.Monitor{}).
			Where("type = 4 OR (type = 1 AND id <> 1)").
			Count(&legacyCount).Error; err != nil {
			return err
		}
		if legacyCount > 0 {
			if err := database.DB.Exec("UPDATE monitors SET type = CASE WHEN type = 1 THEN 2 WHEN type = 2 THEN 3 WHEN type = 4 THEN 1 ELSE type END WHERE type IN (1, 2, 4)").Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func applySchemaUpdates() error {
	// Explicitly remove deprecated columns that AutoMigrate does not drop.
	if database.DB.Migrator().HasColumn(&model.Trigger{}, "action_id") {
		if err := database.DB.Migrator().DropColumn(&model.Trigger{}, "action_id"); err != nil {
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
