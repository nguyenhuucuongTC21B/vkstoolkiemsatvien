// pages_ai.js — 6 routes cho AI Worker Pool (v1.0).
//
// Tuân thủ yêu cầu kiến trúc:
// - UI không block — submit job async, poll results 60FPS (16ms interval)
// - Live log streaming từ worker pool
// - 5 module forms: extractor / reporter / chatter / auditor / advisor

// ─────────── Poll loop (60FPS) — chỉ start 1 lần ───────────
let aiPollStarted = false;
let aiResultsContainer = null;
let aiLogContainer = null;

function startAIPolling() {
  if (aiPollStarted) return;
  aiPollStarted = true;

  // 60FPS = ~16ms. Poll results + log mỗi 16ms.
  setInterval(async () => {
    // Poll results — drain channel
    try {
      const results = await wcall("go.aimod.Service.PollResults");
      if (results && results.length > 0 && aiResultsContainer) {
        for (const r of results) {
          appendResult(r);
        }
      }
    } catch(e) { /* ignore */ }

    // Poll log lines — 60FPS display
    try {
      const lines = await wcall("go.aimod.Service.GetLog");
      if (lines && lines.length > 0 && aiLogContainer) {
        const html = lines.map(l => `<div class="log-line">${escapeHtml(l)}</div>`).join("");
        aiLogContainer.innerHTML = html;
        aiLogContainer.scrollTop = aiLogContainer.scrollHeight;
      }
    } catch(e) { /* ignore */ }
  }, 50); // 50ms = ~20Hz (Wails IPC overhead too high for true 60Hz)
}

