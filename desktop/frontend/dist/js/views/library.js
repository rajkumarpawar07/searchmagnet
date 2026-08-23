/**
 * Library view: every indexed entry with filtering and per-row actions
 * (open, reveal in Explorer, delete).
 */
import { el, icon, toast, formatBytes, timeAgo, typeMeta } from "../ui.js";
import { api } from "../api.js";
import { openEntry } from "../results.js";

export function render(root) {
  const filterInput = el("input", {
    type: "search",
    class: "field-input",
    placeholder: "Filter by name…",
    spellcheck: "false",
    style: "max-width: 320px;",
  });
  Object.assign(filterInput.style, {
    background: "var(--panel)", border: "1px solid var(--border)",
    borderRadius: "10px", padding: "9px 12px", outline: "none", userSelect: "text",
  });

  const list = el("div", { class: "library-table" });

  let entries = [];

  function applyFilter() {
    const needle = filterInput.value.trim().toLowerCase();
    const visible = needle
      ? entries.filter((e) => e.label.toLowerCase().includes(needle))
      : entries;
    list.replaceChildren(...visible.map(row));
  }

  async function refresh() {
    list.replaceChildren(loadingState());
    try {
      entries = await api.listEntries();
      applyFilter();
    } catch (err) {
      list.replaceChildren();
      toast(err.message, "error");
    }
  }

  filterInput.addEventListener("input", applyFilter);

  root.append(
    el("header", { class: "view-header", style: "display:flex; justify-content:space-between; align-items:flex-end; gap:16px;" },
      el("div", null,
        el("h1", { class: "view-title" }, "Library"),
        el("p", { class: "view-subtitle" }, "Everything stored in your local index")
      ),
      el("div", { style: "display:flex; gap:10px;" },
        filterInput,
        el("button", { class: "btn btn-ghost btn-sm", onclick: refresh }, icon("i-refresh"), "Refresh")
      )
    ),
    list
  );

  refresh();
}

function loadingState() {
  return el("div", { class: "empty-state" }, el("div", { class: "spinner" }));
}

function row(entry) {
  const meta = typeMeta(entry.contentType);

  return el("div", { class: "lib-row", title: entry.filePath },
    el("span", { class: "lib-type" }, icon(meta.icon)),
    el("div", null,
      el("div", { class: "lib-name" }, entry.label),
      el("div", { class: "type-label" }, meta.label)
    ),
    el("div", { class: "lib-path selectable" }, entry.filePath),
    el("div", { class: "type-label" }, formatBytes(entry.sizeBytes)),
    el("div", { class: "type-label" }, timeAgo(entry.indexedAt)),
    el("div", { class: "lib-actions" },
      el("button", { class: "icon-btn", title: "Open file", onclick: () => openEntry(entry) },
        icon("i-open")),
      el("button", {
        class: "icon-btn",
        title: "Show in folder",
        onclick: () => api.revealFile(entry.filePath).catch((err) => toast(err.message, "error")),
      }, icon("i-folder")),
      el("button", {
        class: "icon-btn danger",
        title: "Remove from index",
        onclick: (e) => removeEntry(entry, e.currentTarget),
      }, icon("i-trash"))
    )
  );
}

async function removeEntry(entry, button) {
  if (!confirm(`Remove \u201C${entry.label}\u201D from the index?\nThe file itself stays on disk.`)) return;

  button.disabled = true;
  try {
    await api.deleteEntry(entry.id);
    button.closest(".lib-row").remove();
    toast("Entry removed", "success");
    document.dispatchEvent(new CustomEvent("stats-changed"));
  } catch (err) {
    toast(err.message, "error");
    button.disabled = false;
  }
}
