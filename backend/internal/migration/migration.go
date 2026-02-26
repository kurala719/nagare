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

func fixForeignKeyColumnTypes() error {
	tablesToFix := map[string][]string{
		"hosts":                         {"m_id", "group_id"},
		"groups":                        {"m_id"},
		"items":                         {"hid"},
		"item_histories":                {"item_id", "host_id"},
		"host_histories":                {"host_id"},
		"alerts":                        {"alarm_id", "trigger_id", "host_id", "item_id"},
		"actions":                       {"media_id", "trigger_id", "host_id", "group_id"},
		"triggers":                      {"alert_id", "alert_group_id", "alert_monitor_id", "alert_host_id", "alert_item_id"},
		"chats":                         {"user_id", "provider_id"},
		"log_entries":                   {"user_id"},
		"audit_logs":                    {"user_id"},
		"register_applications":         {"approved_by"},
		"password_reset_applications":   {"user_id", "approved_by"},
		"ansible_jobs":                  {"playbook_id", "triggered_by"},
		"site_messages":                 {"user_id"},
		"packet_analyses":               {"provider_id", "user_id"},
	}

	for table, columns := range tablesToFix {
		if !database.DB.Migrator().HasTable(table) {
			continue
		}
		for _, column := range columns {
			if database.DB.Migrator().HasColumn(table, column) {
				// Check current type to avoid redundant ALTER
				columnTypes, err := database.DB.Migrator().ColumnTypes(table)
				if err != nil {
					continue
				}
				var isBigIntUnsigned bool
				for _, ct := range columnTypes {
					if ct.Name() == column {
						dbType := strings.ToUpper(ct.DatabaseTypeName())
						// Check for BIGINT and UNSIGNED. MySQL returns 'BIGINT UNSIGNED' or similar
						isBigIntUnsigned = strings.Contains(dbType, "BIGINT") && strings.Contains(dbType, "UNSIGNED")
						break
					}
				}

				if !isBigIntUnsigned {
					fmt.Printf(">>> Migrating column %s in table %s to BIGINT UNSIGNED...\n", column, table)
					// 1. Ensure it's BIGINT UNSIGNED
					_ = database.DB.Exec(fmt.Sprintf("ALTER TABLE `%s` MODIFY COLUMN `%s` BIGINT UNSIGNED", table, column))

					// 2. Clean up invalid references (orphans) before adding FK
					// Only run this when we actually change the type (usually first time)
					refTable := ""
					switch column {
					case "group_id", "alert_group_id":
						refTable = "groups"
					case "m_id", "alert_monitor_id":
						refTable = "monitors"
					case "hid", "host_id", "alert_host_id":
						refTable = "hosts"
					case "item_id", "alert_item_id":
						refTable = "items"
					case "user_id", "approved_by", "triggered_by":
						refTable = "users"
					case "alarm_id":
						refTable = "alarms"
					case "trigger_id":
						refTable = "triggers"
					case "media_id":
						refTable = "media"
					case "provider_id":
						refTable = "providers"
					case "playbook_id":
						refTable = "ansible_playbooks"
					}

					if refTable != "" && database.DB.Migrator().HasTable(refTable) {
						fmt.Printf(">>> Cleaning up orphans for %s.%s referencing %s...\n", table, column, refTable)
						// 1. Explicitly set 0 to NULL first to avoid FK issues
						_ = database.DB.Exec(fmt.Sprintf("UPDATE `%s` SET `%s` = NULL WHERE `%s` = 0", table, column, column))

						// 2. Set to NULL where the referenced ID doesn't exist
						_ = database.DB.Exec(fmt.Sprintf("UPDATE `%s` SET `%s` = NULL WHERE `%s` NOT IN (SELECT id FROM `%s`) AND `%s` IS NOT NULL", table, column, column, refTable, column))
					}
				}
			}
		}
	}
	return nil
}

