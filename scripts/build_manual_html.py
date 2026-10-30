#!/usr/bin/env python3
"""
Convert HUONG_DAN_SU_DUNG.md → polished standalone HTML user manual
using pandoc for robust markdown parsing.
- Cover page with VKS logo + coder info
- Sidebar TOC (clickable, sticky)
- Search box
- Print-friendly CSS
- VKS brand colors
"""
import base64
import html
import re
import subprocess
from pathlib import Path

ROOT = Path("/home/z/my-project/download/kiem_sat_toolkit")
MD = ROOT / "HUONG_DAN_SU_DUNG.md"
LOGO = ROOT / "assets/logo.png"
OUT = ROOT / "HUONG_DAN_SU_DUNG.html"

md_text = MD.read_text(encoding="utf-8")
logo_b64 = base64.b64encode(LOGO.read_bytes()).decode()

# Use pandoc to convert markdown → HTML body
result = subprocess.run(
    ["pandoc", "--from=markdown", "--to=html5",
     "--no-highlight", str(MD)],
    capture_output=True, text=True, check=True
)
body_html = result.stdout

# Build TOC from h2 headings
toc_entries = []
for m in re.finditer(r'<h2[^>]*id="([^"]+)"[^>]*>(.+?)</h2>', body_html):
    eid = m.group(1)
    text = re.sub(r"<[^>]+>", "", m.group(2))
    toc_entries.append((eid, text))

toc_html = "\n".join(
    f'<a href="#{eid}" class="toc-item">{html.escape(text)}</a>'
    for eid, text in toc_entries
)

