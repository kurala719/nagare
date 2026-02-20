# Nagare Database Schema & Entities

Nagare uses **MySQL** (or any GORM-supported database) as its primary storage engine for all configuration, monitoring history, and state data. This document outlines the core entities and their relationships.

---

## 🏗️ Core Entity ERD
- **Monitor** (1) → (N) **Groups**
- **Group** (1) → (N) **Hosts**
- **Host** (1) → (N) **Items** (Metrics)
- **Item** (1) → (N) **ItemHistory**
- **Host** (1) → (N) **Alerts**
- **Alarm** (1) → (N) **Alerts**
- **Trigger** (N) → (1) **Action**
- **Action** (N) → (1) **Media**

---

## 📋 Entity Details

### 1. Infrastructure & Monitoring

#### `hosts`
Represents a monitored server, network device, or endpoint.
- **Fields**: `name`, `hostid` (external ID), `m_id` (foreign key to Monitor), `group_id`, `description`, `enabled`, `status` (0=inactive, 1=active, 2=error, 3=syncing), `ip_addr`, `ssh_user`, `ssh_password`, `ssh_port`, SNMP configs (`snmp_community`, `snmp_version`, `snmp_v3_*`), `health_score`.

#### `groups`
Logical clusters of hosts.
- **Fields**: `name`, `description`, `m_id` (foreign key to Monitor), `external_id`, `enabled`, `status`, `health_score`.

#### `monitors`
External monitoring systems used as inventory sources (e.g., Zabbix, Prometheus).
- **Fields**: `name`, `url`, `username`, `password`, `auth_token`, `event_token`, `type` (1=zabbix, 2=prometheus, 3=other), `enabled`, `status`, `health_score`.

#### `alarms`
External alert sources that push events into Nagare.
- **Fields**: `name`, `url`, `username`, `password`, `auth_token`, `event_token`, `type`, `enabled`, `status`.

#### `items`
Specific metrics or log sources attached to a host.
- **Fields**: `name`, `hid` (Host ID), `itemid` (external ID), `value_type`, `last_value`, `units`, `enabled`, `status`.

---

### 2. Time-Series History

#### `item_histories`
Tracks item metric values over time.
- **Fields**: `item_id`, `host_id`, `value`, `units`, `status`, `sampled_at`.

#### `host_histories`
Tracks host status over time.
- **Fields**: `host_id`, `status`, `status_description`, `ip_addr`, `sampled_at`.

#### `network_status_histories`
Tracks overall network health over time.
- **Fields**: `score`, `monitor_total`, `monitor_active`, `group_total`, `host_total`, `item_total`, `sampled_at`.

---

### 3. Intelligence & AI

#### `providers`
AI Model providers (e.g., Gemini, OpenAI) used for diagnostics and chat.
- **Fields**: `name`, `url`, `api_key`, `default_model`, `models` (JSON array), `type` (1=Gemini, 2=OpenAI), `enabled`, `status`.

#### `knowledge_bases`
Local documentation used for RAG (Retrieval-Augmented Generation) alert analysis.
- **Fields**: `topic`, `content`, `keywords`, `category`.

#### `chats`
History of AI conversations.
- **Fields**: `user_id`, `provider_id`, `model`, `role` (user/assistant), `content`.

---

### 4. Alerting & Automation

#### `alerts`
An active or resolved system incident.
- **Fields**: `message`, `severity`, `status` (0=active, 1=acknowledged, 2=resolved), `alarm_id`, `host_id`, `item_id`, `comment`.

#### `ansible_playbooks`
YAML scripts used for automated remediation.
- **Fields**: `name`, `description`, `content` (YAML), `tags`.

#### `ansible_jobs`
Execution history of Ansible playbooks.
- **Fields**: `playbook_id`, `status` (pending/running/success/failed), `output`, `triggered_by`, `host_filter`.

#### `triggers`
Rules to fire actions based on specific alert or log conditions.
- **Fields**: `name`, `entity` (alert/log), `severity_min`, `action_id`, `alert_query`, `log_query`, `enabled`, `status`.

#### `actions`
Execution steps triggered by rules, sending messages via Media.
- **Fields**: `name`, `media_id`, `template`, `enabled`, `status`.

---

### 5. Notifications & Delivery

#### `media`
Specific notification delivery targets (e.g., "Admin Email", "Ops Webhook").
- **Fields**: `name`, `type` (gmail, webhook, qq, sms), `target`, `params` (JSON mapping), `enabled`, `status`.

#### `site_messages`
Internal system notifications displayed in the browser.
- **Fields**: `title`, `content`, `type`, `severity`, `is_read`, `user_id`.

#### `qq_whitelists`
Authorized QQ users/groups allowed to interact with the system bot.
- **Fields**: `qq_identifier`, `type` (0=user, 1=group), `nickname`, `can_command`, `can_receive`, `enabled`.

---

### 6. Security, Users, & Audit

#### `users`
Authentication and basic user profiles.
- **Fields**: `username`, `password` (Hashed), `privileges` (1=user, 2=admin, 3=superadmin), `status`, `email`, `phone`, `avatar`, `nickname`.

#### `audit_logs`
Operational history of user actions for compliance.
- **Fields**: `user_id`, `username`, `action`, `method`, `path`, `ip`, `status`, `latency`, `user_agent`.

#### `retention_policies`
Configuration for system data pruning.
- **Fields**: `data_type` (logs/alerts/audit_logs/item_history), `retention_days`, `enabled`.
