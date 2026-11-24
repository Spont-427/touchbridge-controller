const CACHE = 'touchbridge-shell-v9';
const ASSETS = ['/', '/app.css', '/app.js', '/theme.js', '/gamepad-core.js', '/manifest.webmanifest', '/icon.svg'];
self.addEventListener('install', event => event.waitUntil(caches.open(CACHE).then(cache => cache.addAll(ASSETS))));
self.addEventListener('activate', event => event.waitUntil(caches.keys().then(keys => Promise.all(keys.filter(key => key !== CACHE).map(key => caches.delete(key))))));
self.addEventListener('fetch', event => {
  if (event.request.method !== 'GET' || new URL(event.request.url).pathname === '/ws') return;
  // Asset URLs carry a ?v= cache-busting query; ignore it when matching so the
  // offline shell still resolves after a version bump.
  event.respondWith(caches.match(event.request, { ignoreSearch: true }).then(cached => cached || fetch(event.request)));
});