HTML = f"""<!DOCTYPE html>
<html lang="vi">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Hướng dẫn sử dụng — Công cụ ngành Kiểm sát v0.1.0</title>
<meta name="author" content="Nguyễn Hữu Cường — CV CY KTCC">
<meta name="description" content="Hướng dẫn sử dụng phần mềm Công cụ ngành Kiểm sát v0.1.0 — offline, Go + Wails v2">
<style>
* {{ box-sizing: border-box; margin: 0; padding: 0; }}
html, body {{ height: 100%; }}
body {{
  font-family: 'Segoe UI', -apple-system, 'Helvetica Neue', Arial, sans-serif;
  background: #f4f4f5;
  color: #212425;
  font-size: 15px;
  line-height: 1.65;
  -webkit-font-smoothing: antialiased;
}}
.app {{ display: flex; min-height: 100vh; }}
.sidebar {{
  width: 280px;
  flex-shrink: 0;
  background: #ffffff;
  border-right: 1px solid #e9ebec;
  position: sticky;
  top: 0;
  height: 100vh;
  overflow-y: auto;
  padding: 24px 16px;
}}
.sidebar-brand {{
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e9ebec;
  margin-bottom: 16px;
}}
.sidebar-brand img {{ width: 48px; height: 48px; border-radius: 6px; background: #000; }}
.sidebar-brand .title {{ font-size: 14px; font-weight: 700; color: #212425; }}
.sidebar-brand .sub {{ font-size: 11px; color: #7d8487; margin-top: 2px; }}
.search-box {{
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #b8c0c7;
  border-radius: 4px;
  font-size: 13px;
  margin-bottom: 16px;
  font-family: inherit;
}}
.search-box:focus {{ outline: 2px solid #3586ae; border-color: #3586ae; }}
.toc-item {{
  display: block;
  padding: 6px 10px;
  font-size: 13px;
  color: #3e525c;
  text-decoration: none;
  border-radius: 4px;
  font-weight: 500;
  transition: background .12s;
}}
.toc-item:hover {{ background: #e9ebec; }}
.toc-item.active {{ background: #3e525c; color: white; }}
.sidebar-foot {{
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #e9ebec;
  font-size: 11px;
  color: #7d8487;
}}
.sidebar-foot a {{ color: #3586ae; text-decoration: none; }}
.sidebar-foot a:hover {{ text-decoration: underline; }}
.main {{ flex: 1; padding: 0; min-width: 0; }}
.content {{
  max-width: 880px;
  margin: 0 auto;
  padding: 48px 56px 96px;
}}
.cover {{
  background: linear-gradient(135deg, #3e525c 0%, #3586ae 100%);
  color: white;
  padding: 80px 56px;
  text-align: center;
  margin: -48px -56px 48px;
}}
.cover img.logo {{
  width: 96px;
  height: 96px;
  border-radius: 12px;
  background: rgba(0,0,0,0.5);
  margin-bottom: 24px;
}}
.cover h1 {{
  font-size: 36px;
  font-weight: 800;
  margin-bottom: 12px;
  letter-spacing: -0.5px;
  border: none;
  color: white;
}}
.cover .sub {{
  font-size: 18px;
  opacity: 0.9;
  margin-bottom: 32px;
}}
.cover .meta {{
  display: inline-flex;
  gap: 24px;
  font-size: 13px;
  opacity: 0.85;
  flex-wrap: wrap;
  justify-content: center;
}}
.cover .coder {{
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid rgba(255,255,255,0.2);
  font-size: 12px;
  opacity: 0.85;
}}
h1 {{ font-size: 32px; font-weight: 700; margin: 48px 0 16px; color: #212425; border-bottom: 2px solid #3586ae; padding-bottom: 8px; }}
h2 {{ font-size: 24px; font-weight: 700; margin: 40px 0 16px; color: #3e525c; }}
h3 {{ font-size: 18px; font-weight: 600; margin: 28px 0 12px; color: #3e525c; }}
h4 {{ font-size: 16px; font-weight: 600; margin: 20px 0 8px; color: #4a5258; }}
p {{ margin: 0 0 14px; }}
a {{ color: #3586ae; text-decoration: none; }}
a:hover {{ text-decoration: underline; }}
strong {{ color: #212425; font-weight: 700; }}
em {{ color: #4a5258; }}
code {{
  background: #e9ebec;
  color: #3e525c;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
}}
pre {{
  background: #1e2528;
  color: #d4d4d4;
  padding: 16px 20px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 16px 0;
  font-size: 13px;
  line-height: 1.5;
}}
pre code {{ background: transparent; color: inherit; padding: 0; }}
hr {{ border: none; border-top: 1px solid #e9ebec; margin: 32px 0; }}
ul, ol {{ margin: 0 0 14px; padding-left: 24px; }}
li {{ margin-bottom: 6px; }}
li > ul, li > ol {{ margin-top: 6px; }}
table {{
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
  font-size: 14px;
  background: white;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}}
th {{
  background: #3e525c;
  color: white;
  padding: 10px 14px;
  text-align: left;
  font-weight: 600;
}}
td {{
  padding: 10px 14px;
  border-bottom: 1px solid #e9ebec;
}}
tr:nth-child(even) td {{ background: #f4f5f6; }}
tr:hover td {{ background: #e9ebec; }}
td code {{ background: #e9ebec; padding: 2px 6px; border-radius: 3px; }}
.back-top {{
  position: fixed;
  bottom: 24px;
  right: 24px;
  background: #3e525c;
  color: white;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  text-decoration: none;
  font-size: 18px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
  opacity: 0;
  transition: opacity .2s;
  z-index: 100;
}}
.back-top.visible {{ opacity: 1; }}
.back-top:hover {{ background: #3586ae; }}
@media print {{
  .sidebar, .back-top {{ display: none !important; }}
  .main {{ width: 100%; }}
  .content {{ max-width: none; padding: 0; }}
  .cover {{ margin: 0 0 24px; padding: 32px; page-break-after: always; }}
  body {{ font-size: 12px; }}
  h1, h2, h3 {{ page-break-after: avoid; }}
  pre, table {{ page-break-inside: avoid; }}
  a {{ color: #212425; text-decoration: underline; }}
}}
@media (max-width: 900px) {{
  .app {{ flex-direction: column; }}
  .sidebar {{ position: relative; width: 100%; height: auto; max-height: 240px; }}
  .content {{ padding: 24px; }}
  .cover {{ margin: -24px -24px 24px; padding: 40px 24px; }}
}}
</style>
</head>
<body>
<div class="app">
  <aside class="sidebar">
    <div class="sidebar-brand">
      <img src="data:image/png;base64,{logo_b64}" alt="VKS">
      <div>
        <div class="title">Công cụ KS</div>
        <div class="sub">Hướng dẫn sử dụng v0.1.0</div>
      </div>
    </div>
    <input type="text" class="search-box" id="search" placeholder="🔎 Tìm trong hướng dẫn...">
    <nav id="toc">
      {toc_html}
    </nav>
    <div class="sidebar-foot">
      <div><b>© 2026 Nguyễn Hữu Cường</b></div>
      <div>CV CY KTCC</div>
      <div>📞 <a href="tel:0905666624">0905 6666 24</a></div>
      <div>✉ <a href="mailto:doremon.dn01@gmail.com">doremon.dn01@gmail.com</a></div>
    </div>
  </aside>
  <main class="main">
    <div class="content">
      <div class="cover">
        <img src="data:image/png;base64,{logo_b64}" alt="VKS Logo" class="logo">
        <h1>Hướng dẫn sử dụng</h1>
        <div class="sub">Phần mềm Công cụ ngành Kiểm sát v0.1.0</div>
        <div class="meta">
          <span>📦 Single binary</span>
          <span>🔒 100% Offline</span>
          <span>🪟 Windows + Linux</span>
          <span>⚙️ Go + Wails v2</span>
        </div>
        <div class="coder">
          Coder Developed by <b>Nguyễn Hữu Cường</b> — CV CY KTCC<br>
          📞 0905 6666 24 · ✉ doremon.dn01@gmail.com
        </div>
      </div>
      {body_html}
    </div>
  </main>
</div>
<a href="#" class="back-top" id="back-top" title="Lên đầu trang">↑</a>
<script>
const tocItems = document.querySelectorAll('.toc-item');
const sections = [];
tocItems.forEach(a => {{
  const id = a.getAttribute('href').slice(1);
  const el = document.getElementById(id);
  if (el) sections.push({{el, link: a}});
}});
function updateActive() {{
  let active = null;
  const scrollY = window.scrollY + 100;
  for (const s of sections) {{
    if (s.el.offsetTop <= scrollY) active = s;
  }}
  tocItems.forEach(a => a.classList.remove('active'));
  if (active) active.link.classList.add('active');
}}
window.addEventListener('scroll', updateActive);
updateActive();
const backTop = document.getElementById('back-top');
window.addEventListener('scroll', () => {{
  backTop.classList.toggle('visible', window.scrollY > 400);
}});
backTop.addEventListener('click', e => {{
  e.preventDefault();
  window.scrollTo({{top: 0, behavior: 'smooth'}});
}});
const searchBox = document.getElementById('search');
searchBox.addEventListener('input', e => {{
  const q = e.target.value.toLowerCase().trim();
  tocItems.forEach(a => {{
    if (!q) {{ a.style.display = ''; return; }}
    const text = a.textContent.toLowerCase();
    a.style.display = text.includes(q) ? '' : 'none';
  }});
}});
tocItems.forEach(a => {{
  a.addEventListener('click', e => {{
    e.preventDefault();
    const id = a.getAttribute('href').slice(1);
    const el = document.getElementById(id);
    if (el) window.scrollTo({{top: el.offsetTop - 20, behavior: 'smooth'}});
  }});
}});
</script>
</body>
</html>
"""

OUT.write_text(HTML, encoding="utf-8")
print(f"✓ Generated: {OUT}")
print(f"  Size: {OUT.stat().st_size / 1024:.1f} KB")
print(f"  TOC entries: {len(toc_entries)}")
