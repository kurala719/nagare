# Nagare Automation & Chaos Engineering

This manual details the closed-loop remediation and simulation capabilities of Nagare.

## 1. Ansible Integration Engine

Nagare provides native integration with Ansible for automated remediation.

### 1.1 Dynamic Inventory Service
Nagare serves a dynamic JSON inventory at `GET /api/v1/ansible/inventory`.
- **Structure**: Maps Nagare `Groups` to Ansible `Groups` and Nagare `Hosts` to Ansible `Hosts`.
- **Credential Injection**: Automatically attaches configured SSH credentials to the dynamic inventory output for Ansible consumption.

### 1.2 Playbook Management
- **Persistence**: Playbooks are stored as raw text in MySQL.
- **Async Execution**: Triggering a playbook creates an `AnsibleJob`, executed via a background worker.
- **Streaming Logs**: Job output is captured and available for real-time monitoring via the UI.

## 2. Chaos Engineering: Alert Storm Simulator

Nagare includes a built-in stress-test tool to validate SRE responsiveness and AI correlation.

### 2.1 Logic (`internal/api/chaos.go`)
The Alert Storm simulator mimics a cascading infrastructure failure:
- **Injection**: Creates multiple critical `Alert` records across a targeted `Group` within a 1-second window.
- **Validation Path**:
    1. Tests **Notification Rate Limiting** (preventing QQ/Email spam).
    2. Tests **AI Correlation** (Gemini's ability to identify a single root cause from 50+ concurrent alerts).

## 3. Automation Triggers
- **Event-Action Mapping**: Users define triggers (e.g., `CPU > 90%`) linked to `Actions` (e.g., `Run Ansible Playbook: Restart-App`).
- **State Logic**: Triggers include cooldown periods to prevent remediation loops.
