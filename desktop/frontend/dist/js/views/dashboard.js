/**
 * Dashboard view: welcome header, quick search, type filters and the
 * "Recent Results" grid from the newest index entries.
 */
import { el, icon, toast } from "../ui.js";
import { api } from "../api.js";
import { resultCard } from "../results.js";
import { renderSearchControls } from "./search.js";
import { go } from "../router.js";

const RECENT_LIMIT = 8;

export function render(root) {
  const heading = el("h2", null, "Recent Results");
  const grid = el("div", { class: "results-grid" });

  root.append(
    el(
      "header",
      { class: "welcome" },
      el("h1", null, "Welcome back \u{1F44B}"),
      el("p", null, "Search across your local files using natural language")
    ),
    searchControls(heading, grid),
    el(
      "div",
      { class: "section-head" },
      heading,
      el("a", {
        class: "link",
        href: "#library",
        textContent: "View all",
        onclick: (e) => { e.preventDefault(); go("library"); },
      })
    ),
    grid
  );

  loadRecent(grid);
}

/**
 * Search controls that render their results straight into the dashboard
 * grid, replacing the recent list until the view is revisited.
 */
function searchControls(heading, grid) {
  const container = el("section");
  renderSearchControls(container, {
    onResults(query, hits) {
      heading.textContent = `Results for \u201C${query}\u201D`;
      if (hits.length === 0) {
        grid.replaceChildren(noMatches());
      } else {
        grid.replaceChildren(...hits.map((hit) => resultCard(hit.entry, hit.score)));
      }
    },
    onError: (err) => toast(err.message, "error"),
  });
  return container;
}

async function loadRecent(grid) {
  grid.append(el("div", { style: "grid-column: 1/-1;" }, spinnerState()));
  try {
    const entries = await api.recentEntries(RECENT_LIMIT);
    if (entries.length === 0) {
      grid.replaceWith(emptyIndexState());
      return;
    }
    grid.replaceChildren(...entries.map((entry) => resultCard(entry)));
  } catch (err) {
    grid.replaceChildren();
    toast(err.message, "error");
  }
}

function spinnerState() {
  return el("div", { class: "empty-state" }, el("div", { class: "spinner" }));
}

function noMatches() {
  return el(
    "div",
    { class: "empty-state", style: "grid-column: 1/-1;" },
    icon("i-search"),
    el("strong", null, "No matches"),
    "Try a different description or another content type filter."
  );
}

function emptyIndexState() {
  return el(
    "div",
    { class: "empty-state", style: "grid-column: 1/-1;" },
    icon("i-sparkles"),
    el("strong", null, "Your index is empty"),
    "Upload images, videos, audio or documents to start searching.",
    el("br"),
    el("a", {
      class: "btn btn-primary",
      href: "#upload",
      style: "margin-top:14px;",
      onclick: (e) => { e.preventDefault(); go("upload"); },
    }, icon("i-upload"), "Add files")
  );
}
