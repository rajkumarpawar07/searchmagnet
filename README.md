<div align="center">
  <h1>🧲 SearchMagnet</h1>
  <p><strong>A high-performance multimodal semantic search engine powered by Gemini.</strong></p>

  <!-- Badges -->
  <p>
    <img src="https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat-square&logo=go" alt="Go Version" />
    <img src="https://img.shields.io/badge/Next.js-15-black?style=flat-square&logo=next.js" alt="Next.js" />
    <img src="https://img.shields.io/badge/TailwindCSS-v4-38B2AC?style=flat-square&logo=tailwind-css" alt="Tailwind CSS" />
    <img src="https://img.shields.io/badge/Gemini-AI-4285F4?style=flat-square&logo=google" alt="Gemini AI" />
  </p>
</div>

<hr />

## 🌟 Overview

**SearchMagnet** allows you to search across multiple data modalities—images, videos, audio, and text—using natural language. By leveraging Google's cutting-edge Gemini embedding models, it maps your local media library into a unified semantic space, enabling instant and accurate cross-modal discovery.

Whether you prefer a robust **CLI tool** for automation or a sleek **Dashboard UI** for visual discovery, SearchMagnet has you covered.

## ✨ Features

- **🧠 True Multimodality:** Generate semantic embeddings for `.mp4`, `.png`, `.jpg`, `.wav`, `.pdf`, and raw text.
- **⚡ Blazing Fast API:** Built on **Go Fiber**, handling concurrent indexing and lightning-fast cosine similarity lookups.
- **🎨 Modern Web UI:** A beautiful, responsive Next.js 15 frontend featuring dark mode, glassmorphism, and live media previews.
- **🔌 Zero-Config Local Index:** Uses a lightweight JSON vector store (`embeddings.json`) for maximum portability without external database dependencies.
- **🛠️ Powerful CLI:** Index entire directories concurrently or perform quick semantic queries straight from your terminal.

## 🏗️ Architecture Stack

| Component | Technology | Description |
| :--- | :--- | :--- |
| **Backend API** | Go & Fiber | Concurrent worker pools, file processing, and API routing. |
| **Frontend UI** | Next.js 15 & React | Server components, Tailwind v4 styling, Lucide icons. |
| **AI / ML** | Google Gemini API | `gemini-embedding-2-preview` for high-dimensional vectors. |

---

## 🚀 Getting Started

### Prerequisites
Before you begin, ensure you have the following installed:
- [Go 1.20+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)

### 1. Clone & Configure
Clone the repository and set up your environment variables.

```bash
git clone https://github.com/rajkumarpawar07/searchmagnet.git
cd searchmagnet
```

Create a `.env` file in the root directory:
```env
# Required: Your Google Gemini API Key
GEMINI_API_KEY=your_gemini_api_key_here

# Optional: Override the default embedding model
GEMINI_EMBEDDING_MODEL=gemini-embedding-2-preview
```

---

## 💻 Usage

SearchMagnet provides two interfaces: a rich Dashboard UI and a CLI.

### Option A: The Web Dashboard (Recommended)

Run the backend and frontend simultaneously in separate terminals.

**Terminal 1: Start the Go API Server**
```bash
# Starts the server on http://localhost:3000
go run ./cmd/server
```

**Terminal 2: Start the Next.js Frontend**
```bash
cd web
npm install
npm run dev
```

Navigate to **[http://localhost:3001](http://localhost:3001)** in your browser to access the SearchMagnet dashboard. You can upload files, paste text, and search your local index visually.

---

### Option B: The CLI

For power users, SearchMagnet can be driven entirely from the command line.

**1. Index a Single File**
```bash
go run ./cmd/searchmagnet index ./path/to/video.mp4
```

**2. Index an Entire Directory (Concurrent)**
```bash
go run ./cmd/searchmagnet index-dir ./my-assets
```

**3. Perform a Semantic Search**
```bash
go run ./cmd/searchmagnet search "a dog playing in the grass"
```

**4. Find Similar Entries**
```bash
# Find entries similar to an existing ID
go run ./cmd/searchmagnet similar 1777392117840372900
```

---

## 🛣️ Roadmap

- [x] Initial CLI Application
- [x] Multimodal Embeddings (Text, Image, Video)
- [x] Go Fiber REST API
- [x] Next.js 15 Web Interface
- [ ] Migration from local JSON index to SQLite/ChromaDB for scale
- [ ] Release pre-compiled binaries via Goreleaser

---
<div align="center">
  <p>Built with ❤️ by <a href="https://github.com/rajkumarpawar07">Rajkumar Pawar</a></p>
</div>
