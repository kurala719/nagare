# Nagare - Unified Monorepo

Welcome to the **Nagare** project! This is a comprehensive monitoring and automation system featuring a modern web UI and a robust Go-based backend.

## 📂 Project Structure

This repository is organized as a monorepo containing both the frontend and backend codebases:

- **`frontend/`**: The web user interface built with **Vue 3** and **Vite**.
- **`backend/`**: The API server and core logic built with **Go (Golang)** and **Gin**.
- **`doc/`**: Project documentation, guides, and architectural references.
- **`tests/`**: Integration tests and utility scripts.

## 🚀 Getting Started

### Prerequisites

- **Go**: Version 1.21 or higher (for backend).
- **Node.js**: Version 18 or higher (for frontend).
- **Database**: A compatible SQL database (e.g., MySQL, SQLite) supported by GORM.
- **Redis**: (Optional) For task queue management.

### Quick Start

#### Backend

1.  Navigate to the backend directory:
    ```bash
    cd backend
    ```
2.  Install dependencies:
    ```bash
    go mod download
    ```
3.  Run the server:
    ```bash
    go run cmd/server/main.go
    ```
    *Note: Ensure your `configs/nagare_config.json` is properly configured.*

See [Backend README](backend/README.md) for detailed instructions.

#### Frontend

1.  Navigate to the frontend directory:
    ```bash
    cd frontend
    ```
2.  Install dependencies:
    ```bash
    npm install
    ```
3.  Start the development server:
    ```bash
    npm run dev
    ```

See [Frontend README](frontend/README.md) for detailed instructions.

## ✨ Key Features

- **Centralized Monitoring**: Unified management of Zabbix, Prometheus, and custom monitors.
- **Smart Alerting**: Multi-channel notifications including Email, Webhook, and QQ.
- **AI-Powered Analysis**: Intelligent alert diagnostics using Google Gemini LLM.
- **Dynamic Dashboards**: Real-time visualization of system health and metrics.

## 🛠️ Extensions

- **Interactive WebSSH**: Direct shell access to monitored hosts from the browser.
- **Automated Reporting**: Scheduled PDF reports with visual data analytics.
- **RAG Knowledge Base**: Context-aware alert analysis using local operation history.

## 📚 Documentation

Detailed documentation can be found in the `doc/` directory:

- [**Architecture Overview**](doc/ARCHITECTURE.md): High-level system design and API reference.
- [**Extension Features**](doc/EXTENSION_FEATURES.md): Deep dive into WebSSH, Reports, and RAG.
- [**Code Structure**](doc/CODE_STRUCTURE_REFERENCE.md): Data flow and file organization.
- [**Quick Start with QQ API**](doc/QUICK_START_QQ_API.md): Guide for integrating QQ bots.
- [**Implementation Summary**](doc/IMPLEMENTATION_SUMMARY.md): Status of current features.

## 🤝 Contributing

Please ensure you follow the project's coding standards and submit pull requests to the appropriate subdirectory (`frontend` or `backend`).

## 📄 License

[License Information]
