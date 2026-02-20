# Nagare Database Schema: The Storage Engine

Nagare uses **MySQL** (or any GORM-supported database) to store its knowledge, history, and configuration. This document explains the core tables and how they relate to each other.

---

## 🏗️ Core Entity ERD (Simplified)
- **Monitor** (1) → (N) **Groups**
- **Group** (1) → (N) **Hosts**
- **Host** (1) → (N) **Items** (Metrics)
- **Item** (1) → (N) **ItemHistory**
- **Host** (1) → (N) **Alerts**

---

## 📋 Table Details

### 1. Monitoring Infrastructure
| Table | Description | Key Columns |
| :--- | :--- | :--- |
| `monitors` | External monitoring systems (e.g. Zabbix) | `url`, `auth_token`, `event_token`, `type` |
| `groups` | Logical clusters of hosts | `name`, `external_id`, `m_id` (Monitor ID) |
| `hosts` | Individual servers | `name`, `ip_addr`, `ssh_user`, `ssh_password`, `status` |
| `items` | Specific metrics or log sources | `name`, `hid` (Host ID), `last_value`, `units` |

### 2. Intelligence & Knowledge
| Table | Description | Key Columns |
| :--- | :--- | :--- |
| `providers` | AI Model providers (Gemini/OpenAI) | `api_key`, `type` (1=Gemini, 2=OpenAI), `models` |
| `knowledge_bases` | Local RAG documentation | `topic`, `content`, `keywords`, `category` |
| `chats` | History of AI conversations | `user_id`, `model`, `role`, `content` |

### 3. Automation & Control
| Table | Description | Key Columns |
| :--- | :--- | :--- |
| `ansible_playbooks` | YAML fix scripts | `name`, `content` (YAML), `tags` |
| `ansible_jobs` | Execution history of scripts | `playbook_id`, `status`, `output`, `host_filter` |
| `triggers` | Rules to fire actions | `entity` (alert/log), `severity_min`, `action_id` |

### 4. Security & Users
| Table | Description | Key Columns |
| :--- | :--- | :--- |
| `users` | Auth information | `username`, `password` (Hashed), `privileges` (1-3) |
| `user_informations` | User profiles | `email`, `nickname`, `avatar`, `user_id` |
| `audit_logs` | Operational history | `action`, `path`, `ip`, `latency`, `user_id` |

### 5. Notifications
| Table | Description | Key Columns |
| :--- | :--- | :--- |
| `media_types` | Delivery methods (Email/Webhook) | `key`, `template`, `fields` (JSON) |
| `media` | Specific targets (e.g. "Admin Email") | `target`, `params` (JSON), `media_type_id` |
| `qq_whitelists` | Authorized QQ IDs | `qq_identifier`, `type` (User/Group), `can_command` |

---

## 🔍 Data Retention & Performance
- **History Tables**: `item_histories` and `host_histories` can grow very large. Nagare is designed to handle time-series data, but periodic pruning is recommended for performance.
- **Indexes**: Critical indexes exist on `sampled_at`, `host_id`, and `item_id` to ensure dashboard charts load in milliseconds.
- **JSON Fields**: Nagare uses JSON columns for flexible configuration (e.g., `Media.Params`), allowing for new notification types without changing the database structure.
