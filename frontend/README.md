# Nagare Frontend

The web user interface for the Nagare monitoring platform, built with **Vue 3** and **Vite**.

## 🛠️ Tech Stack

- **Framework**: [Vue 3](https://vuejs.org/)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **UI Library**: [Element Plus](https://element-plus.org/)
- **Charting**: [ECharts](https://echarts.apache.org/)
- **HTTP Client**: [Axios](https://axios-http.com/)
- **Routing**: [Vue Router](https://router.vuejs.org/)
- **Internationalization**: [Vue I18n](https://vue-i18n.intlify.dev/)

## 🚀 Getting Started

### Prerequisites

- **Node.js**: Version 20.19.0 or higher.
- **npm**: Included with Node.js.

### Installation

1.  Navigate to the frontend directory:
    ```bash
    cd frontend
    ```
2.  Install dependencies:
    ```bash
    npm install
    ```

### Development

Start the development server with hot-reload:

```bash
npm run dev
```

The application will be available at `http://localhost:5173` (default Vite port).

**Note on API Proxy:**
By default, the development server proxies requests starting with `/api` to `http://localhost:8080`. Ensure your backend server is running on this port or update `vite.config.js` accordingly.

### Production Build

Build the application for production:

```bash
npm run build
```

The output will be in the `dist/` directory.

### Preview Production Build

Preview the built application locally:

```bash
npm run preview
```

## 📂 Directory Structure

- **`src/`**: Source code.
    - **`api/`**: API client modules for backend communication.
    - **`assets/`**: Static assets (CSS, images).
    - **`components/`**: Reusable Vue components.
    - **`router/`**: Route definitions.
    - **`views/`**: Page-level components.
    - **`utils/`**: Helper utilities.
    - **`i18n/`**: Localization files.
- **`public/`**: Static files served as-is (e.g., `favicon.ico`).
- **`vite.config.js`**: Vite configuration file.

## 🤝 Contribution

Please follow the [Vue.js Style Guide](https://vuejs.org/style-guide/) when contributing.