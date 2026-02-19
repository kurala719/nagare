# Nagare WebSSH Technical Implementation

Nagare provides a secure, high-performance terminal experience through the web browser.

## 1. The Proxy Architecture

The WebSSH feature acts as a bridge between the browser's WebSocket and the host's SSH server.

### Backend Bridge (`backend/internal/api/webssh.go`)
-   **Upgrader**: Uses `gorilla/websocket` with a strict `CheckOrigin` policy (customizable for production).
-   **SSH Dialing**: Uses `golang.org/x/crypto/ssh` to connect to hosts via IP/Port.
-   **PTY Handling**: Requests a `xterm-256color` pseudo-terminal. Default size is `80x24`, but this is dynamically updated via the `resize` JSON signal.
-   **Piping Logic**:
    -   `io.Copy` from SSH Stdout/Stderr to WebSocket (Binary Message).
    -   WebSocket Text Message (JSON) -> Parse `type` (data/resize) -> Write to SSH Stdin.

### Frontend Rendering (`frontend/src/views/Terminal.vue`)
-   **Engine**: `xterm.js` for full VT100/Xterm terminal support.
-   **FitAddon**: Ensures the terminal grid perfectly matches its container size on the page.
-   **Resize Support**: Captures browser window resize events and sends a `{"type":"resize", "cols":X, "rows":Y}` signal to the backend.

## 2. Security Engineering

### Cross-Site Scripting (XSS) Prevention
Nagare is immune to common terminal XSS attacks:
-   **DOM Purification**: By using `xterm.js`, data from the remote PTY is never injected directly into the DOM (no `innerHTML`). Instead, it is rendered into a canvas-based or managed-DOM terminal grid.

### WebSocket Safety
-   **wss:// Protocol**: Automatically detects and upgrades the connection to `wss://` on production environments.
-   **Authorization**: Every WebSocket connection requires a valid JWT token passed via a query parameter.
-   **Origin Check**: The backend `CheckOrigin` ensures the connection is only allowed from recognized frontends.

## 3. Ad-Hoc Connections
Nagare supports both "Saved Host" and "Direct Connect":
-   **Saved Host**: Connection details (IP, User, Password) are securely fetched from the `hosts` table.
-   **Direct Connect**: Users can input temporary credentials (IP, Port, User, Password) for one-off troubleshooting.
