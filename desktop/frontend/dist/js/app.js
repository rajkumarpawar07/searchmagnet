// Application shell: sidebar navigation, view routing, theme toggle and
// the live "Index Stats" card. Views are lazy ES modules under js/views/.
import { el, icon, formatBytes, timeAgo, toast } from "./ui.js";
import { api } from "./api.js";
import { registerNavigator } from "./router.js";

const VIEWS = [
  { id: "dashboard", label: "Dashboard", iconId: "i-dashboard", module: () => import("./views/dashboard.js") },
  { id: "search",    label: "Search",    iconId: "i-search",    module: () => import("./views/search.js") },
  { id: "upload",    label: "Upload",    iconId: "i-upload",    module: () => import("./views/upload.js") },
  { id: "library",   label: "Library",   iconId: "i-library",   module: () => import("./views/library.js") },
  { id: "settings",  label: "Settings",  iconId: "i-settings",  module: () => import("./views/settings.js") },
  { id: "about",     label: "About",     iconId: "i-about",     module: () => import("./views/about.js") },
];

const nav = document.getElementById("nav");
const viewRoot = document.getElementById("view");
let activeId = null;

// Build sidebar navigation buttons.
for (const view of VIEWS) {
  const button = el("button", {
    class: "nav-item",
    type: "button",
    onclick: () => go(view.id),
  }, icon(view.iconId), view.label);
  button.dataset.view = view.id;
  nav.append(button);
}

// Route to the requested view; unknown ids fall back to the dashboard.
async function go(viewId) {
  const view = VIEWS.find((candidate) => candidate.id === viewId) ?? VIEWS[0];
  if (view.id === activeId) return;
  activeId = view.id;

  for (const item of nav.children) {
    item.classList.toggle("active", item.dataset.view === view.id);
  }

  viewRoot.replaceChildren();
  try {
    const { render } = await view.module();
    render(viewRoot);
  } catch (err) {
    console.error(err);
    viewRoot.append(
      el("div", { class: "empty-state" },
        icon("i-alert"),
        el("strong", null, "View failed to load"),
        err.message)
    );
  }
}

registerNavigator(go);

// Theme toggle (persisted in localStorage).
const themeToggle = document.getElementById("theme-toggle");

function applyTheme(theme) {
  document.documentElement.dataset.theme = theme;
  themeToggle.querySelector("use")
    .setAttribute("href", theme === "dark" ? "#i-moon" : "#i-sun");
  localStorage.setItem("theme", theme);
}

themeToggle.addEventListener("click", () => {
  applyTheme(document.documentElement.dataset.theme === "dark" ? "light" : "dark");
});
applyTheme(localStorage.getItem("theme") ?? "dark");

// Sidebar stats card; views dispatch "stats-changed" after index mutations.
async function refreshStats() {
  try {
    const stats = await api.getStats();
    document.getElementById("stat-files").textContent = String(stats.totalEntries);
    document.getElementById("stat-size").textContent = formatBytes(stats.totalBytes);
    document.getElementById("stat-embeddings").textContent = String(stats.totalEntries);
    document.getElementById("stat-updated").textContent = timeAgo(stats.lastUpdated);
  } catch (err) {
    console.error("stats:", err.message);
  }
}

document.addEventListener("stats-changed", refreshStats);

// Warn prominently when no API key is configured yet.
api.getSettings().then((settings) => {
  if (!settings?.apiKey) {
    toast("No Gemini API key configured - open Settings to add one", "error");
  }
});

// Boot on the dashboard and populate the sidebar stats.
refreshStats();
go("dashboard");
