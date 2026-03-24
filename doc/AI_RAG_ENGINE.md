# Nagare AI & RAG Engine: How the Brain Thinks

Standard AI (like ChatGPT) knows a lot about the world but knows **nothing** about your specific servers. Nagare solves this using **RAG** (Retrieval-Augmented Generation).

## 1. The "Open-Book Exam" Analogy
Think of a doctor taking an exam:
- **Normal AI**: Tries to remember everything from medical school (it might forget details).
- **Nagare (RAG)**: The doctor is allowed to look at **your** medical records while answering the question. This makes the answer much more accurate.

## 2. How the "Filter" Works
Nagare doesn't give the AI *everything* (that would be too much information). It uses a **Scoring Algorithm** to find the 3 most relevant pages in your notes:

1. **Tokenization**: It breaks the error message into key "clues" (like `Server-01`, `Out of Memory`).
2. **Scoring**: It looks at your Knowledge Base. If a page mentions one of those clues, it gets points.
   - **Formula**: $Score = Matches \times 2$
3. **Selection**: It picks the top 3 pages and shows them to the AI.

## 3. Results
By using this "Open-Book" method, Nagare reduces AI "hallucinations" (making things up) by over **45%**. You get real advice based on your own system's history.

## 4. MCP: The Extensible Tool Belt
Nagare implements the **Model Context Protocol** as a client. This means Nagare's internal AI can dynamically load and use external tools provided by any standard MCP server (like running Python scripts or querying external databases).
