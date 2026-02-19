# Nagare AI & RAG Engine Technical Specification

Nagare's intelligence layer minimizes LLM hallucinations by anchoring analysis in localized infrastructure runbooks.

## 1. Retrieval-Augmented Generation (RAG) Pipeline

### 1.1 Pre-processing: Signal Extraction
Nagare extracts high-signal tokens ($T$) from alert messages:
- **Tokenizer**: Custom `FieldsFunc` splitting on `.,:;!?()[]`.
- **Filtering**: Removal of stop-words and Nagare-specific boilerplate.
- **Entity Preservation**: Logic specifically maintains IP address patterns and service tags.

### 1.2 Retrieval & Scoring Algorithm
For an alert's token set $T$, Nagare queries the knowledge base. Each candidate entry ($KB_i$) is assigned a relevance score $S_i$:

$$S_i = \sum_{t \in T} \mathbb{1}(t \in KB_{content} \cup KB_{topic} \cup KB_{keywords}) \times 2$$

- **Initial Fetch**: Nagare retrieves top 10 candidates via SQL `LIKE` expansion.
- **Re-ranking**: Local Go code re-calculates $S_i$ for precision.
- **Final Context**: The top 3 candidates where $S_i > 0$ are injected into the LLM system prompt.

## 2. Model Orchestration

### 2.1 Multi-Provider Support
Nagare abstracts LLM communication through the `llm.Client` interface:
- **Google Gemini**: Primary high-speed provider (`gemini-2.0-flash`).
- **OpenAI/Ollama**: Supported for enterprise cloud or air-gapped local deployments.

### 2.2 Persona Configuration ("Roast Mode")
Implemented as a dynamic system prompt prefix:
- **Constraints**: Instructs the AI to be sarcastic/witty while requiring at least one concrete, technically valid remediation step.

## 3. Model Context Protocol (MCP)
Nagare acts as an **MCP Server**, allowing external agents (like Claude Desktop) to connect via SSE:
- **Toolbox**: Exposes internal Nagare methods (`get_metrics`, `list_alerts`) as executable tools for autonomous AI agents.
