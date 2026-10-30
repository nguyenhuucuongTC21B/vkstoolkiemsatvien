// app.js — main app logic.
// CRITICAL: Do NOT initialize `window.go = {}` here. Wails runtime + binding
// scripts (loaded in index.html BEFORE this file) populate `window.go.*`
// with the real bindings. If we set `window.go = {}` here, we would
// overwrite the real bindings and buttons would do nothing.

// Helper: call a Wails-bound method safely.
// Returns null on failure (binding missing or call threw).
async function wcall(path, ...args) {
  const parts = path.split(".");
  let obj = window;
  for (const p of parts) {
    obj = obj?.[p];
    if (obj == null) {
      console.error(`[wcall] binding missing: ${path} (failed at "${p}")`);
      return null;
    }
  }
  if (typeof obj !== "function") {
    console.error(`[wcall] not a function: ${path} (typeof=${typeof obj})`);
    return null;
  }
  try {
    return await obj(...args);
  } catch (e) {
    console.error(`[wcall] call threw: ${path}`, e);
    return null;
  }
}

// Check if Wails bindings are loaded.
// Wails runtime injects window.go.<package>.<Struct>.<Method> via SetBindings()
// when the page loads. For our app the paths are:
//   - go.main.App.GetAppInfo
//   - go.calc.Service.AnPhiHS / AnPhiDS / Lai / Tuoi / ThoiHan / ThoiHieuTNHS / BoiThuongNT / History
//   - go.legal.Service.Search / GetByID / Stats / ListBieuMau / ...
//   - go.ocr.Service.Recognize / IsAvailable / AvailableLanguages
//   - go.pii.Service.Mask / MaskFile
//   - go.spell.Service.Check / CheckFile
//   - go.file.Service.Apply
//   - go.evidence.Service.Add / ListByCase / TreeByCase / Update / Delete
//   - go.casefile.Service.Create / Get / List / Update / SetStatus / Delete / UpcomingDeadlines
//   - go.docs.Service.ListTemplates / Fill
function bindingsReady() {
  return !!(window.go && window.go.main && window.go.main.App &&
            typeof window.go.main.App.GetAppInfo === "function");
}

// Wait for bindings to be available (max 10 seconds).
function waitForBindings(timeoutMs = 10000) {
  return new Promise((resolve) => {
    if (bindingsReady()) return resolve(true);
    const start = Date.now();
    const interval = setInterval(() => {
      if (bindingsReady()) {
        clearInterval(interval);
        resolve(true);
      } else if (Date.now() - start > timeoutMs) {
        clearInterval(interval);
        resolve(false);
      }
    }, 100);
  });
}

// ─── Router ───
const routes = {};
function route(name, renderer) { routes[name] = renderer; }
async function navigate(name) {
  document.querySelectorAll(".nav-item").forEach(b =>
    b.classList.toggle("active", b.dataset.route === name));
  const r = routes[name];
  const page = document.getElementById("page");
  if (!r) {
    page.innerHTML = `<div class="card"><p>Chưa xây dựng module <code>${name}</code>.</p></div>`;
    return;
  }
  page.innerHTML = "<div class='card text-muted'>Đang tải...</div>";
  try { await r(page); }
  catch (e) { page.innerHTML = `<div class='callout callout-error'>Lỗi: ${e}</div>`; }
}

document.addEventListener("click", e => {
  const btn = e.target.closest(".nav-item");
  if (btn?.dataset.route) navigate(btn.dataset.route);
});

// ─── Helpers ───
function money(n) {
  if (n == null || isNaN(n)) return "—";
  return new Intl.NumberFormat("vi-VN").format(Math.round(n)) + "đ";
}

