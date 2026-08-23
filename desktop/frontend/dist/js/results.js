/**
 * Shared renderers for search results and library entries.
 * Used by the dashboard grid, the search list and (partially) the library.
 */
import { el, icon, formatBytes, typeMeta, scoreBadge, toast } from "./ui.js";
import { api, previewURL } from "./api.js";

/** Decorative waveform placeholder shown on audio tiles. */
function waveform(bars = 26) {
  return el(
    "div",
    { class: "waveform", "aria-hidden": "true" },
    Array.from({ length: bars }, (_, i) =>
      el("i", { style: `height:${18 + Math.round(28 * Math.abs(Math.sin(i * 1.7)))}%` })
    )
  );
}

/** Miniature document tile for pdf/text entries. */
function docTile() {
  return el("div", { class: "doc-tile" }, Array.from({ length: 7 }, () => el("i")));
}

/**
 * Preview element for an entry.
 * Real thumbnails stream through /preview for images; videos load their
 * first frame via preload="metadata". Other types get styled placeholders.
 */
export function thumbnail(entry) {
  const wrap = el("div", { class: "result-thumb" });
  const url = previewURL(entry.filePath);

  switch (entry.contentType) {
    case "image":
      wrap.append(el("img", { src: url, alt: entry.label, loading: "lazy" }));
      break;
    case "video": {
      wrap.append(el("video", { src: `${url}#t=0.1`, preload: "metadata", muted: true }));
      wrap.append(
        el("div", { class: "play-overlay" }, icon("i-play")),
        el("span", { class: "duration-chip" }, "▶")
      );
      break;
    }
    case "audio":
      wrap.append(waveform());
      break;
    default:
      wrap.append(docTile());
  }
  return wrap;
}

/** Small square preview used in list rows. */
export function rowThumbnail(entry) {
  const meta = typeMeta(entry.contentType);
  const wrap = el("div", { class: "row-thumb" });
  if (entry.contentType === "image") {
    wrap.append(el("img", { src: previewURL(entry.filePath), alt: "", loading: "lazy" }));
  } else {
    wrap.append(icon(meta.icon));
  }
  return wrap;
}

/** Grid card for an entry. score is optional (only present after a query). */
export function resultCard(entry, score) {
  const meta = typeMeta(entry.contentType);
  const card = el(
    "div",
    {
      class: "result-card",
      role: "button",
      tabindex: "0",
      title: entry.filePath,
      onclick: () => openEntry(entry),
    },
    thumbnail(entry),
    el(
      "div",
      { class: "result-meta" },
      el("div", { class: "result-name" }, entry.label),
      el(
        "div",
        { class: "result-sub" },
        el("span", { class: "type-label" }, `${meta.label} · ${formatBytes(entry.sizeBytes)}`),
        score != null ? scoreBadge(score) : null
      )
    )
  );
  card.addEventListener("keydown", (e) => e.key === "Enter" && openEntry(entry));
  return card;
}

/** Compact row used on the Search view. */
export function resultRow(hit) {
  const meta = typeMeta(hit.entry.contentType);
  return el(
    "div",
    {
      class: "result-row",
      role: "button",
      tabindex: "0",
      title: hit.entry.filePath,
      onclick: () => openEntry(hit.entry),
    },
    rowThumbnail(hit.entry),
    el(
      "div",
      { class: "row-main" },
      el("div", { class: "row-name" }, hit.entry.label),
      el(
        "div",
        { class: "row-sub" },
        el("span", { class: "rank-chip" }, `#${hit.rank}`),
        el("span", { class: "type-label" }, `${meta.label} · ${formatBytes(hit.entry.sizeBytes)}`)
      )
    ),
    scoreBadge(hit.score)
  );
}

export async function openEntry(entry) {
  try {
    await api.openFile(entry.filePath);
  } catch (err) {
    toast(err.message, "error");
  }
}
