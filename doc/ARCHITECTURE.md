# Nagare System Architecture: The Digital Nervous System

Nagare is built as a **Monorepo**—a single home for both the Brain (Backend) and the Body (Frontend). This ensures they always work in perfect sync.

## 1. The Backend (The Brain)
Built with **Go (Golang)**, chosen for its incredible speed and efficiency.
- **The Messenger (Gin)**: A high-speed web framework that handles thousands of messages per second.
- **The Fast-Talker (jsoniter)**: Nagare uses a specialized library to read data 30% faster than normal programs.
- **The Notebook (MySQL & GORM)**: Stores all server history, user notes, and AI logs.
- **The Waiting Room (Redis)**: A queue that holds heavy tasks (like making a big PDF report) so the user doesn't have to wait.

## 2. The Frontend (The Interface)
Built with **Vue 3**, designed to be beautiful and responsive.
- **Instant Feedback**: Uses "Skeleton Screens"—gray placeholders that appear while data is loading—so the app never feels "stuck."
- **Live Updates**: Uses WebSockets (a permanent phone line between your browser and the server) to show alerts the second they happen.

## 4. Deep Dive: The Concurrency Engine
Nagare is designed to handle thousands of events without slowing down. It achieves this through three key components:

### A. The Global Hub (WebSocket Orchestrator)
The **Hub** (found in `internal/service/hub.go`) acts as a central traffic controller. 
- It maintains a registry of all active browser connections.
- When an alert is detected, the Hub "broadcasts" the message only to the relevant users.
- This prevents the server from wasting energy sending data to people who aren't even looking at the dashboard.

### B. The Worker Pool (Async Task Queue)
For heavy tasks like generating a 50-page PDF report or running a complex Ansible playbook, Nagare uses a **Worker Pool**:
- **Task Dispatcher**: Puts the task into a "Waiting Room" (Redis).
- **Workers**: Multiple background "laborers" pull tasks from the queue as soon as they are free.
- **Independence**: If a worker crashes while making a report, the rest of the Nagare "Brain" keeps running perfectly.

### C. The Redis Backplane
Redis is more than just a cache for Nagare; it's the **Shared Memory**:
- **Inter-process Communication**: If you run multiple Nagare instances for high availability, Redis ensures they all know what the others are doing.
- **Event Buffering**: If the database is busy, Redis holds the incoming alerts in a high-speed buffer so no "SOS" signal is ever lost.

## 5. Technology Stack Summary
- **Language**: Go 1.24 (The speed of C with the safety of Python).
- **Web Framework**: Gin (The fastest web engine for Go).
- **ORM**: GORM (Sophisticated database mapping).
- **Real-time**: Gorilla WebSockets.
- **Frontend**: Vue 3 + Vite + Element Plus.
- **AI**: Google GenAI SDK (Native integration with Gemini).
