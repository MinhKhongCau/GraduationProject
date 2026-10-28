/*
 * MindCare service worker.
 * - CSS / JS / fonts / images: Cache First (file names are content-hashed, so they never go stale).
 * - Pages (HTML): Network First, cache only as an offline fallback. Pages must hit the server so
 *   proxy.ts can apply the login redirects; a cached page skipped them and caused login loops.
 */
const CACHE_PREFIX = "mindcare-";
const CACHE = `${CACHE_PREFIX}v2`;
const PRECACHE_URLS = [
  "/manifest.webmanifest",
  "/icons/icon-192.png",
  "/icons/icon-512.png",
  "/icons/apple-touch-icon.png",
];

// Never cache API calls, sockets or dev hot-reload traffic.
const BYPASS_PREFIXES = ["/api/", "/socket.io/", "/_next/webpack-hmr", "/_next/data/"];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches
      .open(CACHE)
      .then((cache) => cache.addAll(PRECACHE_URLS))
      .catch(() => undefined)
      .then(() => self.skipWaiting())
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((k) => k.startsWith(CACHE_PREFIX) && k !== CACHE).map((k) => caches.delete(k))))
      .then(() => self.clients.claim())
  );
});

function isStaticAsset(url, request) {
  return (
    url.pathname.startsWith("/_next/static/") ||
    url.pathname.startsWith("/icons/") ||
    url.pathname.startsWith("/images/") ||
    /\.(?:js|css|woff2?|ttf|png|jpg|jpeg|svg|webp|ico)$/.test(url.pathname) ||
    ["script", "style", "font"].includes(request.destination)
  );
}

function isCacheable(response) {
  return response && response.ok && response.type === "basic";
}

async function putInCache(request, response) {
  if (!isCacheable(response)) return;
  const cache = await caches.open(CACHE);
  await cache.put(request, response);
}

// Cache First: serve from cache, otherwise go to the network and store the result.
async function cacheFirst(request) {
  const cached = await caches.match(request);
  if (cached) return cached;
  const response = await fetch(request);
  putInCache(request, response.clone());
  return response;
}

// Network First for pages; the cached copy is only used when the device is offline.
async function networkFirstPage(request) {
  try {
    const response = await fetch(request);
    // Redirects (e.g. proxy.ts sending a user to /auth/login) are never cached.
    if (response.type === "basic" && !response.redirected) putInCache(request, response.clone());
    return response;
  } catch {
    const cached = await caches.match(request);
    return cached || Response.error();
  }
}

self.addEventListener("fetch", (event) => {
  const { request } = event;
  if (request.method !== "GET") return;

  const url = new URL(request.url);
  if (url.origin !== self.location.origin) return;
  if (BYPASS_PREFIXES.some((p) => url.pathname.startsWith(p))) return;
  // Next.js client navigations (RSC payloads) must stay fresh.
  if (request.headers.get("RSC") || url.searchParams.has("_rsc")) return;

  if (request.mode === "navigate") {
    event.respondWith(networkFirstPage(request));
    return;
  }
  if (isStaticAsset(url, request)) {
    event.respondWith(cacheFirst(request));
  }
});
