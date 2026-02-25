# Nagare Database Schema & Entities

Nagare uses **MySQL** (or any GORM-supported database) as its primary storage engine for all configuration, monitoring history, and state data. This document outlines the core entities and their relationships as defined in the system models.

---

## 🏗️ Core Entity ERD
- **Monitor** (1) → (N) **Groups**
- **Group** (1) → (N) **Hosts**
- **Host** (1) → (N) **Items** (Metrics)
- **Item** (1) → (N) **ItemHistory**
- **Host** (1) → (N) **Alerts**
- **Alarm** (1) → (N) **Alerts**
- **Trigger** (N) → (1) **Action** (via filter rules)
- **Action** (N) → (1) **Media**

---

## 📋 Entity Details

### 1. Infrastructure & Monitoring

#### `hosts` (`Host` struct)
Represents a monitored server, network device, or endpoint.
- **Fields**: `name`, `hostid` (external ID), `m_id` (Monitor ID), `group_id`, `description`, `enabled`, `status` (0=inactive, 1=active, 2=error, 3=syncing), `status_description`, `active_available`, `ip_addr`, `comment`, `ssh_user`, `ssh_password`, `ssh_port`, SNMP configs (`snmp_community`, `snmp_version`, `snmp_port`, `snmp_v3_*`), `last_sync_at`, `external_source`, `health_score`.

#### `groups` (`Group` struct)
Logical clusters of hosts.
- **Fields**: `name`, `description`, `m_id` (Monitor ID), `external_id`, `enabled`, `status`, `status_description`, `active_available`, `last_sync_at`, `external_source`, `health_score`.

#### `monitors` (`Monitor` struct)
External monitoring systems used as inventory sources.
- **Fields**: `name`, `url`, `username`, `password`, `auth_token`, `event_token`, `description`, `type` (1=snmp, 2=zabbix, 3=other), `enabled`, `status`, `status_description`, `health_score`.

#### `alarms` (`Alarm` struct)
External alert sources that push events into Nagare.
- **Fields**: `name`, `url`, `username`, `password`, `auth_token`, `event_token`, `description`, `type` (1=snmp, 2=zabbix, 3=other), `enabled`, `status`, `status_description`.

#### `items` (`Item` struct)
Specific metrics or log sources attached to a host.
- **Fields**: `name`, `hid` (Internal Host ID), `itemid` (External ID), `hostid` (External Host ID), `value_type`, `last_value`, `units`, `enabled`, `status`, `status_description`, `comment`, `last_sync_at`, `external_source`.

---

### 2. Time-Series & System History

#### `item_histories` (`ItemHistory` struct)
Tracks item metric values over time.
- **Fields**: `item_id`, `host_id`, `value`, `units`, `status`, `sampled_at`.

#### `host_histories` (`HostHistory` struct)
Tracks host status over time.
- **Fields**: `host_id`, `status`, `status_description`, `ip_addr`, `sampled_at`.

#### `network_status_histories` (`NetworkStatusHistory` struct)
Tracks overall network health over time.
- **Fields**: `score`, `monitor_total`, `monitor_active`, `group_total`, `group_active`, `group_impacted`, `host_total`, `host_active`, `item_total`, `item_active`, `sampled_at`.

---

### 3. Intelligence & AI

#### `providers` (`Provider` struct)
AI Model providers (e.g., Gemini, OpenAI) used for diagnostics and chat.
- **Fields**: `name`, `url`, `api_key`, `default_model`, `models` (JSON list), `type` (1=Gemini, 2=OpenAI), `description`, `enabled`, `status`.

#### `knowledge_bases` (`KnowledgeBase` struct)
Local documentation used for RAG (Retrieval-Augmented Generation) alert analysis.
- **Fields**: `topic`, `content`, `keywords`, `category`.

#### `chats` (`Chat` struct)
History of AI conversations.
- **Fields**: `user_id`, `provider_id`, `model`, `role` (user/assistant), `content`.

---