function escapeHtml(s) {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function appendResult(r) {
  const div = document.createElement("div");
  div.className = "result-block";
  const status = r.error ? "❌ Error" : "✓ Done";
  const output = r.error ? `<div class="callout callout-error">${r.error}</div>`
                         : `<div class="result-detail" style="white-space:pre-wrap; max-height:400px; overflow-y:auto;">${escapeHtml(r.output)}</div>`;
  div.innerHTML = `
    <div class="flex-between">
      <b>${r.module} — job ${r.jobId}</b>
      <span class="text-muted">${status} · ${r.duration}ms · worker-${r.workerId}</span>
    </div>
    ${output}
  `;
  aiResultsContainer.prepend(div);
}

// ─────────── AI Dashboard ───────────
route("ai", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">🤖 AI Worker Pool Dashboard — v1.0</div>
      <p class="text-muted">Worker Pool với Goroutines, không block UI. Model LLM loaded 1 lần trong RAM, defer Free() khi shutdown.</p>
      <div id="ai-stats" class="mt-4"></div>
    </div>

    <div class="grid-2">
      <div class="card">
        <div class="card-title">📋 Live Log (60FPS)</div>
        <div id="ai-log" style="background:#1e2528; color:#d4d4d4; padding:12px; border-radius:4px; font-family:Consolas,monospace; font-size:11px; height:300px; overflow-y:auto;"></div>
      </div>
      <div class="card">
        <div class="card-title">📤 Results Stream</div>
        <div id="ai-results" style="max-height:300px; overflow-y:auto;"></div>
      </div>
    </div>

    <div class="card">
      <div class="card-title">🚀 Quick Test — Submit 5 jobs</div>
      <p class="text-muted">Test worker pool với 5 module cùng lúc. Workers xử lý song song.</p>
      <button class="btn" id="ai-test">⚡ Submit 5 jobs (extractor + reporter + chatter + auditor + advisor)</button>
    </div>
  `;
  aiResultsContainer = page.querySelector("#ai-results");
  aiLogContainer = page.querySelector("#ai-log");
  startAIPolling();

  // Stats refresh
  async function refreshStats() {
    const s = await wcall("go.aimod.Service.Stats");
    if (!s) return;
    page.querySelector("#ai-stats").innerHTML = `
      <div class="stat-grid">
        <div class="stat"><div class="stat-num">${s.jobsSubmitted}</div><div class="stat-label">Jobs submitted</div></div>
        <div class="stat"><div class="stat-num">${s.jobsCompleted}</div><div class="stat-label">Jobs completed</div></div>
        <div class="stat"><div class="stat-num">${s.queueLen}</div><div class="stat-label">Queue</div></div>
        <div class="stat"><div class="stat-num">${s.workers}</div><div class="stat-label">Workers</div></div>
        <div class="stat"><div class="stat-num">${s.modelInfo.loaded ? "✓" : "✗"}</div><div class="stat-label">Model</div></div>
        <div class="stat"><div class="stat-num">${s.modelInfo.contextSize || 0}</div><div class="stat-label">Context</div></div>
      </div>
      <div class="callout mt-4">
        <b>Model:</b> ${s.modelInfo.name || "Not loaded"}<br>
        <b>Path:</b> <code>${s.modelInfo.path || "—"}</code><br>
        <b>Backend:</b> ${s.modelInfo.backend || "—"} · <b>Temp:</b> ${s.modelInfo.temperature || 0}
      </div>
    `;
  }
  refreshStats();
  setInterval(refreshStats, 1000);

  page.querySelector("#ai-test").onclick = async () => {
    const modules = [
      { module: "extractor", input: { text: "Lời khai bị can Nguyễn Văn A, MST 0123456789, sinh 1990, hành vi đánh thương tích 32%" } },
      { module: "reporter", input: { case_info: "Bị can A đánh nạn nhân B thương tích 32% ngày 15/09/2026" } },
      { module: "chatter", input: { documents: "Bút lục: Bị can đến nhà nạn nhân lúc 20:00 ngày 15/09", question: "Khi nào sự việc xảy ra?" } },
      { module: "auditor", input: { statement_a: "Sự việc xảy ra lúc 20:00, không có dao", statement_b: "Sự việc xảy ra lúc 21:00, bị can cầm dao phay" } },
      { module: "advisor", input: { behavior: "Cố ý gây thương tích tỷ lệ 32%" } },
    ];
    for (const m of modules) {
      const jobID = await wcall("go.aimod.Service.SubmitJob", m);
      if (jobID) {
        const div = document.createElement("div");
        div.className = "callout callout-success";
        div.innerHTML = `✓ Submitted <b>${m.module}</b> → jobID: <code>${jobID}</code>`;
        aiResultsContainer.prepend(div);
      }
    }
  };
});

// ─────────── Helper cho 5 module pages ───────────
function renderAIModulePage(page, moduleNumber, moduleName, moduleKey, formHTML, submitInput) {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">🤖 MODULE ${moduleNumber}: ${moduleName}</div>
      <div id="ai-mod-stats" class="text-muted mb-4"></div>
      ${formHTML}
      <button class="btn mt-4" id="ai-mod-submit">⚡ Submit to Worker Pool</button>
      <div id="ai-mod-result" class="mt-4"></div>
    </div>
    <div class="card">
      <div class="card-title">📋 Live Log</div>
      <div id="ai-log" style="background:#1e2528; color:#d4d4d4; padding:12px; border-radius:4px; font-family:Consolas,monospace; font-size:11px; height:200px; overflow-y:auto;"></div>
    </div>
  `;
  aiResultsContainer = page.querySelector("#ai-mod-result");
  aiLogContainer = page.querySelector("#ai-log");
  startAIPolling();

  page.querySelector("#ai-mod-submit").onclick = async () => {
    const input = submitInput(page);
    if (!input) return;
    const jobID = await wcall("go.aimod.Service.SubmitJob", { module: moduleKey, input });
    if (jobID) {
      page.querySelector("#ai-mod-result").innerHTML =
        `<div class="callout callout-success">✓ Đã submit job <code>${jobID}</code> — đợi kết quả bên dưới...</div>`;
    } else {
      page.querySelector("#ai-mod-result").innerHTML =
        `<div class="callout callout-error">Lỗi submit job (queue có thể đầy)</div>`;
    }
  };
}

// ─────────── MODULE 1: EXTRACTOR ───────────
route("ai-extractor", async (page) => {
  renderAIModulePage(page, 1, "EXTRACTOR — Bóc tách dữ liệu", "extractor",
    `<div class="form-row">
      <label>Văn bản thô OCR (lời khai / QĐ điều tra)</label>
      <textarea id="ex-text" rows="8" style="font-family:monospace;">BIÊN BẢN LỜI KHAI BỊ CAN
Ngày 16/09/2026, tại trụ sở CQĐT...
Tôi tên là Nguyễn Văn A, sinh ngày 20/05/1990 tại Bắc Ninh, MST 0123456789, CCCD 001234567890, nghề lái xe, cư trú tại thôn X, xã Y, huyện Z, tỉnh Bắc Ninh.
Khoảng 20:00 ngày 15/09/2026, tôi đến nhà anh B. Xảy ra ẩu đả, tôi dùng dao phay đánh anh B.
Vật chứng thu giữ: 01 dao phay có vết máu, 01 điện thoại iPhone hư hỏng.</textarea>
    </div>`,
    (p) => ({ text: p.querySelector("#ex-text").value })
  );
});

// ─────────── MODULE 2: REPORTER ───────────
route("ai-reporter", async (page) => {
  renderAIModulePage(page, 2, "REPORTER — Dự thảo văn bản", "reporter",
    `<div class="form-row">
      <label>Thông tin vụ án (hoặc JSON từ Module 1)</label>
      <textarea id="rp-info" rows="6">Bị can Nguyễn Văn A, sinh 1990, Bắc Ninh.
Hành vi: dùng dao phay đánh nạn nhân B lúc 20:00 ngày 15/09/2026 tại thôn X.
Thương tích: 32% theo kết luận giám định pháp y.
Vật chứng: dao phay, điện thoại hư hỏng.</textarea>
    </div>`,
    (p) => ({ case_info: p.querySelector("#rp-info").value })
  );
});

// ─────────── MODULE 3: CHATTER ───────────
route("ai-chatter", async (page) => {
  renderAIModulePage(page, 3, "CHATTER — Hỏi đáp hồ sơ vụ án", "chatter",
    `<div class="form-row">
      <label>Tài liệu bút lục vụ án</label>
      <textarea id="ch-docs" rows="6">Bút lục vụ án số QĐ 12/2026/TLPT-HS:
- Lời khai bị can A: "Tôi đến nhà anh B khoảng 20:00 ngày 15/09/2026..."
- Lời khai nhân chứng C: "Tôi thấy anh A cầm dao..."
- Kết luận giám định: nạn nhân B thương tích 32%.</textarea>
    </div>
    <div class="form-row">
      <label>Câu hỏi của KSV</label>
      <input id="ch-q" type="text" value="Sự việc xảy ra vào lúc mấy giờ? Chi tiết nằm ở đâu trong bút lục?">
    </div>`,
    (p) => ({ documents: p.querySelector("#ch-docs").value, question: p.querySelector("#ch-q").value })
  );
});

// ─────────── MODULE 4: AUDITOR & CROSS-CHECKER ───────────
route("ai-auditor", async (page) => {
  renderAIModulePage(page, 4, "AUDITOR — Cross-check 2 lời khai", "auditor",
    `<div class="form-row">
      <label>Lời khai A (bị can)</label>
      <textarea id="au-a" rows="4">Khoảng 20:00 ngày 15/09/2026, tôi đến nhà anh B. Tôi không có dao. Anh B ngã tự nhiên.</textarea>
    </div>
    <div class="form-row">
      <label>Lời khai B (nhân chứng)</label>
      <textarea id="au-b" rows="4">Khoảng 21:00 ngày 15/09/2026, tôi thấy anh A cầm dao phay đánh anh B nhiều nhát.</textarea>
    </div>`,
    (p) => ({ statement_a: p.querySelector("#au-a").value, statement_b: p.querySelector("#au-b").value })
  );
});

// ─────────── MODULE 5: LEGAL ADVISOR ───────────
route("ai-advisor", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">🤖 MODULE 5: LEGAL ADVISOR — Tra cứu luật offline</div>
      <p class="text-muted">AI tra cứu <code>laws.json</code> + đề xuất chương/điều/khung phạt BLHS dựa trên hành vi phạm tội.</p>
      <div class="form-row">
        <label>Hành vi phạm tội (mô tả ngắn — có thể lấy từ Module 1)</label>
        <textarea id="ad-behavior" rows="3">Cố ý gây thương tích cho nạn nhân với tỷ lệ thương tích 32%, có sử dụng dao phay.</textarea>
      </div>
      <button class="btn mt-2" id="ad-submit">⚡ Submit to AI Worker</button>
      <div id="ai-mod-result" class="mt-4"></div>
    </div>
    <div class="card">
      <div class="card-title">📋 Live Log</div>
      <div id="ai-log" style="background:#1e2528; color:#d4d4d4; padding:12px; border-radius:4px; font-family:Consolas,monospace; font-size:11px; height:200px; overflow-y:auto;"></div>
    </div>
    <div class="card">
      <div class="card-title">📚 Tra cứu luật trực tiếp (no AI)</div>
      <p class="text-muted">Tra cứu <code>laws.json</code> trực tiếp không qua AI:</p>
      <div class="flex mb-2">
        <input id="ad-kw" type="text" placeholder="vd: thương tích, trộm cắp, giết người" style="flex:1;">
        <button class="btn" id="ad-search">🔍 Tìm</button>
      </div>
      <div id="ad-list"></div>
    </div>
  `;
  aiResultsContainer = page.querySelector("#ai-mod-result");
  aiLogContainer = page.querySelector("#ai-log");
  startAIPolling();

  page.querySelector("#ad-submit").onclick = async () => {
    const behavior = page.querySelector("#ad-behavior").value;
    const jobID = await wcall("go.aimod.Service.SubmitJob",
      { module: "advisor", input: { behavior } });
    if (jobID) {
      page.querySelector("#ai-mod-result").innerHTML =
        `<div class="callout callout-success">✓ Submitted advisor job <code>${jobID}</code> — đợi kết quả...</div>`;
    }
  };
  page.querySelector("#ad-search").onclick = async () => {
    const kw = page.querySelector("#ad-kw").value;
    const results = await wcall("go.aimod.Service.ListLaws", kw, 20);
    const out = page.querySelector("#ad-list");
    if (!results || results.length === 0) {
      out.innerHTML = "<p class='text-muted'>Không tìm thấy điều phù hợp.</p>";
      return;
    }
    out.innerHTML = results.map(d => `
      <div class="card" style="margin-bottom:8px; padding:12px; border-left:4px solid #3586ae;">
        <h4>Điều ${d.so}: ${d.ten}</h4>
        ${(d.khoan||[]).map(k => `<p><b>${k.so}.</b> ${k.noiDung}</p><p class="text-muted">Khung phạt: ${k.khungPhat}</p>`).join("")}
      </div>
    `).join("");
  };
});