// Show an error banner in the page content when bindings are missing.
function showBindingsError(page) {
  // Diagnostic: list what bindings are present
  const diag = [];
  const checks = [
    ["window.go", window.go],
    ["window.go.main", window.go?.main],
    ["window.go.main.App", window.go?.main?.App],
    ["window.go.main.App.GetAppInfo", window.go?.main?.App?.GetAppInfo],
    ["window.go.calc", window.go?.calc],
    ["window.go.calc.Service", window.go?.calc?.Service],
    ["window.go.calc.Service.AnPhiHS", window.go?.calc?.Service?.AnPhiHS],
    ["window.go.legal", window.go?.legal],
    ["window.go.legal.Service.Search", window.go?.legal?.Service?.Search],
    ["window.go.ocr", window.go?.ocr],
    ["window.go.ocr.Service.Recognize", window.go?.ocr?.Service?.Recognize],
    ["window.go.pii", window.go?.pii],
    ["window.go.pii.Service.Mask", window.go?.pii?.Service?.Mask],
    ["window.go.spell", window.go?.spell],
    ["window.go.spell.Service.Check", window.go?.spell?.Service?.Check],
    ["window.go.file", window.go?.file],
    ["window.go.file.Service.Apply", window.go?.file?.Service?.Apply],
    ["window.go.evidence", window.go?.evidence],
    ["window.go.evidence.Service.Add", window.go?.evidence?.Service?.Add],
    ["window.go.casefile", window.go?.casefile],
    ["window.go.casefile.Service.Create", window.go?.casefile?.Service?.Create],
    ["window.go.docs", window.go?.docs],
    ["window.go.docs.Service.Fill", window.go?.docs?.Service?.Fill],
  ];
  for (const [name, val] of checks) {
    diag.push(`<tr><td><code>${name}</code></td><td>${val == null ? "❌ Missing" : "✓ Loaded"}</td></tr>`);
  }

  page.innerHTML = `
    <div class="callout callout-error">
      <h3 style="margin-bottom: 8px;">⚠ Lỗi: Wails bindings chưa load</h3>
      <p>JavaScript không tìm thấy <code>window.go.main.App</code> và các binding module.
         Các nút bấm trong phần mềm sẽ <b>không trả về kết quả</b>.</p>

      <p style="margin-top: 16px;"><b>🔍 Trạng thái binding:</b></p>
      <table class="tbl" style="font-size: 12px;">
        <thead><tr><th>Binding</th><th>Trạng thái</th></tr></thead>
        <tbody>${diag.join("")}</tbody>
      </table>

      <p style="margin-top: 16px;"><b>💡 Nguyên nhân & cách fix:</b></p>
      <ol style="margin: 8px 0 12px 24px; line-height: 1.8;">
        <li>Phần mềm được build bằng <code>go build</code> thay vì <code>wails build</code>
            — <b>đây là nguyên nhân phổ biến nhất</b>.<br>
            Fix: build lại bằng <code>wails build -platform windows/amd64</code></li>
        <li>File <code>frontend/wailsjs/go/main/App.js</code> không tồn tại trong source.
            Fix: chạy <code>wails dev</code> hoặc <code>wails build</code> để Wails sinh binding tự động.</li>
        <li>Wails runtime không load — mở DevTools (F12) xem console có lỗi gì không.</li>
        <li>Bạn mở file HTML trực tiếp bằng browser (file://) thay vì chạy qua phần mềm.
            Fix: chạy <code>./kiem_sat_toolkit.exe</code>, không mở HTML bằng browser.</li>
      </ol>

      <p style="margin-top: 12px;"><b>📞 Hỗ trợ:</b> Nguyễn Hữu Cường — 0905 6666 24 — doremon.dn01@gmail.com</p>
    </div>
  `;
}

// ─── Boot ───
async function boot() {
  // Wait for Wails bindings to be ready
  const ready = await waitForBindings(10000);
  const page = document.getElementById("page");

  if (!ready) {
    document.querySelector(".brand-sub").textContent = "⚠ Bindings not loaded";
    if (page) showBindingsError(page);
    // Try navigate anyway — it will use the error renderer above
    return;
  }

  const info = await wcall("go.main.App.GetAppInfo");
  if (info) {
    document.querySelector(".brand-sub").textContent =
      `${info.version} · ${info.platform}`;
    const badge = document.getElementById("offline-badge");
    if (badge) badge.textContent = info.offlineOnly ? "100% Offline" : "Online";

    // Update DB stats in sidebar footer
    const stats = await wcall("go.legal.Service.Stats");
    if (stats) {
      const total = Object.values(stats).reduce((a, b) => a + b, 0);
      const dbStats = document.getElementById("db-stats");
      if (dbStats) dbStats.textContent = `${total} văn bản pháp luật`;
    }
  }
  navigate("dashboard");
}

window.addEventListener("DOMContentLoaded", boot);
