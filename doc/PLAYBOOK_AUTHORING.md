# Nagare Ansible Guide: Writing Robot Scripts

Nagare uses **Ansible** to fix problems automatically. This guide explains how to write "Playbooks" that Nagare understands.

---

## 1. The Dynamic Inventory

Nagare automatically generates an Ansible Inventory file based on your current server list. You don't need to maintain a separate `hosts` file.

### How Nagare Groups Servers
- **`all`**: Every enabled server.
- **Group Names**: Normalized from your Nagare Groups (e.g., "Web Servers" -> `web_servers`).
- **`_meta`**: Contains host variables like SSH credentials.

### Example Inventory structure (JSON)
```json
{
  "all": {
    "hosts": ["web-01", "db-01"]
  },
  "web_servers": {
    "hosts": ["web-01"]
  },
  "_meta": {
    "hostvars": {
      "web-01": {
        "ansible_host": "192.168.1.10",
        "ansible_user": "root",
        "ansible_port": 22
      }
    }
  }
}
```

---

## 2. Writing a Playbook

Playbooks are standard YAML files. The only difference is that you should use **variable names** for hosts if you want flexibility.

### Basic Example: Restart Nginx
```yaml
---
- name: Restart Web Server
  hosts: all  # Can be overridden at runtime
  become: yes
  tasks:
    - name: Restart Nginx service
      service:
        name: nginx
        state: restarted
```

### Advanced Example: Disk Cleanup
```yaml
---
- name: Clean Disk Space
  hosts: all
  tasks:
    - name: Clean apt cache
      apt:
        autoclean: yes
      when: ansible_os_family == "Debian"

    - name: Remove old log files (> 7 days)
      find:
        paths: /var/log
        age: 7d
        recurse: yes
      register: files_to_delete

    - name: Delete files
      file:
        path: "{{ item.path }}"
        state: absent
      with_items: "{{ files_to_delete.files }}"
```

---

## 3. Using AI to Write Playbooks

Don't want to write YAML? Ask the AI.

1. Go to **Automation -> Playbooks -> New**.
2. Click **"Ask AI"**.
3. Type: *"Write a playbook to install Docker on Ubuntu."*
4. The AI will generate the YAML for you. **Always review it before saving.**

---

## 4. Running Playbooks

### Manual Run
1. Go to **Automation -> Playbooks**.
2. Click **Run** on a playbook.
3. Select the target: **All Hosts** or a specific **Group** (e.g., `db_servers`).

### Automated Trigger (Self-Healing)
1. Go to **Automation -> Triggers**.
2. Create a new Trigger:
   - **If**: Alert contains "High CPU" AND Severity > 2.
   - **Then**: Run Playbook "Restart Service".
   - **Host**: The host that triggered the alert.

---

## 5. Troubleshooting Execution
- **"Failed to connect"**: Check if the Nagare server has SSH access to the target. Verify keys/passwords in the Host settings.
- **"Permission denied"**: Ensure `become: yes` is used if root privileges are needed.
- **Windows Targets**: Nagare supports limited Windows automation via WinRM, but Linux is the primary target.
