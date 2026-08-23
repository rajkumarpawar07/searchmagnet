<div align="center">
  <h1>🧲 SearchMagnet</h1>
  <p><strong>Semantic search across your local files, powered by Gemini.</strong></p>

  <p>
    <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go Version" />
    <img src="https://img.shields.io/badge/Wails-v2-DF0000?style=flat-square&logo=wails" alt="Wails" />
    <img src="https://img.shields.io/badge/Next.js-15-black?style=flat-square&logo=next.js" alt="Next.js" />
    <img src="https://img.shields.io/badge/Gemini-AI-4285F4?style=flat-square&logo=google" alt="Gemini AI" />
  </p>
</div>

---

## 🌟 Overview

SearchMagnet lets you find images, videos, audio, PDFs, and text using **natural language** — no filenames required. Every file is embedded into a vector via Google's Gemini embedding models, and searches run as fast cosine-similarity lookups against a local index. Your files never leave your machine (only embedding requests go to Gemini).

Three interfaces are included, all sharing the same core engine:

| Interface | Best for |
| :--- | :--- |
| 🖥️ **Desktop App** (Wails) | Everyday use — native window, drag & drop, no browser |
| 🌐 **Web Dashboard** (Next.js) | Browser-based visual discovery |
| ⌨️ **CLI** | Automation and scripting |

## ✨ Features

- **🧠 True multimodality** — semantic embeddings for `.png` `.jpg` `.mp4` `.wav` `.pdf` and raw text
- **🖥️ Native desktop app** — drag & drop indexing, live progress, real media thumbnails, dark/light themes
- **⚡ Concurrent indexing** — worker pools embed multiple files in parallel
- **🔒 Local & private** — index lives in a portable JSON file; previews are served only for indexed files
- **🛠️ Scriptable CLI** — index files or entire directories, search, and find similar items from the terminal

## 🏗️ Architecture

```
searchmagnet/
├── internal/
│   ├── embedder/     # Gemini multimodal embeddings
│   ├── store/        # Vector store (JSON / ChromaDB) + cosine search
│   └── format/       # Terminal formatting helpers
├── cmd/
│   ├── server/       # REST API (Go Fiber)   → serves the web dashboard
│   ├── searchmagnet/ # CLI
│   └── migrate/      # JSON → ChromaDB migration tool
├── desktop/          # 🖥️ Wails desktop app (reuses internal/*)
│   └── frontend/     #    UI (vanilla HTML/CSS/JS, no build step)
└── web/              # 🌐 Next.js 15 dashboard
```

| Component | Technology |
| :--- | :--- |
| Core engine | Go — embeddings, vector store, cosine similarity |
| Desktop app | [Wails v2](https://wails.io) — Go backend + web UI in a native window |
| Web dashboard | Next.js 15, Tailwind CSS v4 |
| AI / ML | Google Gemini — `gemini-embedding-2-preview` |

---

## 🚀 Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/) — only for the web dashboard
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) — only for the desktop app:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

### 1. Clone

```bash
git clone https://github.com/rajkumarpawar07/searchmagnet.git
cd searchmagnet
```

### 2. Configure

Get a free API key at [aistudio.google.com/apikey](https://aistudio.google.com/apikey), then create a `.env` file in the repo root:

```env
GEMINI_API_KEY=your_gemini_api_key_here
GEMINI_EMBEDDING_MODEL=gemini-embedding-2-preview   # optional override
```

> The **desktop app** can also be configured in its own Settings page (stored in `%AppData%/SearchMagnet/config.json`) — the `.env` values are used as fallback.

---

## 💻 Usage

### 🖥️ Option A: Desktop App

```bash
cd desktop
wails dev      # development, with live reload
wails build    # production build → desktop/build/bin/SearchMagnet.exe
```

Then: **Upload** files (or drag & drop them onto the window) → **Search** in plain language.

Data locations on disk:

| What | Where |
| :--- | :--- |
| Settings | `%AppData%/SearchMagnet/config.json` |
| Vector index | `%AppData%/SearchMagnet/embeddings.json` |

### 🌐 Option B: Web Dashboard

Run the API server and the frontend in two terminals:

```bash
# Terminal 1 — API server on :3000
go run ./cmd/server
```

```bash
# Terminal 2 — dashboard on :3001
cd web
npm install
npm run dev
```

Open **http://localhost:3001**.

<details>
<summary><strong>REST API endpoints</strong></summary>

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/index` | Upload & index a file (multipart field `file`) |
| `POST` | `/api/index-text` | Index a text snippet |
| `GET` | `/api/search?q=&top=&type=` | Semantic search |
| `GET` | `/api/list` | List all indexed entries |
| `GET` | `/api/similar/:id` | Find entries similar to one ID |
| `DELETE` | `/api/entries/:id` | Delete an entry |
| `GET` | `/api/stats` | Index statistics |

</details>

### ⌨️ Option C: CLI

```bash
# Index a single file
go run ./cmd/searchmagnet index ./path/to/video.mp4

# Index an entire directory (concurrent)
go run ./cmd/searchmagnet index-dir ./my-assets

# Semantic search
go run ./cmd/searchmagnet search "a dog playing in the grass"

# Find entries similar to an existing ID
go run ./cmd/searchmagnet similar 1777392117840372900
```

---

## 🛣️ Roadmap

- [x] CLI application
- [x] Multimodal embeddings (text, image, video, audio, PDF)
- [x] REST API + Next.js dashboard
- [x] Native desktop app (Wails)
- [ ] Scale index beyond JSON (SQLite / ChromaDB)
- [ ] Pre-built binaries via GoReleaser

---

<div align="center">
  <p>Built with ❤️ by <a href="https://github.com/rajkumarpawar07">Rajkumar Pawar</a></p>
</div>
