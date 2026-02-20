# Nagare Backend

The backend server for the Nagare monitoring platform, built with **Go (Golang)** and **Gin**.

## 🛠️ Tech Stack

- **Language**: [Go](https://go.dev/) (v1.24+)
- **Web Framework**: [Gin](https://gin-gonic.com/)
- **ORM**: [GORM](https://gorm.io/)
- **Database**: MySQL (default), SQLite, Postgres (supported by GORM)
- **Configuration**: [Viper](https://github.com/spf13/viper)
- **AI Integration**: Google GenAI SDK

## 🚀 Getting Started

### Prerequisites

- **Go**: Version 1.24 or higher.
- **Database**: A running SQL database instance (e.g., MySQL).

### Configuration

The application uses a JSON configuration file located at `configs/nagare_config.json`.

**Example Configuration:**
```json
{
  "server": {
    "port": 8080,
    "mode": "debug"
  },
  "database": {
    "dsn": "user:password@tcp(127.0.0.1:3306)/nagare?charset=utf8mb4&parseTime=True&loc=Local"
  }
}
```

Copy the example config (if available) or create your own based on the structure above.

### Running Locally

1.  Navigate to the backend directory:
    ```bash
    cd backend
    ```
2.  Install dependencies:
    ```bash
    go mod download
    ```
3.  Run the server using the Makefile:
    ```bash
    make run
    ```
    Or directly with Go:
    ```bash
    go run cmd/server/main.go
    ```

### Building

To build the binary:

```bash
make build
```
The binary will be output to `bin/nagare-web-server`.

## 📂 Directory Structure

- **`cmd/server/`**: Application entry point (`main.go`) and router setup.
- **`internal/`**: Private application code.
    - **`api/`**: HTTP handlers (Controllers).
    - **`service/`**: Business logic.
    - **`repository/`**: Data access layer.
    - **`model/`**: Data entities and structs.
    - **`mcp/`**: Model Context Protocol server implementation.
- **`pkg/`**: Public libraries.
- **`configs/`**: Configuration files.

## 🧪 Testing

Run unit tests:

```bash
make test
```

## 🤝 Contribution

Please ensure your code is formatted and passes `go vet` before submitting a PR.
```bash
make fmt
make vet
```
