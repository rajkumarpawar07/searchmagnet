/**
 * Settings view: Gemini API key + embedding model, persisted by the Go
 * backend, plus diagnostics (index location, key status).
 */
import { el, icon, toast } from "../ui.js";
import { api } from "../api.js";

export function render(root) {
  const apiKey = el("input", { type: "password", placeholder: "AIza…", spellcheck: "false" });
  const model = el("input", { type: "text", spellcheck: "false" });

  const diagList = el("dl", { class: "diag-list" });

  const saveButton = el("button", {
    class: "btn btn-primary",
    onclick: () => save(saveButton),
  }, icon("i-check"), "Save settings");

  async function save(button) {
    button.disabled = true;
    try {
      await api.saveSettings(apiKey.value.trim(), model.value.trim());
      toast("Settings saved — connection verified", "success");
      document.dispatchEvent(new CustomEvent("stats-changed"));
      loadDiagnostics();
    } catch (err) {
      toast(err.message, "error");
    } finally {
      button.disabled = false;
    }
  }

  const revealButton = el("button", {
    class: "icon-btn reveal",
    title: "Show / hide key",
    onclick: () => {
      apiKey.type = apiKey.type === "password" ? "text" : "password";
      revealButton.replaceChildren(icon(apiKey.type === "password" ? "i-eye" : "i-eye-off"));
    },
  }, icon("i-eye"));

  async function loadDiagnostics() {
    try {
      const stats = await api.getStats();
      diagList.replaceChildren(
        diagRow("API key", stats.hasApiKey
          ? el("span", { class: "pill-ok" }, "Configured")
          : el("span", { class: "pill-missing" }, "Missing")),
        diagRow("Embedding model", stats.model),
        diagRow("Total entries", String(stats.totalEntries)),
        diagRow("Index file", stats.indexFile),
      );
    } catch (err) {
      toast(err.message, "error");
    }
  }

  root.append(
    el("header", { class: "view-header" },
      el("h1", { class: "view-title" }, "Settings"),
      el("p", { class: "view-subtitle" }, "Stored locally in your user profile — never synced")
    ),
    el("section", { class: "panel settings-form" },
      field("Gemini API key", apiKey, revealButton,
        "Get a free key at aistudio.google.com/apikey. Used only to generate embeddings."),
      field("Embedding model", model, null,
        "Default: gemini-embedding-2-preview"),
      saveButton
    ),
    el("h2", { style: "font-size:16px; margin:24px 0 12px;" }, "Diagnostics"),
    diagList
  );

  // Hydrate current values.
  api.getSettings().then((settings) => {
    if (!settings) return;
    apiKey.value = settings.apiKey ?? "";
    model.value = settings.model ?? "";
  });
  loadDiagnostics();
}

function field(labelText, input, trailing, hint) {
  const group = el("div", { class: "input-group" }, input);
  if (trailing) group.append(trailing);
  return el("div", { class: "field" },
    el("label", null, labelText),
    group,
    hint ? el("div", { class: "hint" }, hint) : null
  );
}

function diagRow(name, value) {
  return el("div", null, el("dt", null, name), el("dd", { title: typeof value === "string" ? value : "" }, value));
}
