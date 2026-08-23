const API_BASE = "http://localhost:3000/api";

export type SearchResult = {
  rank: number;
  score: number;
  id: string;
  label: string;
  file_path: string;
  content_type: string;
  mime_type: string;
  indexed_at: string;
};

export type IndexEntry = {
  id: string;
  label: string;
  file_path: string;
  content_type: string;
  mime_type: string;
  indexed_at: string;
};

export type Stats = {
  total_entries: number;
  entries_by_type: Record<string, number>;
  supported_extensions: string[];
  max_file_size: string;
  embedding_model: string;
};

export type IndexResponse = {
  message: string;
  entry: IndexEntry;
  dim: number;
};

export const api = {
  async getStats(): Promise<Stats> {
    const res = await fetch(`${API_BASE}/stats`);
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },

  async search(q: string, top: number = 5, type: string = ""): Promise<{ results: SearchResult[]; total: number }> {
    const params = new URLSearchParams({ q, top: top.toString() });
    if (type) params.append("type", type);
    const res = await fetch(`${API_BASE}/search?${params.toString()}`);
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },

  async listEntries(): Promise<{ entries: IndexEntry[]; total: number }> {
    const res = await fetch(`${API_BASE}/list`);
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },

  async indexText(text: string): Promise<IndexResponse> {
    const res = await fetch(`${API_BASE}/index-text`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text }),
    });
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },

  async uploadFile(file: File, label: string): Promise<IndexResponse> {
    const formData = new FormData();
    formData.append("file", file);
    if (label) formData.append("label", label);

    const res = await fetch(`${API_BASE}/index`, {
      method: "POST",
      body: formData,
    });
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },

  async deleteEntry(idPrefix: string): Promise<{ deleted: number }> {
    const res = await fetch(`${API_BASE}/entries/${idPrefix}`, {
      method: "DELETE",
    });
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },

  async findSimilar(idPrefix: string, top: number = 5): Promise<{ source: IndexEntry; results: SearchResult[] }> {
    const res = await fetch(`${API_BASE}/similar/${idPrefix}?top=${top}`);
    if (!res.ok) throw new Error(await res.text());
    return res.json();
  },
};
