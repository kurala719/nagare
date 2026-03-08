# Nagare (流): The Complete Manual for Humans 🚀

Nagare is an AI-powered "Brain" for your IT infrastructure. This manual explains how all its parts work together, from monitoring sensors to AI-driven diagnostics.

---

## 🗺️ The Map of Nagare

Nagare is divided into six main "Functional Zones" that you will see in your dashboard:

### 1. 🧠 The Intelligence Zone (AI & Knowledge)
This is where the "Thinking" happens.
- **AI Consulting**: Specialized buttons on Alerts and Hosts that ask the AI to diagnose a specific problem.
- **Knowledge Base**: Your private library. Upload PDF manuals or text notes so the AI knows your specific environment.
- **AI Chat**: A direct line to the AI for general technical questions.
- **MCP (Model Context Protocol)**: A "Universal Language" that lets other AIs (like Claude or ChatGPT) talk directly to Nagare.

### 2. 🖥️ The Infrastructure Zone (Hosts & Groups)
The "Digital Assets" Nagare is watching.
- **Hosts**: Individual servers. You can see their history, current health, and connect via **WebSSH**.
- **Groups**: Logical clusters (e.g., "Web Servers", "Database Team").
- **Synchronization**: Nagare can "Pull" data from Zabbix, so you don't have to add servers manually.

### 3. 🔔 The Alerting Zone (Sensors)
The "Security Guard" of your system.
- **Alerts**: Real-time notifications when something goes wrong.
- **Webhook Ingest**: A universal "Ear" that listens to SOS signals from any other monitoring tool.
- **Trigger Rules**: "If the CPU is too high, then send a message and run a fix script."

### 4. 🤖 The Automation Zone (Muscle)
The "Robot Technicians" that fix things.
- **Ansible Playbooks**: Pre-written scripts to fix common issues (e.g., "Restart Nginx").
- **AI Recommendations**: The AI can suggest the best script to run based on the error it sees.
- **Chaos Engineering**: "Fire Drills" where you intentionally trigger an alert storm to see if your team is ready.

### 5. 📊 The Analytics Zone (Review)
The "Health Report Card."
- **PDF Reporting**: Automated weekly or monthly summaries with AI-generated "Executive Summaries."
- **Alert Analytics**: Charts showing if your system is getting more stable or more chaotic over time.

### 6. 💬 The Communication Zone (Notifications)
How Nagare talks to you.
- **Site Messages**: Real-time pop-up notifications in your browser.
- **IM & QQ Integration**: Get alerts directly in your chat apps.
- **Whitelists**: Security controls to ensure only authorized people can talk to the Nagare Bot.

---

## 🔑 Understanding "Privilege Levels"
Nagare uses a simple 1-2-3 system for security:
- **Level 1 (User)**: Can view dashboards, use AI Chat, and open WebSSH.
- **Level 2 (Manager)**: Can add servers, upload to the Knowledge Base, and run Automation scripts.
- **Level 3 (Admin)**: Can change system settings, approve new users, and manage the "Brain's" core config.

---

## 📚 Deep Dive Manuals
Explore the specific details of each zone:

- [**Tutorial Walkthrough**](./doc/TUTORIAL_WALKTHROUGH.md) - Step-by-step guides.
- [**AI & RAG Engine**](./doc/AI_RAG_ENGINE.md) - How the AI thinks.
- [**AI Configuration**](./doc/AI_CONFIGURATION.md) - Setting up Gemini/OpenAI.
- [**Automation & Chaos**](./doc/AUTOMATION_CHAOS.md) - Fire Drills and Robot Assistants.
- [**Playbook Authoring**](./doc/PLAYBOOK_AUTHORING.md) - Writing Ansible scripts.
- [**WebSSH & Security**](./doc/WEBSSH_SECURITY.md) - Connecting to servers safely.
- [**Security & RBAC**](./doc/RBAC_SECURITY_MODEL.md) - Permissions and Access Control.
- [**Reporting & Analytics**](./doc/REPORTING_SYSTEM.md) - Making PDF reports.
- [**Communication & IM**](./doc/COMMUNICATION_NOTIFICATIONS.md) - Site messages and QQ bots.
- [**Database Schema**](./doc/DATABASE_SCHEMA.md) - The Storage Engine.
- [**Deployment Guide**](./doc/DEPLOYMENT_GUIDE.md) - Production & Staging.
- [**Developer Guide**](./doc/DEVELOPER_GUIDE.md) - Coding Standards.
- [**Troubleshooting**](./doc/TROUBLESHOOTING.md) - Common Errors & Fixes.
- [**Integrations**](./doc/INTEGRATIONS.md) - Zabbix Integration.
- [**Architecture & API**](./doc/ARCHITECTURE.md) - How it's built (For Techies).
