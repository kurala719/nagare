# API Group: Ansible Automation

The Ansible API group provides endpoints for managing automation tasks, playbooks, and real-time execution monitoring on your infrastructure.

---

## 🏗️ 1. Inventory Management

### **GET** `/api/v1/ansible/inventory`
Generates a dynamic Ansible inventory in JSON format compatible with the `ansible-playbook -i` command.
- **Access**: Public (Internal use by Ansible runner)
- **Output**: A standard Ansible JSON inventory containing `all`, individual groups (e.g., `web_servers`), and `_meta` host variables (IP, port, user).

---

## 📋 2. Playbook Management

### **GET** `/api/v1/ansible/playbooks`
Lists all stored Ansible playbooks.
- **Access**: Privilege Level 2+
- **Parameters**: 
  - `q` (Query): Optional string to search by name or description.
- **Response**: List of playbook objects.

### **POST** `/api/v1/ansible/playbooks`
Creates a new playbook.
- **Access**: Privilege Level 2+
- **Body**:
  ```json
  {
    "name": "Restart Nginx",
    "description": "Restarts the nginx service on Ubuntu/Debian",
    "content": "--- 
- hosts: all
  tasks:
    - name: restart
      service: name=nginx state=restarted",
    "tags": "web,nginx"
  }
  ```

### **GET** `/api/v1/ansible/playbooks/:id`
Retrieves the details and content of a specific playbook.

### **PUT** `/api/v1/ansible/playbooks/:id`
Updates an existing playbook's metadata or content.

### **DELETE** `/api/v1/ansible/playbooks/:id`
Removes a playbook from the system.

---

## 🚀 3. Execution & AI Recommendations

### **POST** `/api/v1/ansible/playbooks/:id/run`
Triggers the execution of a playbook.
- **Access**: Privilege Level 2+
- **Body**:
  ```json
  {
    "host_filter": "web_servers" // Optional: filter hosts by name or group
  }
  ```
- **Response**: Returns a `job_id` which can be used to track progress.

### **POST** `/api/v1/ansible/playbooks/recommend`
Uses the AI to generate or recommend a playbook based on a textual description or error message.
- **Access**: Privilege Level 2+
- **Body**: `{ "context": "I need to clear the /tmp folder on all linux servers" }`
- **Response**: `{ "content": "--- 
- name: clear tmp..." }`

---

## 📊 4. Job Monitoring

### **GET** `/api/v1/ansible/jobs`
Lists previous and currently running Ansible jobs.
- **Parameters**:
  - `playbook_id`: Filter by a specific playbook.
  - `limit`: Number of jobs to return (Default: 20).

### **GET** `/api/v1/ansible/jobs/:id`
Retrieves the full output and final status (`success`, `failed`, `running`) of a specific job.

---

## 💡 Real-time Streaming
When a job is running, Nagare streams the logs via the Global WebSocket Hub. Listen for the `ansible_log` event to display a real-time terminal output in the UI.
