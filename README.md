# Nagare (流) - Next-Gen AI-Powered Operations Platform

Nagare is a unified monitoring, automation, and diagnostic platform designed for modern SRE and DevOps teams. It combines high-concurrency data ingestion with Large Language Models (LLM) and Retrieval-Augmented Generation (RAG) to transform raw monitoring data into actionable intelligence.

## 🌟 Key Features

- **Unified Monitoring**: Seamless integration with Zabbix and Prometheus.
- **AI Diagnostic Engine (RAG)**: Context-aware alert analysis using local operations history and Gemini/OpenAI.
- **High Performance**: Go-based backend optimized with `jsoniter` for 20-30% faster serialization.
- **Integrated WebSSH**: Secure, browser-based terminal access to monitored hosts.
- **Automated Reporting**: Professional PDF reports with AI-generated summaries and server-side charts.
- **Modern UI**: Vue 3 + Vite frontend with skeleton screens for optimized perceived speed.
- **MCP Protocol**: Support for Model Context Protocol, enabling Nagare to act as an AI Agent toolkit.

## 📂 Project Structure

- **`frontend/`**: Vue 3, Vite, Element Plus, ECharts, xterm.js.
- **`backend/`**: Go 1.24, Gin, GORM, Redis, jsoniter, Gemini SDK.
- **`doc/`**: Detailed architectural and feature documentation.

## 🚀 Quick Start

### Backend (Go)
1. `cd backend`
2. `go mod download`
3. `go build -tags=jsoniter -o nagare-server ./cmd/server`
4. `./nagare-server` (Ensure `configs/nagare_config.json` is set up)

### Frontend (Vue 3)
1. `cd frontend`
2. `npm install`
3. `npm run dev`

## 🛠️ Performance & Security Highlights
- **Build Tags**: Uses `-tags=jsoniter` for high-speed JSON processing.
- **Code Splitting**: Optimized Vite build with manual chunks for UI and Charting libraries.
- **Tunnel Friendly**: Built-in support for Microsoft Dev Tunnels with anti-phishing bypass.
- **RAG Optimization**: Custom keyword re-ranking algorithm for superior AI context retrieval.

## 📄 License
Apache License 2.0
