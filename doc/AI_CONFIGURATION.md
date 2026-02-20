# Nagare AI Configuration: Setting Up the Brain

Nagare relies on an external AI provider (LLM) to perform analysis, RAG retrieval, and playbook generation. This guide explains how to configure these providers.

---

## 1. Supported Providers

Nagare currently supports the following AI backends:

| Type ID | Provider | Description |
| :--- | :--- | :--- |
| **1** | **Google Gemini** | (Recommended) Best balance of speed and reasoning. Uses `gemini-1.5-pro` or `gemini-2.0-flash`. |
| **2** | **OpenAI** | Standard GPT-4o or GPT-3.5-Turbo models. |

---

## 2. Configuration via API

You can manage providers via the **Settings -> AI Providers** page in the frontend, or use the API directly.

### Adding a Gemini Provider (Recommended)
```json
POST /api/v1/providers
{
  "name": "Gemini Pro",
  "type": 1,
  "api_key": "AIzaSy...",
  "default_model": "gemini-1.5-pro-latest",
  "enabled": 1
}
```

### Adding an OpenAI Provider
```json
POST /api/v1/providers
{
  "name": "OpenAI GPT-4",
  "type": 2,
  "api_key": "sk-...",
  "default_model": "gpt-4o-mini",
  "enabled": 1
}
```

---

## 3. The RAG Engine Configuration

The RAG (Retrieval-Augmented Generation) engine is enabled by default. It enhances AI answers by looking up your local **Knowledge Base**.

### How to Tune RAG
Currently, RAG tuning is done via the `configs/nagare_config.json` file:

```json
"ai": {
  "analysis_enabled": true,
  "analysis_timeout_seconds": 30,
  "analysis_min_severity": 2
}
```

- **`analysis_enabled`**: Set to `false` to disable automatic AI analysis of new alerts (saves tokens).
- **`analysis_min_severity`**: Only analyze alerts with severity >= 2 (Warning/Critical).
- **`analysis_timeout_seconds`**: Increase this if you are using a slow local model (like Llama 2 70b on CPU).

---

## 4. Prompt Customization
Nagare uses "System Prompts" to define the AI's persona. These are currently hardcoded in the Go backend (`internal/service/ai.go`) for stability, but they adapt based on the user's locale (English/Chinese).

### Persona Modes
When chatting, you can select a "Mode":
- **Standard**: Professional, concise.
- **Roast**: Sarcastic, critical (good for spotting bad configurations). 
  - *Trigger*: Set `"mode": "roast"` in the chat API.

---

## 5. Privacy & Data Safety
- **Sanitization**: Before sending data to the AI, Nagare attempts to mask sensitive patterns (like passwords) using regex.
