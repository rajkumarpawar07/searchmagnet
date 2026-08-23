/**
 * Tiny navigation bus between views.
 * app.js registers the real navigate() implementation at startup;
 * views call go("upload") etc. without importing app.js (avoids cycles).
 */
let navigate = (id) => console.warn(`navigate(${id}) called before router was ready`);

export function registerNavigator(fn) {
  navigate = fn;
}

export function go(viewId) {
  navigate(viewId);
}
