# Nagare Automated Reporting & Task Orchestration

Nagare automates the generation of professional IT infrastructure reports.

## 1. Report Generation Workflow

### Aggregation Layer (`internal/service/report.go`)
-   **Data Points**: Total Alerts, Average Uptime, Resource Consumers (CPU/Memory), Alert Trends.
-   **Host Analytics**: Identifies "Top 5 Most Frequent Failures" and "Critical Host Issues".
-   **AI Executive Summary**: Nagare calls the LLM (Gemini/OpenAI) to generate a concise summary of the data, providing a human-readable interpretation of the weekly/monthly health.

### The Rendering Engine
Nagare uses a server-side Go-native approach for PDF creation:
1.  **Go-Chart Engine (`service/utils/charts.go`)**: Renders high-quality PNG charts (Pie, Line, Bar) from raw monitoring data.
2.  **Maroto Library**: A layout engine that builds the PDF page-by-page.
3.  **Layout Logic**:
    -   **Page 1**: Professional Header, AI Executive Summary, Health & Trend Charts.
    -   **Page 2**: Critical Host Analytics, Resource Consumption Tables, Stability Issue Frequency.

### Non-Blocking Processing
-   **Async Generation**: Generation is triggered as a background Goroutine to avoid blocking the user's web session.
-   **Storage**: PDFs are saved in the `public/reports/` directory with unique, secure IDs.

## 2. Task Orchestration & Scheduling

### Cron Scheduler (`internal/service/cron.go`)
Nagare uses the `robfig/cron/v3` library to automate system tasks:
-   **Weekly Reports**: Triggered every Monday at 00:00.
-   **Monthly Reports**: Triggered on the 1st of every month at 00:00.
-   **Auto-Sync**: Background syncing of monitor (Zabbix/Prometheus) data.

### Health Feedback
Nagare monitors its own reporting system:
-   **Logging**: Every Cron Job execution is logged with its duration and outcome.
-   **Site Messages**: Users receive a real-time notification via the WebSocket hub when a new report is ready for download.
