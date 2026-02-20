# Nagare Communication: How the Brain Talks to You

Nagare doesn't just sit there quietly; it has several ways to reach out and tell you what's happening. We call this the "Communication Layer."

## 1. Site Messages: The Real-Time Pop-up
When you are logged into the Nagare dashboard, you are connected via a **WebSocket** (a permanent, high-speed phone line).
- **Instant Alerts**: If a server fails, a message pops up immediately in your browser.
- **Unread Count**: Just like a smartphone, Nagare shows a red badge with the number of messages waiting for you.
- **Read All**: You can clear all your notifications with one click once you've checked them.

## 2. IM & QQ Integration: Alerts in Your Pocket
Nagare can talk to external chat apps, specifically **QQ** (a popular messaging platform).
- **The QQ Bot**: Nagare can send alert messages directly to a QQ group or a private message.
- **IM Commands**: You can actually "talk back" to Nagare. By typing specific commands in QQ, you can ask for the status of a server or acknowledge an alert.
- **Whitelist Security**: To prevent random people from messaging your bot, Nagare has a **Whitelist**. Only QQ IDs on this list are allowed to interact with the system.

## 3. Email & Webhooks
- **Standard Email**: For formal records of critical failures.
- **Outbound Webhooks**: Nagare can send its own "SOS" to other systems, like Slack, Teams, or your company's internal tools.

## 4. Privilege Controls
Communication is sensitive. 
- **Users (Level 1)** can receive messages and chat.
- **Managers (Level 2)** can manage the Whitelist and configure which alerts go to which channels.

## 5. Summary
Whether you are at your desk looking at the dashboard, or on the go with your phone, Nagare ensures that you are always the first to know when your infrastructure needs attention.
