# Nagare Extension Features

Nagare extends traditional monitoring with powerful modules for automation and intelligence.

## 1. AI Diagnostic Engine (RAG)

### The Problem
Traditional LLMs often "hallucinate" when diagnosing specific infrastructure alerts because they lack local context (e.g., custom IP segments, specific service runbooks).

### The Solution: Optimized RAG
Nagare implements a three-step **Retrieve-Augmented Generation** pipeline:
1.  **Tokenization & Entity Extraction**: Captures key entities like IP segments, service names, and error codes.
2.  **Keyword-Based Re-ranking**: Queries the `KnowledgeBase` and applies a custom scoring algorithm:
    -   `Score += 2` for direct keyword matches.
    -   Filters out "stop-words" to focus on high-signal terms.
3.  **Context-Aware Prompting**: The top 3 most relevant results are injected into the Gemini/OpenAI prompt as "Local Knowledge Reference Information".

## 2. Integrated WebSSH

### Technology
- **Backend**: `golang.org/x/crypto/ssh` for secure node communication.
- **Frontend**: `xterm.js` for the terminal UI.
- **Protocol**: WebSocket (Binary/JSON) for data and control signals (resize).

### Security Measures
- **DOM Purification**: No use of `innerHTML` for PTY data to prevent XSS.
- **WebSocket Origin Check**: Strict origin validation to prevent CSRF.
- **wss Support**: Automatically detects and uses secure WebSocket protocol.

## 3. Automated Reporting

### Generation Workflow
1.  **Aggregation**: Collects weekly/monthly metrics (Alert trends, Host status).
2.  **Server-Side Charting**: Generates PNG charts (Pie, Line, Bar) using a Go-native chart engine.
3.  **PDF Compilation**: Uses the `Maroto` library to build professional documents.
4.  **Asynchronous Delivery**: Offloaded to Redis workers to ensure the main UI remains responsive.

## 4. MCP (Model Context Protocol)

Nagare can act as a **tool provider** for other AI Agents.
-   **SSE Handler**: Provides an event stream for real-time tool execution.
-   **Message Handler**: Allows LLMs to "query" the Nagare database or "trigger" actions via standardized JSON messages.
