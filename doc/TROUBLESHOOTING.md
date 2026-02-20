# Nagare Troubleshooting: Fixing the Fixer

If Nagare isn't working as expected, follow these steps to diagnose and resolve common issues.

---

## 1. 🤖 AI & RAG Issues

### Q: The AI says "I don't know" or hallucinates.
**Cause**: The Knowledge Base is empty or irrelevant.
**Fix**:
1. Upload your server manuals or past incident reports to the Knowledge Base.
2. Ensure the file names and content contain relevant keywords (e.g., "Error 500", "MySQL Crash").
3. Check if the "RAG Context" is being retrieved correctly in the backend logs (`backend/logs/system.log`).

### Q: AI Analysis is timing out (30s).
**Cause**: Using a slow local model (like Llama 2 70b on CPU) or network latency to Gemini/OpenAI.
**Fix**:
1. Edit `configs/nagare_config.json`.
2. Increase `ai.analysis_timeout_seconds` to `60` or `120`.
3. Switch to a faster model (e.g., `gemini-1.5-flash`).

---

## 2. 🔌 Zabbix/Prometheus Webhook Failures

### Q: "Failed to parse alert" in Site Messages.
**Cause**: The JSON payload format sent by Zabbix doesn't match what Nagare expects.
**Fix**:
1. Check the **Audit Logs** for the failed `POST /api/v1/alerts/webhook` request.
2. Compare the payload with the documentation in `doc/INTEGRATIONS.md`.
3. Ensure the `event_token` is correct and matches the Monitor configuration in Nagare.

---

## 3. 🛠️ Ansible Automation Failures

### Q: "Permission denied (publickey)" during playbook execution.
**Host**: Target server (e.g., 192.168.1.50).
**Fix**:
1. Ensure the SSH key for the `nagare` user is added to `~/.ssh/authorized_keys` on the target server.
2. Verify the SSH user in Nagare's Host settings matches the user on the target machine.
3. Check file permissions: `chmod 600 ~/.ssh/id_rsa` on the Nagare server.

### Q: "Windows Ansible Error: chcp not found" or encoding issues.
**Platform**: Windows Host.
**Fix**:
1. Nagare uses `cmd /c chcp 65001` to force UTF-8 on Windows. Ensure your Windows environment supports this.
2. Consider using **WSL (Windows Subsystem for Linux)** for running Ansible on Windows control nodes.

---

## 4. 📉 Performance Issues

### Q: The dashboard is slow to load history charts.
**Cause**: Too much data in `item_histories` table (e.g., > 10 million rows).
**Fix**:
1. Enable data pruning in `configs/nagare_config.json` (feature coming soon).
2. Manually truncate old history:
   ```sql
   DELETE FROM item_histories WHERE sampled_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
   OPTIMIZE TABLE item_histories;
   ```

## 5. 📧 Email Notifications Not Sending

### Q: "SMTP Connection Timeout"
**Cause**: Firewall blocking port 25/587 or incorrect credentials.
**Fix**:
1. Verify SMTP settings in **Settings -> Notification Channels**.
2. Check if your cloud provider (AWS/GCP/Azure) blocks outbound port 25. Use port 587 (TLS) or 465 (SSL).
3. Check backend logs for detailed SMTP errors.
