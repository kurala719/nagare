# Nagare AI & RAG Diagnostic Engine

This document provides a detailed technical breakdown of how Nagare transforms raw monitoring alerts into intelligent, context-aware diagnostic reports.

## 1. The Diagnostic Pipeline

When an alert is received, Nagare executes a multi-stage pipeline:

### Stage A: Tokenization & Entity Extraction
Nagare does not just pass raw text. It performs "Smarter Tokenization":
-   **FieldsFunc**: Uses a custom function to split text by common punctuation (`.,:;!?()[]`) while preserving IP addresses and service tags.
-   **Stop-word Filtering**: Removes high-frequency, low-signal words (`the`, `is`, `at`, `nagare`, `detected`) to focus on key error signatures.

### Stage B: The Retrieval Layer
Nagare queries the `KnowledgeBase` table (MySQL) using the extracted tokens.
-   **Initial Search**: Fetches the top 10 relevant entries from the database using a `LIKE %token%` query across `topic`, `content`, and `keywords`.

### Stage C: Keyword Scoring & Re-ranking
To minimize LLM hallucinations, Nagare applies a custom **Re-ranking Algorithm**:
1.  **Scoring**: For each retrieved entry, it calculates a `score`.
2.  **Boost Logic**: If an entry's content, topic, or keywords contain an exact match for one of the extracted tokens, its score increases by `+2`.
3.  **Final Ranking**: Entries are sorted by score in descending order.
4.  **Selection**: Only the top 3 most relevant results are kept to conserve LLM context window space.

### Stage D: LLM Prompt Construction
The results are injected into the system prompt as "Local Knowledge Reference Information". This allows Gemini (or OpenAI) to say: *"Based on your company's runbook for 'DB-Master-Failover'..."* rather than giving generic internet advice.

## 2. Model Support
Nagare supports a hybrid model approach via the `internal/service/chat.go` layer:
-   **Google Gemini (Primary)**: Using the `gemini-2.0-flash` model for high-speed analysis.
-   **OpenAI/Ollama**: Standardized via the `llm.Client` abstraction for local or cloud deployments.

## 3. "Roast Mode" (Persona Logic)
For interactive chat, Nagare supports a "Roast" mode:
-   **Logic**: A specialized system prompt that instructs the AI to be a "witty, slightly sarcastic senior SRE".
-   **Goal**: To make monitoring more engaging while still providing concrete, actionable fixes.
