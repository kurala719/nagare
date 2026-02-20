# API Group: Communication & Notifications

This group manages how Nagare communicates with users, including browser notifications, chat bots (QQ), and formal alert channels (Email/Webhooks).

---

## 🔔 1. Site Messages (Browser Notifications)

Internal notifications displayed in the Nagare dashboard.

### **GET** `/api/v1/site-messages`
Lists internal messages for the current user.
- **Parameters**: `unread_only` (Boolean), `limit`, `offset`.

### **GET** `/api/v1/site-messages/unread-count`
Returns the count of unread messages. Often used for the "Red Badge" in the UI.

### **PUT** `/api/v1/site-messages/:id/read`
Marks a specific message as read.

### **PUT** `/api/v1/site-messages/read-all`
Marks all messages for the current user as read.

### **GET** `/api/v1/site-messages/ws`
**WebSocket Endpoint**. Used to push messages to the browser in real-time.

---

## 🤖 2. QQ & IM Bot Integration

Integration with the OneBot 11 protocol (e.g., NapCat/Gocqhttp).

### **GET** `/api/v1/media/qq/ws`
**Reverse WebSocket Endpoint** for OneBot 11 (e.g., NapCat).
- **Protocol**: OneBot 11 Reverse WebSocket.
- **Logic**: 
  1. Receives a connection from the QQ Bot (NapCat).
  2. Handles incoming events (messages) and outgoing actions (API calls).
  3. Provides faster, two-way communication and is the preferred method for QQ integration.

### **POST** `/api/v1/media/qq/message`
**HTTP Webhook Ingest Point** for QQ messages (Legacy/Alternative).
- **Logic**: 
  1. Receives a message from the QQ Bot via HTTP POST.
  2. Checks if the sender (User or Group) is in the **QQ Whitelist**.
  3. If authorized and message starts with `/`, it executes as a command.
  4. Sends a reply back to QQ via HTTP API.

### **POST** `/api/v1/im/command`
Directly triggers an IM command (used for testing).

---

## 🛡️ 3. QQ Whitelist Management

Controls who can talk to the Nagare Bot on QQ.

### **GET** `/api/v1/qq-whitelist`
Lists authorized QQ IDs and groups.

### **POST** `/api/v1/qq-whitelist`
Adds a new ID to the whitelist.
- **Body**: `{ "qq_identifier": "123456", "type": 0, "can_command": 1, "can_receive": 1 }`

---

## 📧 4. Media & Media Types (Alert Channels)

Defines *how* and *where* alerts are sent.

### **GET** `/api/v1/media-types`
Lists supported notification protocols (e.g., `email`, `webhook`, `qq`).

### **POST** `/api/v1/media`
Creates a specific destination (e.g., "SRE Team Email").
- **Body**: `{ "name": "Team Email", "type": "email", "target": "sre@company.com" }`
