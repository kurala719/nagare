# Nagare System Architecture

Nagare is designed as an intelligent "Brain" overlay for traditional monitoring tools (like Zabbix and Prometheus). It follows a **decoupled, layered architecture** to ensure maintainability, scalability, and clean separation of concerns.

---

## 🏗️ System Layers

Nagare is organized into distinct logical layers, each with a specific responsibility:

### 1. Presentation Layer (Frontend)
- **Technology**: Vue 3 (Composition API), Vite, Element Plus.
- **Responsibility**: User interface, real-time data visualization (ECharts), WebSSH terminal interaction (xterm.js), and AI chat interface.
- **Communication**: Communicates with the Backend via RESTful APIs (using Axios) and WebSockets for real-time updates.

### 2. API / Controller Layer (Backend API)
- **Technology**: Gin Gonic (HTTP Framework).
- **Location**: `backend/internal/api`
- **Responsibility**: Request parsing, input validation, authentication/authorization middleware (JWT), and routing requests to the appropriate Service functions.

### 3. Business Logic Layer (Service)
- **Technology**: Pure Go.
- **Location**: `backend/internal/service`
- **Responsibility**: The "Heart" of the system. Contains all core business logic, including:
    - **Intelligence Engine**: RAG (Retrieval-Augmented Generation) orchestration, AI provider integration (Gemini/OpenAI).
    - **Trigger Engine**: Logic for evaluating item thresholds and alert conditions.
    - **Automation**: Ansible playbook execution and WebSSH session management.
    - **Status Management**: Cascading health score calculations and system state synchronization.

### 4. Data Access Layer (Repository/DAO)
- **Technology**: GORM (Object-Relational Mapping).
- **Location**: `backend/internal/repository`
- **Responsibility**: Abstracting database operations. Ensures that the Service layer does not need to know the specifics of SQL queries or schema structure.

### 5. Infrastructure Layer
- **Technology**: MySQL/MariaDB, Redis (optional), External APIs (Zabbix, Google Gmail, NapCat OneBot).
- **Responsibility**: Persistent storage, message queuing, and external system integrations.

---

## 🔄 Core Data Flows

### The Trigger-Alert-Action Pipeline
Nagare automates incident response through a high-performance three-stage pipeline:

1.  **Ingestion & Detection** (`Items`): System pulls or receives metrics from external sources.
2.  **Evaluation** (`Triggers`): The Trigger Engine evaluates these metrics against defined thresholds. If a condition is met, an **Internal Alert** is generated.
3.  **Execution** (`Actions` & `Media`): Alerts are filtered by rule-based actions, which then dispatch notifications via various Media channels (Gmail, Webhooks, QQ).

### AI-Assisted Diagnostics (RAG Flow)
1.  An alert is detected.
2.  The `ai.go` service fetches context from the `KnowledgeBase` (RAG) and historical logs.
3.  The system queries an LLM (e.g., Gemini) for root cause analysis.
4.  The diagnostic result is attached to the alert as a `Comment` for the administrator.

---

## 📂 Directory Structure Highlights

```text
backend/
├── cmd/server/         # Application entry point & router initialization
├── internal/
│   ├── api/            # Controller Layer (Gin Handlers)
│   ├── service/        # Business Logic Layer (Core Engine)
│   ├── repository/     # Data Access Layer (GORM/DAO)
│   ├── model/          # Data Models & Entities
│   ├── database/       # DB Connection management
│   ├── migration/      # Schema initialization & updates
│   └── mcp/            # Model Context Protocol client implementation
└── pkg/                # Shared utilities (Queues, Crypto, etc.)

frontend/
├── src/
│   ├── api/            # Backend API client definitions
│   ├── views/          # Page components (Dashboard, Alerts, etc.)
│   ├── components/     # Reusable UI components
│   ├── layout/         # Standard application frame
│   └── utils/          # Frontend helpers (Auth, Request interceptors)
```

---

## 🛡️ Security & Compliance
- **Authentication**: Stateless JWT-based authentication.
- **Authorization**: Role-Based Access Control (RBAC) with three levels: User, Admin, SuperAdmin.
- **Audit**: Every destructive or sensitive action is recorded in the `audit_logs` table via middleware.
- **Data Integrity**: Automated retention policies ensure that historical data is pruned periodically to maintain performance.

For detailed information on specific entities, refer to the [Database Schema Documentation](DATABASE_SCHEMA.md).
