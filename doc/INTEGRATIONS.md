# Nagare Integrations: Connecting Your Monitors

Nagare is designed to be the "Brain" on top of existing monitoring systems. This guide explains how to connect Zabbix, Prometheus, and any other tool to the Nagare Brain.

---

## 1. Universal Webhook Ingest
The main entry point for all external alerts is:
`POST /api/v1/alerts/webhook`

Nagare automatically maps the incoming JSON to its internal alert system.

---

## 2. Zabbix Integration 🟦

### Step 1: Add a Monitor in Nagare
1. Go to **Monitors** → **Add Monitor**.
2. **Type**: Zabbix (1).
3. **URL**: `http://your-zabbix/zabbix/api_jsonrpc.php`.
4. **Auth**: Username/Password.
5. Nagare will sync hosts, groups, and items from Zabbix.

### Step 2: Configure Zabbix Media Type (SOS)
1. In Zabbix: **Alerts** → **Media Types** → **Create**.
2. **Type**: Webhook.
3. **Parameters**:
   - `URL`: `http://nagare-server:8080/api/v1/alerts/webhook`
   - `EventToken`: (Get this from the Nagare Monitor details page).
   - `Message`: `{ALERT.MESSAGE}`
   - `Subject`: `{ALERT.SUBJECT}`
   - `HostID`: `{HOST.ID}`
4. **Script**:
```javascript
var request = new HttpRequest();
request.addHeader('Content-Type: application/json');
var params = JSON.parse(value);
request.post(params.URL, JSON.stringify(params));
return 'OK';
```

---

## 3. Prometheus Integration 🟧

### Alertmanager Configuration
Add a `webhook_config` to your `alertmanager.yml`:

```yaml
receivers:
- name: 'nagare-webhook'
  webhook_configs:
  - url: 'http://nagare-server:8080/api/v1/alerts/webhook'
    http_config:
      bearer_token: "YOUR_EVENT_TOKEN"
```

---

## 4. Custom Scripts (Bash/Python/PowerShell)
You can send custom alerts from any script:

```bash
curl -X POST http://nagare-server:8080/api/v1/alerts/webhook 
     -H "Content-Type: application/json" 
     -d '{
       "message": "Critical backup failure on Server-04",
       "severity": 2,
       "event_token": "YOUR_TOKEN_HERE"
     }'
```

---

## 5. Synchronizing Hosts & Items
Nagare can "Pull" data from Zabbix to keep its inventory up to date.
- **Auto-Sync**: Nagare runs a background job (if configured) to sync every hour.
- **Manual Sync**: Click the **"Sync Now"** button on the Monitor or Group details page.

---

## 6. How Nagare Maps Data
| Nagare Field | Zabbix Mapping | Prometheus Mapping |
| :--- | :--- | :--- |
| `Host.Name` | `host` | `instance` |
| `Item.LastValue` | `lastvalue` | `metric_value` |
| `Alert.Severity` | `trigger.severity` | `labels.severity` |

---

## 🛠️ Debugging Integrations
- Visit the **Site Messages** section in the Nagare Dashboard. If a webhook fails, you'll see an error message like: *"Failed to parse alert from 192.168.1.50."*
- Check the **Audit Logs** for `POST /api/v1/alerts/webhook` to see incoming payloads.
