/**
 * Typed wrapper around the Go methods Wails exposes on window.go.main.App.
 * Every call is guarded so a missing backend surfaces as a readable error
 * instead of a TypeError deep inside a view.
 */

function backend() {
  const app = window.go?.main?.App;
  if (!app) throw new Error("Backend not ready — is the app running inside Wails?");
  return app;
}

async function call(action) {
  try {
    return await action();
  } catch (err) {
    // Wails rejects promises with plain strings from Go errors.
    throw new Error(typeof err === "string" ? err : err.message ?? String(err));
  }
}

export const api = {
  search:          (query, filter, topK) => call(() => backend().Search(query, filter, topK)),
  similar:         (id, topK)            => call(() => backend().Similar(id, topK)),

  pickFilesAndIndex: () => call(() => backend().PickFilesAndIndex()),
  pickFolderAndIndex: () => call(() => backend().PickFolderAndIndex()),
  indexDroppedFiles: (paths) => call(() => backend().OnFilesDropped(paths)),
  indexTextSnippet:  (text) => call(() => backend().IndexTextSnippet(text)),

  listEntries:     ()      => call(() => backend().ListEntries()),
  recentEntries:   (limit) => call(() => backend().RecentEntries(limit)),
  deleteEntry:     (id)    => call(() => backend().DeleteEntry(id)),

  getStats:        () => call(() => backend().GetStats()),
  getSettings:     () => call(() => backend().GetSettings()),
  saveSettings:    (apiKey, model) => call(() => backend().SaveSettings(apiKey, model)),

  openFile:        (path) => call(() => backend().OpenFile(path)),
  revealFile:      (path) => call(() => backend().RevealFile(path)),
};

/** URL that streams an indexed file into the webview via the asset handler. */
export function previewURL(filePath) {
  return `/preview?path=${encodeURIComponent(filePath)}`;
}

/** Subscribe to backend events ("indexer:progress", "wails:file-drop", …). */
export function onEvent(name, handler) {
  window.runtime?.EventsOn(name, handler);
}
