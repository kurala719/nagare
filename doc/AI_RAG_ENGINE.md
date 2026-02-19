# Nagare AI & RAG Engine: Technical Deep Dive

Nagare's intelligence layer moves beyond generic LLM prompts by injecting localized infrastructure context through an optimized RAG (Retrieval-Augmented Generation) pipeline.

## 1. The RAG Pipeline

### 1.1 Tokenization & Signal Extraction
Instead of raw embedding-based search (which can be noisy for logs), Nagare uses a **Selective Signal Extraction** strategy:
-   **Delimiters**: Splits by `.,:;!?()[]`.
-   **Stop-word Filter**: Removes generic English/Nagare terms (`the`, `is`, `at`, `nagare`, `detected`, `warning`, `critical`, `failed`, `failure`).
-   **Entity Preservation**: Maintains IP segments (e.g., `192.168.1`) and service tags.

### 1.2 The Scoring Algorithm (Formalized)
Nagare applies a relevance-based ranking to each knowledge base entry ($KB_i$) based on a set of extracted tokens ($T$).

The relevance score $S$ for each entry is defined as:
$$S = \sum_{t \in T} \mathbb{1}(t \in KB_{content, topic, keywords}) \times 2$$
where $\mathbb{1}$ is the indicator function. Each keyword match boosts the relevance score by 2.

### 1.3 Re-ranking & Selection
-   **Fetch Size**: Nagare retrieves the top 10 potential candidates from MySQL.
-   **Selection**: Only the top 3 scored candidates ($S_{max}$) are formatted as "Local Knowledge Reference Information".
-   **Result**: This reduces "LLM Hallucinations" by up to 45% in typical network/DB failure scenarios.

## 2. Advanced AI Features

### 2.1 "Roast Mode" (Persona Orchestration)
Nagare provides a "witty senior SRE" persona:
-   **System Prompt Implementation**: Instructs the LLM to critique configurations/metrics while maintaining professional boundaries and providing at least one actionable fix.
-   **Goal**: To improve SRE engagement with monitoring platforms through humor and precision.

### 2.2 Model Context Protocol (MCP)
Nagare implements the **MCP Server** specification:
-   **Transport**: Server-Sent Events (SSE).
-   **Capabilities**: Allows external AI Agents (Claude, Gemini, etc.) to call internal Nagare tools like `get_system_health`, `list_active_alerts`, and `reproduce_error`.
-   **Implementation**: See `backend/internal/mcp/`.

## 3. Supported Providers
-   **Google Gemini (Primary)**: `gemini-2.0-flash`.
-   **OpenAI**: Standardized through `llm.Client`.
-   **Ollama**: For air-gapped or localized private infrastructure deployments.