### 4. Alerting & Automation

#### `alerts` (`Alert` struct)
An active or resolved system incident.
- **Fields**: `message`, `severity`, `status` (0=active, 1=acknowledged, 2=resolved), `alarm_id`, `trigger_id`, `host_id`, `item_id`, `comment`.

#### `triggers` (`Trigger` struct)
Rules to fire actions based on specific alert, log, or item conditions.
- **Fields**: `name`, `entity` (alert/log), `severity`, `alert_id`, `alert_status`, `alert_group_id`, `alert_monitor_id`, `alert_host_id`, `alert_item_id`, `alert_query`, `log_type`, `log_level`, `log_query`, `item_status`, `item_value_threshold`, `item_value_threshold_max`, `item_value_operator`, `enabled`, `status`.

#### `actions` (`Action` struct)
Execution steps triggered by rules, sending messages via Media.
- **Fields**: `name`, `media_id`, `template`, `enabled`, `status`, `description`, `severity_min`, `trigger_id`, `host_id`, `group_id`, `alert_status`.

#### `ansible_playbooks` (`AnsiblePlaybook` struct)
YAML scripts used for automated remediation.
- **Fields**: `name`, `description`, `content` (YAML), `tags`.

#### `ansible_jobs` (`AnsibleJob` struct)
Execution history of Ansible playbooks.
- **Fields**: `playbook_id`, `status` (pending/running/success/failed), `output`, `triggered_by`, `host_filter`.

---

### 5. Notifications, Reports & Logs

#### `media` (`Media` struct)
Specific notification delivery targets.
- **Fields**: `name`, `type` (email, qq, etc.), `target`, `params` (JSON map), `enabled`, `status`, `description`.

#### `report_configs` (`ReportConfig` struct)
Configuration for automated report generation.
- **Fields**: `auto_generate_daily/weekly/monthly`, `generate_time`, `include_alerts/metrics`, `top_hosts_count`, `enable_llm_summary`, `email_notify`, `email_recipients`, `language`.

#### `reports` (`Report` struct)
Generated system reports (PDF).
- **Fields**: `report_type`, `title`, `file_path`, `download_url`, `status`, `generated_at`, `content_data`.

#### `site_messages` (`SiteMessage` struct)
Internal system notifications displayed in the browser.
- **Fields**: `title`, `content`, `type`, `severity`, `is_read`, `user_id`.

#### `log_entries` (`LogEntry` struct)
System and service log entries.
- **Fields**: `type`, `level` (0=info, 1=warn, 2=error), `message`, `context`, `user_id`, `ip`.

---

### 6. Security, Users & Compliance

#### `users` (`User` struct)
Authentication and user profiles.
- **Fields**: `username`, `password` (Hashed), `privileges` (1=user, 2=admin, 3=superadmin), `status`, `email`, `phone`, `avatar`, `address`, `introduction`, `nickname`, `qq`.

#### `audit_logs` (`AuditLog` struct)
Operational history of user actions.
- **Fields**: `user_id`, `username`, `action`, `method`, `path`, `ip`, `status`, `latency`, `user_agent`.

#### `retention_policies` (`RetentionPolicy` struct)
Configuration for system data pruning.
- **Fields**: `data_type` (logs/alerts/audit_logs/item_history/host_history), `retention_days`, `enabled`, `description`.

#### `register_applications` (`RegisterApplication` struct)
Pending registration requests.
- **Fields**: `username`, `password`, `email`, `status` (pending/approved/rejected), `reason`, `approved_by`.

#### `password_reset_applications` (`PasswordResetApplication` struct)
Requests to reset user passwords.
- **Fields**: `user_id`, `username`, `new_password`, `status`, `reason`, `approved_by`.

#### `qq_whitelists` (`QQWhitelist` struct)
Authorized QQ users/groups for commands and alerts.
- **Fields**: `qq_identifier`, `type` (0=user, 1=group), `nickname`, `can_command`, `can_receive`, `enabled`, `comment`.
