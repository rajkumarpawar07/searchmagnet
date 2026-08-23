/**
 * Small DOM & formatting helpers shared by all views.
 * Keeps view code declarative and free of boilerplate.
 */

/** Create an element: el("div", { class: "x", onclick }, child, ...) */
export function el(tag, attrs = {}, ...children) {
  const node = document.createElement(tag);
  for (const [key, value] of Object.entries(attrs ?? {})) {
    if (value == null || value === false) continue;
    if (key.startsWith("on") && typeof value === "function") {
      node.addEventListener(key.slice(2).toLowerCase(), value);
    } else if (key === "html") {
      node.innerHTML = value; // only used with static markup in this codebase
    } else if (value === true) {
      node.setAttribute(key, "");
    } else {
      node.setAttribute(key, value);
    }
  }
  append(node, children);
  return node;
}

function append(node, children) {
  for (const child of children.flat()) {
    if (child == null || child === false) continue;
    node.append(child instanceof Node ? child : document.createTextNode(child));
  }
}

/** <svg><use href="#id"/></svg> from the sprite in index.html */
export function icon(name) {
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("class", "icon");
  const use = document.createElementNS("http://www.w3.org/2000/svg", "use");
  use.setAttribute("href", `#${name}`);
  svg.append(use);
  return svg;
}

/** 24.3 MB style byte formatting */
export function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return "—";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  return `${value >= 100 || unit === 0 ? Math.round(value) : value.toFixed(1)} ${units[unit]}`;
}

/** Human friendly timestamp ("Just now", "5m ago", "Jun 10") */
export function timeAgo(iso) {
  if (!iso) return "—";
  const then = new Date(iso).getTime();
  const seconds = Math.floor((Date.now() - then) / 1000);
  if (seconds < 45) return "Just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 7 * 86400) return `${Math.floor(seconds / 86400)}d ago`;
  return new Date(iso).toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

/** Metadata for each content type: icon id + display label */
export function typeMeta(contentType) {
  const meta = {
    image: { icon: "i-image", label: "Image" },
    video: { icon: "i-video", label: "Video" },
    audio: { icon: "i-audio", label: "Audio" },
    pdf:   { icon: "i-doc",   label: "PDF" },
    text:  { icon: "i-text",  label: "Text" },
  };
  return meta[contentType] ?? { icon: "i-doc", label: contentType };
}

export function scoreBadge(score) {
  return el("span", { class: "score-badge" }, score.toFixed(2));
}

export function toast(message, kind = "info") {
  const icons = { info: "i-sparkles", success: "i-check", error: "i-alert" };
  const box = el("div", { class: `toast ${kind}` }, icon(icons[kind] ?? icons.info), message);
  document.getElementById("toasts").append(box);
  setTimeout(() => box.remove(), kind === "error" ? 7000 : 4000);
}

export function spinner() {
  return el("span", { class: "spinner", role: "status", "aria-label": "Loading" });
}
