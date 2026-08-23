/**
 * About view: product info and supported formats.
 */
import { el, icon } from "../ui.js";

const SUPPORTED_FORMATS = [
  "PNG", "JPG", "GIF", "WEBP", "BMP", "TIFF",
  "MP4", "MOV", "AVI", "MKV", "WEBM",
  "MP3", "WAV", "OGG", "FLAC", "AAC",
  "PDF",
];

export function render(root) {
  root.append(
    el("div", { class: "about-hero" },
      el("span", { class: "brand-mark" }, icon("i-logo")),
      el("h1", null, "Search", el("em", null, "Magnet")),
      el("p", null, "A high-performance multimodal semantic search engine for your local files."),
      el("p", null, "Powered by Google Gemini embeddings · Built with Wails"),
      el("div", { class: "format-list" },
        SUPPORTED_FORMATS.map((format) => el("code", null, format))
      ),
      el("p", { class: "muted", style: "margin-top:22px; font-size:13px;" },
        "Version 1.0.0 — built with \u2764\uFE0F by Rajkumar Pawar")
    )
  );
}
