/**
 * Search view: full-width semantic search with type filters and a
 * ranked results list.
 */
import { el, icon, toast } from "../ui.js";
import { api } from "../api.js";
import { resultRow } from "../results.js";

/** Content-type filter chips shown under the search bar. */
export const TYPE_FILTERS = [
  { id: "",     label: "All",       icon: "i-sparkles" },
  { id: "image", label: "Images",   icon: "i-image" },
  { id: "video", label: "Videos",   icon: "i-video" },
  { id: "audio", label: "Audio",    icon: "i-audio" },
  { id: "pdf",   label: "Documents", icon: "i-doc" },
  { id: "text",  label: "Text",     icon: "i-text" },
];

/**
 * Render the search bar + filter chips into `container`.
 *
 * callbacks:
 *   onResults(query, hits) — called with ranked hits after a successful search
 *   onLoading()            — called when a query starts
 *   onError(err)           — called with a readable error on failure
 */
export function renderSearchControls(container, { onResults, onLoading, onError }) {
  let activeFilter = "";

  const input = el("input", {
    type: "search",
    placeholder: "Search anything… (e.g. sunset over mountains, interview audio, research paper)",
    spellcheck: "false",
  });

  const chips = el(
    "div",
    { class: "chips", role: "tablist" },
    TYPE_FILTERS.map((filter) =>
      el("button", {
        class: `chip ${filter.id === "" ? "active" : ""}`,
        type: "button",
        onclick: (e) => {
          activeFilter = filter.id;
          container.querySelectorAll(".chip").forEach((c) => c.classList.remove("active"));
          e.currentTarget.classList.add("active");
          submit(); // re-run the current query under the new filter
        },
      }, icon(filter.icon), filter.label)
    )
  );

  const searchButton = el("button", {
    class: "btn btn-primary",
    type: "submit",
  }, icon("i-search"), "Search");

  async function submit() {
    const query = input.value.trim();
    if (!query) return;

    onLoading?.();
    searchButton.disabled = true;
    try {
      const hits = await api.search(query, activeFilter, 24);
      onResults?.(query, hits ?? []);
    } catch (err) {
      onError?.(err);
    } finally {
      searchButton.disabled = false;
    }
  }

  container.append(
    el("form", {
      class: "search-bar",
      onsubmit: (e) => { e.preventDefault(); submit(); },
    }, icon("i-search"), input, searchButton),
    chips
  );
}

export function render(root) {
  const status = el("div", { style: "margin-top: 20px;" });

  root.append(
    el(
      "header",
      { class: "view-header" },
      el("h1", { class: "view-title" }, "Semantic Search"),
      el("p", { class: "view-subtitle" }, "Describe what you're looking for — no filenames required")
    )
  );

  renderSearchControls(root, {
    onLoading() {
      status.replaceChildren(el("div", { class: "empty-state" }, el("div", { class: "spinner" })));
    },
    onResults(query, hits) {
      if (hits.length === 0) {
        status.replaceChildren(noMatches(query));
        return;
      }
      status.replaceChildren(...hits.map((hit) => resultRow(hit)));
    },
    onError(err) {
      status.replaceChildren();
      toast(err.message, "error");
    },
  });

  root.append(el("div", { class: "results-list" }, status));
}

function noMatches(query) {
  return el(
    "div",
    { class: "empty-state" },
    icon("i-search"),
    el("strong", null, `No matches for \u201C${query}\u201D`),
    "Try describing the content differently."
  );
}
