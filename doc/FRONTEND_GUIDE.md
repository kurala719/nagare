# Nagare Frontend & UX Engineering

Nagare's frontend is designed for high responsiveness, low initial latency, and modern remote-development workflows.

## 1. UX Engineering: Perceived Speed & Perception

Nagare employs several strategies to make a data-heavy monitoring dashboard feel instantaneous.

### 1.1 Dynamic Skeleton Screens (`el-skeleton`)
-   **Context**: The dashboard (`Dashboard.vue`) triggers 5+ asynchronous API calls upon entry.
-   **Implementation**: A custom-designed skeleton container replaces the entire dashboard content while `loading` is true and `lastUpdated` is empty.
-   **Benefit**: Users see the "shape" of the application (layout, charts, tables) before the data arrives, eliminating the jarring "jump" associated with late-loading content and reducing Cumulative Layout Shift (CLS).

## 2. Modern Build Optimization: Vite & Rollup

Nagare optimizes its production binary through a sophisticated **Manual Chunking Strategy**.

### 2.1 Chunk Partitioning Logic (`vite.config.js`)
To ensure efficient browser-side caching and faster initial loads, large dependencies are split:
-   **`vendor-element-plus`**: All Element Plus UI components.
-   **`vendor-echarts`**: The visualization and topology engine.
-   **`vendor-xterm`**: The WebSSH terminal logic.
-   **`vendor`**: All other third-party utilities.

### 2.2 Performance Metrics
-   **Initial JS Payload**: Reduced by ~35% on the main dashboard page.
-   **Caching**: Updates to the business logic (Vue files) do not invalidate the large vendor chunks, allowing for faster subsequent visits.

## 3. Network Connectivity: Dev Tunnel Interoperability

Nagare is natively compatible with **Microsoft Dev Tunnels**.

### 3.1 Anti-Phishing Bypass (`request.js` & `authFetch.js`)
-   **The Problem**: Tunnel providers inject a "Warning: Anti-Phishing" page that blocks non-interactive API calls (e.g., Axios from the browser or Webhooks).
-   **The Solution**: Nagare's HTTP utilities automatically inject the `X-Tunnel-Skip-AntiPhishing-Page: true` header into all requests.
-   **Outcome**: Transparent connectivity for external testers and webhook providers (Zabbix/Prometheus/OneBot).

## 4. Component Architecture
-   **State Management**: Composition API with dedicated refs for metrics, topology, and alerts.
-   **Visualization**: ECharts 5 with reactive `setOption` calls for real-time updates without full chart re-initialization.
-   **Terminal**: `xterm.js` with `FitAddon` for high-fidelity SSH emulation.
