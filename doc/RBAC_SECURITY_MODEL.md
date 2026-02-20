# Nagare Security Model: Roles & Privileges

Nagare uses a **Role-Based Access Control (RBAC)** system to ensure that only authorized personnel can make changes to critical infrastructure.

---

## 1. Privilege Levels

Nagare defines three distinct user roles, each building on the permissions of the level below it.

| Level | Role Name | Description | Key Capabilities |
| :--- | :--- | :--- | :--- |
| **1** | **User (Operator)** | Read-only access for monitoring. | View Dashboards, Consult AI, Run WebSSH (limited), Chat. |
| **2** | **Manager (DevOps)** | Operational control. | Add Hosts, Edit Alerts, Manage Knowledge Base, Run Playbooks. |
| **3** | **Admin (SRE Lead)** | Full system control. | Edit Config, Manage Users, Approve Registrations, View Audit Logs. |

---

## 2. Detailed Permission Matrix

### Intelligence & Monitoring
| Feature | User (L1) | Manager (L2) | Admin (L3) |
| :--- | :---: | :---: | :---: |
| View Dashboards | ✅ | ✅ | ✅ |
| Consult AI (RAG) | ✅ | ✅ | ✅ |
| Add/Edit Monitors | ❌ | ✅ | ✅ |
| Manage Knowledge Base | ❌ | ✅ | ✅ |

### Infrastructure & Automation
| Feature | User (L1) | Manager (L2) | Admin (L3) |
| :--- | :---: | :---: | :---: |
| View Hosts | ✅ | ✅ | ✅ |
| Connect WebSSH | ✅ | ✅ | ✅ |
| Run Playbooks | ❌ | ✅ | ✅ |
| Edit Playbooks | ❌ | ✅ | ✅ |
| Trigger Chaos | ❌ | ❌ | ✅ |

### Administration & Security
| Feature | User (L1) | Manager (L2) | Admin (L3) |
| :--- | :---: | :---: | :---: |
| Manage Users | ❌ | ❌ | ✅ |
| View Audit Logs | ❌ | ❌ | ✅ |
| Edit System Config | ❌ | ❌ | ✅ |
| Approve Registrations | ❌ | ❌ | ✅ |

---

## 3. WebSSH Security

WebSSH access is particularly sensitive. Even with Level 1 access, a user can only connect to hosts they have been explicitly granted access to via **Group Permissions** (feature roadmap).

- **Session Recording**: All WebSSH sessions are logged (metadata only, full keystroke logging is planned).
- **Timeouts**: Idle connections are automatically terminated after 15 minutes.

---

## 4. API Authentication

All API requests (except login/register) must include a **JWT Token** in the `Authorization` header:

`Authorization: Bearer <your_jwt_token>`

- **Token Expiry**: 24 hours.
- **Refresh Token**: Used to obtain a new access token without re-login.
