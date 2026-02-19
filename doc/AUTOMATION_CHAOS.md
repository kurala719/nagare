# Nagare Automation & Chaos Engineering

Nagare integrates configuration management and fault injection directly into the monitoring dashboard.

## 1. Ansible Automation Engine

Nagare's Ansible module allows for "Monitor-to-Action" closed-loop automation.

### 1.1 Architecture (`backend/internal/service/ansible.go`)
-   **Dynamic Inventory**: Nagare provides a `GET /ansible/inventory` endpoint that serves an Ansible-compatible JSON inventory of all `hosts` and `groups` in the Nagare database.
-   **Playbook Orchestration**: 
    -   Users can create/edit Ansible playbooks (`AnsiblePlaybook`) via the UI.
    -   Execution is offloaded to background workers using the `AnsibleJob` model.
-   **AI Integration**: A "Recommend Playbook" feature (`POST /ansible/playbooks/recommend`) uses Gemini to suggest a specific playbook based on the current alert's context.

### 1.2 Execution Lifecycle
1.  **Selection**: A user (or trigger) selects a playbook.
2.  **Job Creation**: A record is created in `ansible_jobs`.
3.  **Command Execution**: Nagare runs `ansible-playbook` with its own dynamic inventory as the source.
4.  **Log Capture**: Output is streamed and stored in the database for auditing and post-mortem analysis.

## 2. Chaos Engineering: Alert Storm Simulation

To test the robustness of alerting channels and LLM analysis, Nagare includes a "Chaos" module.

### 2.1 The "Alert Storm" Trigger (`backend/internal/api/chaos.go`)
-   **Functionality**: Simulates a high-intensity failure event by injecting dozens of critical alerts across multiple hosts in a very short window.
-   **Parameters**:
    -   `intensity`: Number of alerts per second.
    -   `target_group`: Specific infrastructure segment to affect.
-   **Purpose**: 
    -   Stress-test the **Rate Limiting** logic for QQ/Email notifications.
    -   Validate the **AI's ability to correlate** multiple simultaneous failures.

## 3. Automation Triggers (`internal/service/trigger.go`)
-   Nagare allows users to link monitoring `Items` to `Actions`.
-   **Auto-Remediation**: If an alert matches a trigger's condition (e.g., `Severity > 2`), it can automatically launch an Ansible playbook to restart the service or clear a cache.