func migrateSeverityLevels() error {
	// Check if we already migrated. We use a config flag or check if values shifted.
	// For simplicity, we'll shift them once.
	// Old: 0:Info, 1:Low, 2:Medium, 3:High, 4:Critical
	// New: 0:Not Classified, 1:Info, 2:Warning, 3:Average, 4:High, 5:Disaster
	
	// We'll increment existing 0-4 values by 1 to roughly match new mapping
	// Info (0) -> Info (1)
	// Low (1) -> Warning (2)
	// Medium (2) -> Average (3)
	// High (3) -> High (4)
	// Critical (4) -> High (4) or Disaster (5)? 
	// Let's just do a simple shift for now if they are in range [0, 4]
	
	// Only run if Disaster (5) is not used yet? Or use a marker.
	// We'll check if any Trigger has severity 5.
	var count int64
	database.DB.Model(&model.Trigger{}).Where("severity = 5").Count(&count)
	if count > 0 {
		return nil // Already migrated or has disaster triggers
	}

	tables := []string{"triggers", "alerts"}
	for _, t := range tables {
		if database.DB.Migrator().HasTable(t) {
			_ = database.DB.Exec(fmt.Sprintf("UPDATE `%s` SET severity = severity + 1 WHERE severity >= 0 AND severity <= 4", t))
		}
	}
	
	// Actions use severity_min
	if database.DB.Migrator().HasTable("actions") {
		_ = database.DB.Exec("UPDATE actions SET severity_min = severity_min + 1 WHERE severity_min >= 0 AND severity_min <= 4")
	}

	return nil
}

func preSchemaUpdates() error {
	// Fix column types for foreign keys to ensure they are compatible (BIGINT UNSIGNED)
	if err := fixForeignKeyColumnTypes(); err != nil {
		fmt.Printf("Warning: failed to fix foreign key column types: %v\n", err)
	}

	// Deduplicate users before adding unique index if table exists
	if database.DB.Migrator().HasTable("users") {
		// This query finds duplicates and deletes all but the one with the smallest ID
		_ = database.DB.Exec(`
			DELETE u1 FROM users u1
			INNER JOIN users u2 
			WHERE u1.id > u2.id AND u1.username = u2.username
		`)
	}

	if err := migrateSeverityLevels(); err != nil {
		fmt.Printf("Warning: failed to migrate severity levels: %v\n", err)
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
		// Only run if there are non-numeric values
		if err := database.DB.Exec("UPDATE triggers SET log_level = CASE WHEN CAST(log_level AS CHAR) IN ('info','warn','warning','error') THEN CASE CAST(log_level AS CHAR) WHEN 'info' THEN 0 WHEN 'warn' THEN 1 WHEN 'warning' THEN 1 WHEN 'error' THEN 2 END WHEN CAST(log_level AS CHAR) REGEXP '^[0-9]+$' THEN CAST(log_level AS UNSIGNED) ELSE NULL END WHERE log_level IS NOT NULL AND (CAST(log_level AS CHAR) REGEXP '^[a-zA-Z]+$')").Error; err != nil {
			return err
		}
	}
	if database.DB.Migrator().HasTable(&model.LogEntry{}) && database.DB.Migrator().HasColumn(&model.LogEntry{}, "level") {
		// Only run if there are non-numeric values
		if err := database.DB.Exec("UPDATE log_entries SET level = CASE WHEN CAST(level AS CHAR) IN ('info','warn','warning','error') THEN CASE CAST(level AS CHAR) WHEN 'info' THEN 0 WHEN 'warn' THEN 1 WHEN 'warning' THEN 1 WHEN 'error' THEN 2 END WHEN CAST(level AS CHAR) REGEXP '^[0-9]+$' THEN CAST(level AS UNSIGNED) ELSE NULL END WHERE level IS NOT NULL AND (CAST(level AS CHAR) REGEXP '^[a-zA-Z]+$')").Error; err != nil {
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
		// Check current type to avoid redundant ALTER
		columnTypes, err := database.DB.Migrator().ColumnTypes("log_entries")
		if err == nil {
			var isInt bool
			for _, ct := range columnTypes {
				if ct.Name() == "level" {
					dbType := strings.ToUpper(ct.DatabaseTypeName())
					isInt = strings.Contains(dbType, "INT") && !strings.Contains(dbType, "BIGINT")
					break
				}
			}
			if !isInt {
				fmt.Println(">>> Migrating log_entries.level to INT...")
				if err := database.DB.Exec("ALTER TABLE log_entries MODIFY COLUMN level INT").Error; err != nil {
					return err
				}
			}
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
