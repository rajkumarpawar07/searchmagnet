/**
 * Upload view: file/folder pickers, drag & drop indexing, text snippets,
 * plus a live progress panel fed by "indexer:progress" backend events.
 */
import { el, icon, toast, formatBytes } from "../ui.js";
import { api, onEvent } from "../api.js";

export function render(root) {
  const progress = buildProgressPanel();

  const dropzone = el("div", { class: "dropzone" },
    icon("i-upload"),
    el("strong", null, "Drop files anywhere or click to browse"),
    el("span", null, "Images · Videos · Audio · PDFs — up to 20 MB per file")
  );
  dropzone.addEventListener("click", () => runIndexing(dropzone, progress, () => api.pickFilesAndIndex()));

  root.append(
    el("header", { class: "view-header" },
      el("h1", { class: "view-title" }, "Add to Index"),
      el("p", { class: "view-subtitle" }, "Files are embedded locally via Gemini and stored in your private index")
    ),
    el("div", { class: "upload-grid" },
      el("div", null,
        el("section", { class: "panel" },
          el("h3", null, "Files & folders"),
          el("p", { class: "panel-hint" }, "Duplicates are skipped automatically."),
          dropzone,
          el("div", { class: "upload-actions" },
            el("button", {
              class: "btn btn-ghost",
              onclick: (e) => runIndexing(e.currentTarget, progress, () => api.pickFolderAndIndex()),
            }, icon("i-folder"), "Index a folder…")
          )
        ),
        progress.panel
      ),
      el("section", { class: "panel" },
        el("h3", null, "Text snippet"),
        el("p", { class: "panel-hint" }, "Notes, transcripts, code — anything textual becomes searchable."),
        snippetCard()
      )
    )
  );

  // Live progress from the Go indexer (all indexing entry points feed this).
  onEvent("indexer:progress", (event) => {
    if (!event || !event.total) return;
    if (!progress.panel.classList.contains("visible")) progress.panel.classList.add("visible");
    progress.step(event);
  });

  // Files dropped onto the window land here (Wails DragAndDrop option).
  onEvent("wails:file-drop", (...args) => {
    const paths = extractDroppedPaths(args);
    if (paths.length === 0) return;
    api.indexDroppedFiles(paths).then((summary) => finishRun(progress, summary));
  });
}

/* ── actions ───────────────────────────────────────────────── */

async function runIndexing(button, progress, action) {
  button.disabled = true;
  try {
    const summary = await action();
    if (summary.indexed + summary.skipped + summary.failed === 0) return; // dialog cancelled
    finishRun(progress, summary);
  } catch (err) {
    toast(err.message, "error");
  } finally {
    button.disabled = false;
  }
}

function snippetCard() {
  const textarea = el("textarea", {
    class: "snippet",
    placeholder: "Paste any text to index…",
    spellcheck: "false",
  });

  async function indexText(button) {
    const text = textarea.value.trim();
    if (!text) return;
    button.disabled = true;
    try {
      await api.indexTextSnippet(text);
      textarea.value = "";
      toast("Snippet added to the index", "success");
      bumpStats();
    } catch (err) {
      toast(err.message, "error");
    } finally {
      button.disabled = false;
    }
  }

  const saveButton = el("button", {
    class: "btn btn-primary",
    onclick: (e) => indexText(e.currentTarget),
  }, icon("i-sparkles"), "Index text");

  return el("div", null, textarea,
    el("div", { class: "upload-actions" }, saveButton));
}

/* ── progress panel ────────────────────────────────────────── */

function buildProgressPanel() {
  const fill = el("div", { class: "progress-fill" });
  const labelLeft = el("span", null, "");
  const labelRight = el("span", null, "");
  const log = el("div", { class: "index-log" });

  return {
    panel: el("section", { class: "panel progress-panel" },
      el("h3", null, "Indexing…"),
      el("div", { class: "progress-track" }, fill),
      el("div", { class: "progress-label" }, labelLeft, labelRight),
      log
    ),
    begin(total) {
      fill.style.width = "0%";
      labelLeft.textContent = `0 / ${total} files`;
      labelRight.textContent = "";
      log.replaceChildren();
      this.panel.classList.add("visible");
      this.panel.querySelector("h3").textContent = "Indexing…";
    },
    step(event) {
      fill.style.width = `${Math.round((event.completed / event.total) * 100)}%`;
      labelLeft.textContent = `${event.completed} / ${event.total} files`;
      labelRight.textContent = event.file;
      log.append(el("div", { class: `log-line ${event.status}` },
        el("span", { class: "name" }, event.file),
        el("span", null, statusWord(event))
      ));
      log.scrollTop = log.scrollHeight;
    },
    end(summary) {
      this.panel.querySelector("h3").textContent =
        `Done — ${summary.indexed} indexed, ${summary.skipped} skipped, ${summary.failed} failed`;
      setTimeout(() => this.panel.classList.remove("visible"), 6000);
    },
  };
}

function statusWord(event) {
  switch (event.status) {
    case "done":    return "indexed";
    case "skipped": return event.error ?? "skipped";
    default:        return `failed: ${event.error ?? "unknown error"}`;
  }
}

function finishRun(progress, summary) {
  progress.end(summary);
  if (summary.failed > 0) {
    toast(`${summary.failed} file(s) failed — see log`, "error");
  } else if (summary.indexed > 0) {
    toast(`Indexed ${summary.indexed} new file(s)`, "success");
  } else {
    toast(summary.skipped > 0 ? "Nothing new — everything was already indexed" : "No supported files found", "info");
  }
  if (summary.indexed > 0) bumpStats();
}

/** Wails delivers (x, y, [...paths]) on Windows; normalise all platforms. */
function extractDroppedPaths(args) {
  const flat = args.flat();
  return flat.filter((item) => typeof item === "string");
}
