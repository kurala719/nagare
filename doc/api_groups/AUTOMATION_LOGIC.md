# API Group: Automation Logic & Chaos

This group manages the "If-Then" rules (Triggers) and the "Execution" (Actions) that enable self-healing infrastructure.

---

## ⚡ 1. Trigger Management

Triggers are the "Filters" that decide when to take action.

### **GET** `/api/v1/triggers`
Searches and lists trigger rules.
- **Parameters**: `q` (search), `severity_min`, `entity` (alert/log).

### **POST** `/api/v1/triggers`
Creates a new rule.
- **Body**:
  ```json
  {
    "name": "Auto-Fix Disk",
    "entity": "alert",
    "severity_min": 2,
    "alert_query": "Disk space low",
    "action_id": 5
  }
  ```

---

## 🎬 2. Action Management

Actions are the "Methods" used to respond (e.g., Run a Playbook, Send Email).

### **GET** `/api/v1/actions`
Lists available actions.

### **POST** `/api/v1/actions`
Defines a new action.
- **Fields**: `media_id` (Link to notification channel), `template` (Message template).

---

## 🔥 3. Chaos Engineering

### **POST** `/api/v1/chaos/alert-storm`
Simulates a disaster scenario by injecting 200 random high-severity alerts into the system over 5 seconds.
- **Use Case**: Testing if the AI can correctly identify the "Root Cause" in a noise storm, and verifying that the Global Hub handles high message throughput.

---

## 🔄 4. Execution Flow
1. **Alert Ingest**: Alert arrives at `/webhook`.
2. **Trigger Match**: System checks all `Triggers`.
3. **Action Invoke**: If a match is found, the linked `Action` is executed.
4. **Log**: Every execution is recorded in the Audit Logs for accountability.
