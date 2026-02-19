# Nagare Frontend & UX Engineering

This document outlines the performance, architecture, and user experience (UX) strategies used in the Nagare frontend.

## 1. Perceived Performance Strategy

### Dynamic Skeleton Screens (`el-skeleton`)
Nagare uses **Skeleton Screens** (e.g., in `Dashboard.vue`) to improve the "subjective perception" of load speed:
-   **Why**: When a user lands on the dashboard, several APIs (Health, Alerts, Hosts, Monitors, Providers) are called asynchronously.
-   **Logic**: Instead of a single spinning loader, Nagare renders a gray, animated layout that matches the final structure of the charts and tables. This reduces layout shift (CLS) and makes the UI feel "alive" even while data is still in transit.

## 2. Modern Build Optimization

### Vite 7 & Code Splitting (`vite.config.js`)
Nagare optimizes the production bundle through **Manual Chunks**:
-   **Split Logic**: Large libraries are separated into dedicated `.js` files:
    -   `element-plus`: Core UI components.
    -   `echarts`: All charting and visualization logic.
    -   `xterm`: The WebSSH terminal engine.
    -   `vendor`: Remaining third-party utilities.
-   **Benefit**: Users only download the libraries needed for a specific page. It also allows for efficient browser-side caching.

## 3. Network Connectivity Engineering

### Dev Tunnel Interoperability
Nagare is engineered for modern remote development environments (like Microsoft Dev Tunnels).
-   **The Problem**: Anti-phishing intercepts from tunnel providers often block API calls.
-   **The Solution**: 
    -   **Axios (`request.js`)**: All requests include the `X-Tunnel-Skip-AntiPhishing-Page: true` header.
    -   **Fetch (`authFetch.js`)**: The custom fetch utility also injects the same header.
-   **Benefit**: Devs can test webhooks and public endpoints (like `/api/v1/media/qq/message`) without manual bypasses.

## 4. UI Library & Styling
-   **Framework**: Vue 3 (Composition API).
-   **UI Kit**: Element Plus (Standardized for a clean, professional SRE look).
-   **CSS Strategy**: Vanilla CSS with a focus on CSS Variables for consistency and performance over Tailwind utility bloat.
-   **Data Vis**: ECharts 5 for real-time monitoring trends and topology.
