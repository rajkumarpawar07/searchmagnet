"use client";

import { useState, useEffect } from "react";
import { api, Stats, SearchResult, IndexEntry } from "@/lib/api";
import { 
  Search, 
  UploadCloud, 
  Database, 
  FileText, 
  Image as ImageIcon, 
  Video, 
  Music, 
  File, 
  Trash2, 
  RefreshCcw,
  Sparkles,
  ArrowRight
} from "lucide-react";

export default function Home() {
  const [activeTab, setActiveTab] = useState<"search" | "index" | "database">("search");
  const [stats, setStats] = useState<Stats | null>(null);

  useEffect(() => {
    fetchStats();
  }, []);

  const fetchStats = async () => {
    try {
      const data = await api.getStats();
      setStats(data);
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 font-sans selection:bg-zinc-800">
      {/* Top Navigation */}
      <header className="border-b border-zinc-900 bg-zinc-950/50 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-5xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-zinc-100 rounded-md flex items-center justify-center text-zinc-950">
              <Sparkles className="w-5 h-5" />
            </div>
            <span className="font-semibold text-lg tracking-tight">SearchMagnet</span>
          </div>

          {stats && (
            <div className="flex items-center gap-6 text-sm">
              <div className="flex items-center gap-2 text-zinc-400">
                <span>Model:</span>
                <span className="text-zinc-200 font-medium bg-zinc-900 px-2 py-0.5 rounded border border-zinc-800">
                  {stats.embedding_model.replace("gemini-embedding-", "")}
                </span>
              </div>
              <div className="flex items-center gap-2 text-zinc-400">
                <span>Entries:</span>
                <span className="text-zinc-200 font-medium">{stats.total_entries}</span>
              </div>
            </div>
          )}
        </div>
      </header>

      {/* Main Layout */}
      <main className="max-w-5xl mx-auto px-6 py-8 flex flex-col md:flex-row gap-8">
        
        {/* Sidebar Navigation */}
        <aside className="w-full md:w-56 shrink-0 flex flex-col gap-1">
          <p className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-2 px-3">Menu</p>
          <TabButton 
            active={activeTab === "search"} 
            onClick={() => setActiveTab("search")} 
            icon={<Search className="w-4 h-4" />}
            label="Search" 
          />
          <TabButton 
            active={activeTab === "index"} 
            onClick={() => setActiveTab("index")} 
            icon={<UploadCloud className="w-4 h-4" />}
            label="Index Content" 
          />
          <TabButton 
            active={activeTab === "database"} 
            onClick={() => setActiveTab("database")} 
            icon={<Database className="w-4 h-4" />}
            label="Database" 
          />
        </aside>

        {/* Content Area */}
        <section className="flex-1 min-w-0">
          <div className="animate-in fade-in slide-in-from-bottom-2 duration-300 ease-out">
            {activeTab === "search" && <SearchTab onIndexUpdated={fetchStats} />}
            {activeTab === "index" && <IndexTab onIndexUpdated={fetchStats} />}
            {activeTab === "database" && <DatabaseTab onIndexUpdated={fetchStats} />}
          </div>
        </section>
      </main>
    </div>
  );
}

function TabButton({ active, onClick, icon, label }: { active: boolean, onClick: () => void, icon: React.ReactNode, label: string }) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
        active 
          ? "bg-zinc-900 text-zinc-100 border border-zinc-800" 
          : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/50 border border-transparent"
      }`}
    >
      {icon}
      {label}
    </button>
  );
}

// ──────────────────────────────────────────────────────────────
// Search Tab
// ──────────────────────────────────────────────────────────────

function SearchTab({ onIndexUpdated }: { onIndexUpdated: () => void }) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [topK, setTopK] = useState(5);
  const [typeFilter, setTypeFilter] = useState("");

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query) return;
    setLoading(true);
    try {
      const res = await api.search(query, topK, typeFilter);
      setResults(res.results || []);
    } catch (e: any) {
      alert("Search failed: " + e.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSimilar = async (id: string) => {
    setLoading(true);
    try {
      const res = await api.findSimilar(id, topK);
      setResults(res.results || []);
    } catch (e: any) {
      alert("Find similar failed: " + e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-2xl font-semibold mb-1">Search</h2>
        <p className="text-zinc-400 text-sm">Query your multimodal embeddings using natural language.</p>
      </div>

      <div className="bg-zinc-900/40 border border-zinc-800/80 rounded-xl p-1">
        <form onSubmit={handleSearch} className="flex flex-col sm:flex-row gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-zinc-500" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="e.g. A cat playing in the grass..."
              className="w-full bg-zinc-900 border border-zinc-800 rounded-lg pl-11 pr-4 py-3 text-sm focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-shadow placeholder:text-zinc-500"
            />
          </div>
          
          <div className="flex gap-2 p-1 sm:p-0">
            <select 
              value={typeFilter} 
              onChange={(e) => setTypeFilter(e.target.value)}
              className="bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-3 text-sm focus:outline-none focus:ring-1 focus:ring-zinc-700 text-zinc-300"
            >
              <option value="">All Types</option>
              <option value="image">Images</option>
              <option value="video">Videos</option>
              <option value="audio">Audio</option>
              <option value="text">Text</option>
            </select>
            
            <button 
              type="submit" 
              disabled={loading || !query}
              className="bg-zinc-100 text-zinc-950 hover:bg-white px-6 py-3 rounded-lg text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
            >
              {loading ? "Searching..." : "Search"}
            </button>
          </div>
        </form>
      </div>

      {results.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between border-b border-zinc-900 pb-2">
            <h3 className="text-sm font-medium text-zinc-400">Results ({results.length})</h3>
            <div className="flex items-center gap-2 text-sm text-zinc-500">
              <span>Top K:</span>
              <input 
                type="number" 
                min="1" max="20" 
                value={topK}
                onChange={(e) => setTopK(parseInt(e.target.value))}
                className="bg-transparent border-b border-zinc-700 w-12 focus:outline-none focus:border-zinc-500 text-center"
              />
            </div>
          </div>
          <div className="grid gap-3">
            {results.map((res) => (
              <ResultCard key={res.id} result={res} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function getTypeIcon(type: string) {
  switch (type) {
    case "image": return <ImageIcon className="w-4 h-4 text-zinc-400" />;
    case "video": return <Video className="w-4 h-4 text-zinc-400" />;
    case "audio": return <Music className="w-4 h-4 text-zinc-400" />;
    case "text": return <FileText className="w-4 h-4 text-zinc-400" />;
    default: return <File className="w-4 h-4 text-zinc-400" />;
  }
}

function ResultCard({ result }: { result: SearchResult }) {
  // Map score to a simple percentage 
  const scorePercent = (result.score * 100).toFixed(1);
  
  // Professional minimal score indicator
  let scoreColor = "text-zinc-500";
  if (result.score >= 0.7) scoreColor = "text-emerald-500";
  else if (result.score >= 0.4) scoreColor = "text-amber-500";

  const fileUrl = `http://localhost:3000/api/file?path=${encodeURIComponent(result.file_path)}`;

  return (
    <div className="bg-zinc-900/40 border border-zinc-800/80 rounded-xl overflow-hidden transition-colors hover:bg-zinc-900/60 flex flex-col sm:flex-row h-[120px]">
      {/* Preview Section */}
      <div className="w-full sm:w-48 h-full shrink-0 bg-zinc-950 flex items-center justify-center border-b sm:border-b-0 sm:border-r border-zinc-800/50 relative overflow-hidden">
        {result.content_type === "image" && (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={fileUrl} alt={result.label} className="absolute inset-0 w-full h-full object-cover" />
        )}
        {result.content_type === "video" && (
          <video src={fileUrl} controls className="absolute inset-0 w-full h-full object-cover bg-black" />
        )}
        {result.content_type === "audio" && (
          <div className="w-full px-2">
            <audio src={fileUrl} controls className="w-full h-8" />
          </div>
        )}
        {result.content_type === "text" && (
          <div className="p-3 text-[10px] text-zinc-400 font-mono overflow-hidden h-full whitespace-pre-wrap text-left w-full break-all leading-tight">
             {result.label}
          </div>
        )}
        {/* Fallback if no preview matched */}
        {result.content_type !== "image" && result.content_type !== "video" && result.content_type !== "audio" && result.content_type !== "text" && (
          <File className="w-8 h-8 text-zinc-700" />
        )}
      </div>

      {/* Info Section */}
      <div className="p-4 flex flex-col justify-between flex-1 min-w-0">
        <div>
          <div className="flex items-center gap-2 mb-1">
            {getTypeIcon(result.content_type)}
            <span className="text-xs uppercase tracking-wider text-zinc-500 font-medium">
              {result.content_type}
            </span>
          </div>
          <h4 className="font-medium text-zinc-200 line-clamp-1" title={result.label}>{result.label}</h4>
          {result.file_path && result.content_type !== "text" && (
            <p className="text-xs text-zinc-500 mt-1 truncate" title={result.file_path}>{result.file_path}</p>
          )}
        </div>
        
        <div className="mt-auto flex items-end justify-between">
          <div>
            <span className="text-[10px] uppercase tracking-wider text-zinc-500 font-semibold block mb-0.5">Similarity</span>
            <span className={`text-sm font-mono font-medium ${scoreColor}`}>{scorePercent}%</span>
          </div>
        </div>
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────
// Index Tab
// ──────────────────────────────────────────────────────────────

function IndexTab({ onIndexUpdated }: { onIndexUpdated: () => void }) {
  const [file, setFile] = useState<File | null>(null);
  const [label, setLabel] = useState("");
  const [text, setText] = useState("");
  const [loading, setLoading] = useState(false);

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return;
    setLoading(true);
    try {
      await api.uploadFile(file, label);
      alert("File indexed successfully!");
      setFile(null);
      setLabel("");
      onIndexUpdated();
    } catch (e: any) {
      alert("Upload failed: " + e.message);
    } finally {
      setLoading(false);
    }
  };

  const handleTextIndex = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!text) return;
    setLoading(true);
    try {
      await api.indexText(text);
      alert("Text indexed successfully!");
      setText("");
      onIndexUpdated();
    } catch (e: any) {
      alert("Indexing failed: " + e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-8 max-w-3xl">
      <div>
        <h2 className="text-2xl font-semibold mb-1">Index Content</h2>
        <p className="text-zinc-400 text-sm">Add new files or text to the semantic database.</p>
      </div>

      <div className="bg-zinc-900/40 border border-zinc-800/80 rounded-xl p-6">
        <div className="flex items-center gap-3 mb-6 pb-4 border-b border-zinc-800/50">
          <div className="bg-zinc-900 border border-zinc-800 p-2 rounded-lg">
            <UploadCloud className="w-5 h-5 text-zinc-300" />
          </div>
          <div>
            <h3 className="font-medium text-zinc-200">Upload Media</h3>
            <p className="text-xs text-zinc-500">Images, Video, Audio, or PDF (Max 20MB)</p>
          </div>
        </div>

        <form onSubmit={handleUpload} className="space-y-5">
          <div className="border border-dashed border-zinc-700 bg-zinc-900/50 rounded-xl p-8 text-center hover:bg-zinc-900 transition-colors">
            <input 
              type="file" 
              onChange={(e) => setFile(e.target.files?.[0] || null)}
              className="hidden" 
              id="file-upload" 
            />
            <label htmlFor="file-upload" className="cursor-pointer flex flex-col items-center">
              <UploadCloud className="w-8 h-8 text-zinc-500 mb-3" />
              <span className="text-sm font-medium text-zinc-300 mb-1">
                {file ? file.name : "Click to select a file"}
              </span>
              <span className="text-xs text-zinc-500">
                {file ? "Click to change" : "or drag and drop here"}
              </span>
            </label>
          </div>
          
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-zinc-400">Custom Label (optional)</label>
            <input 
              type="text" 
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              placeholder="Provide context or a descriptive name"
              className="w-full bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-zinc-700 placeholder:text-zinc-600"
            />
          </div>
          
          <button 
            type="submit" 
            disabled={loading || !file}
            className="w-full bg-zinc-100 text-zinc-950 hover:bg-white px-4 py-2.5 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
          >
            {loading ? "Processing..." : "Index File"}
          </button>
        </form>
      </div>

      <div className="bg-zinc-900/40 border border-zinc-800/80 rounded-xl p-6">
        <div className="flex items-center gap-3 mb-6 pb-4 border-b border-zinc-800/50">
          <div className="bg-zinc-900 border border-zinc-800 p-2 rounded-lg">
            <FileText className="w-5 h-5 text-zinc-300" />
          </div>
          <div>
            <h3 className="font-medium text-zinc-200">Raw Text</h3>
            <p className="text-xs text-zinc-500">Snippets, paragraphs, or code blocks</p>
          </div>
        </div>

        <form onSubmit={handleTextIndex} className="space-y-5">
          <div className="space-y-1.5">
            <textarea 
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="Paste content here..."
              rows={5}
              className="w-full bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-3 text-sm focus:outline-none focus:ring-1 focus:ring-zinc-700 resize-none placeholder:text-zinc-600"
            />
          </div>
          
          <button 
            type="submit" 
            disabled={loading || !text}
            className="w-full bg-zinc-100 text-zinc-950 hover:bg-white px-4 py-2.5 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
          >
            {loading ? "Processing..." : "Index Text"}
          </button>
        </form>
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────
// Database Tab
// ──────────────────────────────────────────────────────────────

function DatabaseTab({ onIndexUpdated }: { onIndexUpdated: () => void }) {
  const [entries, setEntries] = useState<IndexEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchEntries();
  }, []);

  const fetchEntries = async () => {
    setLoading(true);
    try {
      const res = await api.listEntries();
      setEntries(res.entries || []);
    } catch (e: any) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this entry?")) return;
    try {
      await api.deleteEntry(id);
      fetchEntries();
      onIndexUpdated();
    } catch (e: any) {
      alert("Delete failed: " + e.message);
    }
  };

  return (
    <div className="space-y-8 max-w-4xl">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold mb-1">Database</h2>
          <p className="text-zinc-400 text-sm">Manage your indexed embeddings.</p>
        </div>
        <button 
          onClick={fetchEntries} 
          className="p-2 text-zinc-400 hover:text-zinc-100 bg-zinc-900/50 hover:bg-zinc-800 rounded-lg border border-zinc-800 transition-colors"
          title="Refresh"
        >
          <RefreshCcw className="w-4 h-4" />
        </button>
      </div>
      
      <div className="bg-zinc-900/40 border border-zinc-800/80 rounded-xl overflow-hidden">
        {loading ? (
          <div className="p-12 text-center text-sm text-zinc-500">Loading...</div>
        ) : entries.length === 0 ? (
          <div className="p-12 text-center text-sm text-zinc-500 flex flex-col items-center">
            <Database className="w-8 h-8 mb-3 opacity-20" />
            Database is currently empty.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm whitespace-nowrap">
              <thead className="text-xs uppercase tracking-wider text-zinc-500 bg-zinc-900/80 border-b border-zinc-800">
                <tr>
                  <th className="px-6 py-4 font-medium">Type</th>
                  <th className="px-6 py-4 font-medium">Label</th>
                  <th className="px-6 py-4 font-medium">File</th>
                  <th className="px-6 py-4 font-medium">Date</th>
                  <th className="px-6 py-4 font-medium text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/50">
                {entries.map((e) => (
                  <tr key={e.id} className="hover:bg-zinc-900/50 transition-colors">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        {getTypeIcon(e.content_type)}
                        <span className="capitalize text-zinc-300">{e.content_type}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-zinc-200 max-w-[200px] truncate" title={e.label}>{e.label}</td>
                    <td className="px-6 py-4 text-zinc-500 max-w-[200px] truncate" title={e.file_path}>{e.file_path || "—"}</td>
                    <td className="px-6 py-4 text-zinc-500 font-mono text-[11px]">{new Date(e.indexed_at).toLocaleDateString()}</td>
                    <td className="px-6 py-4 text-right">
                      <button 
                        onClick={() => handleDelete(e.id)}
                        className="p-1.5 text-zinc-500 hover:text-rose-400 hover:bg-rose-500/10 rounded transition-colors"
                        title="Delete entry"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
