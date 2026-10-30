// Wails runtime stub — production loads real wails.js from /wails binding.
// In a normal Wails build, this file is auto-generated and replaces our stub.
window.go = window.go || {};
window.go.main = window.go.main || {};
window.go.modules = window.go.modules || {};

// ─── Mock bindings for browser preview (no Wails) ───
// When the app runs inside Wails, these are replaced with real bindings
// generated from the Go Bind field declarations in main.go.
const MOCK = {
  App: {
    GetAppInfo: async () => ({
      name: "Công cụ ngành Kiểm sát",
      version: "0.1.0-mvp (browser preview)",
      dataDir: "/tmp/preview",
      platform: navigator.platform,
      buildDate: "2026-10",
      offlineOnly: true,
      modules: ["calc","legal","ocr","pii","spell","file","evidence","case","docs"],
    }),
    OpenDataDir: () => "/tmp/preview",
  },
};
window.go.main.App = window.go.main.App || MOCK.App;

// Helper: call a Wails-bound method safely.
async function wcall(path, ...args) {
  const parts = path.split(".");
  let obj = window;
  for (const p of parts) {
    obj = obj?.[p];
    if (!obj) return null;
  }
  if (typeof obj !== "function") return obj;
  try { return await obj(...args); }
  catch (e) { console.error("wcall", path, e); return null; }
}

// ─── Router ───
const routes = {};
function route(name, renderer) { routes[name] = renderer; }
async function navigate(name) {
  document.querySelectorAll(".nav-item").forEach(b => b.classList.toggle("active", b.dataset.route === name));
  const r = routes[name];
  const page = document.getElementById("page");
  if (!r) { page.innerHTML = `<div class="card"><p>Chưa xây dựng module <code>${name}</code>.</p></div>`; return; }
  page.innerHTML = "<div class='card text-muted'>Đang tải...</div>";
  try { await r(page); }
  catch (e) { page.innerHTML = `<div class='callout callout-error'>Lỗi: ${e}</div>`; }
}

document.addEventListener("click", e => {
  const btn = e.target.closest(".nav-item");
  if (btn?.dataset.route) navigate(btn.dataset.route);
});

// ─── Helpers ───
function $(html) {
  const tpl = document.createElement("template");
  tpl.innerHTML = html.trim();
  return tpl.content.firstElementChild;
}
function money(n) {
  if (n == null || isNaN(n)) return "—";
  return new Intl.NumberFormat("vi-VN").format(Math.round(n)) + "đ";
}
function el(tag, attrs = {}, ...children) {
  const e = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === "class") e.className = v;
    else if (k === "html") e.innerHTML = v;
    else e.setAttribute(k, v);
  }
  for (const c of children) {
    if (c == null) continue;
    e.appendChild(typeof c === "string" ? document.createTextNode(c) : c);
  }
  return e;
}

// ─── Boot ───
async function boot() {
  const info = await wcall("go.main.App.GetAppInfo");
  if (info) {
    document.querySelector(".brand-sub").textContent = `${info.version} · ${info.platform}`;
    document.getElementById("offline-badge").textContent = info.offlineOnly ? "100% Offline" : "Online";
  }
  navigate("dashboard");
}

window.addEventListener("DOMContentLoaded", boot);
