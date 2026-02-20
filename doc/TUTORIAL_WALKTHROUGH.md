# Nagare Tutorials: From Beginner to Expert

This guide provides step-by-step walkthroughs for the most common and powerful workflows in Nagare.

---

## 튜토리얼 1: Setting Up Your First AI-Powered Monitor 🧠
*Goal: Connect a server and see a diagnosis for a simulated crash.*

### Step 1: Add your AI Provider
1. Navigate to **Settings > AI Providers**.
2. Click **Add Provider**.
3. Choose **Google Gemini** (Type 1).
4. Enter your API Key and set the default model to `gemini-1.5-pro`.
5. Click **Check Status** to ensure Nagare can "talk" to the brain.

### Step 2: Add a Host
1. Go to **Infrastructure > Hosts**.
2. Click **Add Host**.
3. Enter Name: `Production-Web-01`, IP: `192.168.1.100`.
4. (Optional) Provide SSH credentials for WebSSH access.
5. Save the host.

### Step 3: Simulate an Alert
1. Go to **Settings > Monitors**.
2. Create a "Test Monitor" or use the API to send a manual alert:
```bash
curl -X POST http://localhost:8080/api/v1/alerts/webhook 
     -H "Content-Type: application/json" 
     -d '{
       "message": "Nginx process is down on Production-Web-01",
       "severity": 2,
       "host_id": 1
     }'
```
3. Go to the **Alerts** dashboard. You will see the new alert.
4. Click the **"Consult AI"** button. The AI will analyze the message and tell you that Nginx has likely crashed due to a configuration error or resource exhaustion.

---

## 튜토리얼 2: From Alert to Auto-Resolution 🤖
*Goal: Automatically restart a service when it fails.*

### Step 1: Create the "Fix-It" Playbook
1. Go to **Automation > Playbooks**.
2. Click **New Playbook**.
3. Title: `Restart Nginx`.
4. Content:
```yaml
---
- name: Restart Nginx
  hosts: all
  become: yes
  tasks:
    - name: Ensure nginx is running
      service:
        name: nginx
        state: restarted
```
5. Save the playbook.

### Step 2: Configure the Trigger
1. Go to **Automation > Triggers**.
2. Click **Add Trigger**.
3. **If**: 
   - Entity: `alert`
   - Message Contains: `Nginx process is down`
   - Severity >= `2`
4. **Then**:
   - Action: `Run Ansible Playbook`
   - Playbook: `Restart Nginx`
5. Save the trigger.

### Step 3: Test the "Self-Healing"
Send the same alert as in Tutorial 1. Nagare will:
1. Detect the alert.
2. Match it to the Trigger.
3. Automatically launch an Ansible Job.
4. Notify you via Site Message: *"Self-healing initiated: Restarting Nginx on Production-Web-01."*

---

## 튜토리얼 3: Advanced Knowledge Management 📚
*Goal: Teach the AI about your company's unique internal errors.*

### Step 1: Document an "Internal Mystery"
Suppose your database throws a weird error like `Error 999: Flux Capacitor Mismatch`.
1. Go to **Knowledge Base > Add Entry**.
2. **Topic**: `Flux Capacitor Mismatch (Error 999)`.
3. **Content**: 
   "This error happens when the backup drive is 99% full and the cron job tries to rotate logs. To fix: Run /opt/scripts/cleanup.sh and then restart the DB."
4. **Keywords**: `999, flux, mismatch, database`.
5. **Category**: `Database`.

### Step 2: Verify RAG Retrieval
Next time an alert with the word "999" or "Flux" arrives:
1. Click **Consult AI**.
2. You will see a section in the AI response: **"Relevant Knowledge Base Found"**.
3. The AI will now say: *"Based on your internal documentation, you should run the cleanup script..."* instead of giving a generic answer.

---

## 튜토리얼 4: Using the "Roast" Mode for Audits 🌶️
*Goal: Get a critical, expert-level audit of your system status.*

1. Go to the **AI Chat** interface.
2. Switch the **Mode** selector to **"Roast"**.
3. Type: *"What do you think of my current system health?"*
4. Nagare will gather all current alerts and metrics and give you a sarcastic but highly accurate critique.
   *Example: "Your health score is 70. Is this a server room or a digital graveyard? You have 5 critical alerts on the DB-Master that have been sitting there for 3 hours. Maybe try fixing things instead of chatting with me?"*
5. Use this "tough love" to identify neglected areas of your infrastructure.
