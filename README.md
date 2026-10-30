# Công cụ ngành Kiểm sát — Phần mềm offline cho ngành Kiểm sát Việt Nam

Phần mềm hỗ trợ công tác Kiểm sát chạy **100% ngoại tuyến** dưới dạng
**single binary** Windows/Linux, **không cần cài đặt**, xây dựng bằng
**Go + Wails v2**.

## Tính năng MVP

### Module nghiệp vụ (đã implement)
- **Tính án phí hình sự** — NR 326/2016/UBTVQH14 điều 4, 5, 6, 9, 10
- **Tính án phí dân sự** — bậc thang 5/4/3/2% theo giá trị (điều 7 NR 326)
- **Tính tiền lãi** — trong hạn / quá hạn / chầm trả (BLDS 306)
- **Tính tuổi & thời hạn** tạm giữ (theo giờ) / tạm giam (theo tháng) — BLTTHS 128/2025
- **Tính thời hiệu TNHS** — BLHS điều 27 (5/10/15/20 năm theo loại tội)
- **Bồi thường thiệt hại ngoài hợp đồng** — BLDS 589-609
- **Tra cứu văn bản pháp luật** offline (13 văn bản + 6 biểu mẫu seed, có thể bổ sung)
- **OCR PDF/ảnh** — Tesseract vie/eng, không sửa file gốc
- **Che thông tin cá nhân (PII)** — 10 loại regex: CCCD, CMND, passport, phone, email, DOB, address, STK, MST, biển số xe
- **Rà soát chính tả** Vietnamese — telex typo, spacing, punct, term_norm
- **Đổi tên file hàng loạt** — number/replace/diacritics_strip/prefix/suffix/lowercase
- **Quản lý hồ sơ vụ án** + deadline tracking
- **Gắn nhãn chứng cứ** + cây nhãn phiên tòa
- **Soạn thảo biểu mẫu** với placeholder `{{...}}`
- **Audit log** — mọi thao tác ghi vào store, hỗ trợ trách nhiệm giải trình

### Văn bản pháp luật áp dụng
BLHS 2015 (sửa 2017), BLTTHS 2015 (sửa 128/2025), BLDS 2015, BLTTDS 2015 (sửa 85/2025),
NR 326/2016/UBTVQH14, Luật 106/2025 (THA DS), Luật 127/2025 (THA HS), NĐ 151-152/2026,
QĐ 39/QĐ-VKSTC 2018, QĐ 410/QĐ-VKSTC 2021, QĐ 50/QĐ-VKSTC 2026, NR 205/QH15.

## Kiến trúc

```
kiem_sat_toolkit/
├── main.go                          # entry point, Wails Run()
├── app.go                           # App binding (info, dataDir)
├── wails.json                       # Wails build config
├── go.mod / go.sum                  # deps: wails v2.10.1 (no SQLite — pure Go)
├── frontend/
│   └── dist/                        # static frontend (HTML/CSS/JS thuần)
│       ├── index.html               # SPA shell + sidebar nav
│       ├── styles.css               # design system (kỹ thuật trang trọng)
│       ├── wails.js                 # stub (real bindings auto-generated)
│       ├── app.js                   # router + helpers
│       └── pages.js                 # 12 route renderers
├── internal/
│   ├── audit/audit.go               # AuditLogger (writes to store)
│   ├── store/store.go               # JSON file store (no external DB)
│   ├── db/
│   │   ├── manager.go               # opens stores, seeds data
│   │   └── seed.go                  # 13 legal docs + 6 bieu_mau + 17 spell rules
│   └── modules/
│       ├── registry.go              # wires all modules
│       ├── calc/service.go          # án phí + lãi + tuổi + thời hạn + thời hiệu + bồi thường
│       ├── legal/service.go         # tra cứu văn bản + biểu mẫu
│       ├── ocr/service.go           # Tesseract wrapper (PDF→PNG via pdftoppm)
│       ├── pii/service.go           # 10-category regex PII masker
│       ├── spell/service.go         # Vietnamese spell check (telex + rules)
│       ├── file/service.go          # bulk rename
│       ├── evidence/service.go      # labeled evidence tree
│       ├── casefile/service.go      # case CRUD + deadline tracking
│       └── docs/service.go          # template fill + export
├── build/
│   ├── scripts/build.sh             # Linux + Windows cross-build
│   └── bin/                         # output binaries
```

## Build

### Yêu cầu deps hệ thống (chỉ khi build Linux native)
```bash
# Debian/Ubuntu:
sudo apt install libwebkit2gtk-4.1-dev libgtk-3-dev libsoup-3.0-dev \
  libjavascriptcoregtk-4.1-dev build-essential pkg-config
# Optional (cho OCR PDF):
sudo apt install tesseract-ocr tesseract-ocr-vie poppler-utils
```

### Linux (host)
```bash
./build/scripts/build.sh linux
# → build/bin/kiem_sat_toolkit_linux (single binary ~5MB)
```

### Windows (cross-build từ Linux)
```bash
sudo apt install gcc-mingw-w64-x86-64
./build/scripts/build.sh windows
# → build/bin/kiem_sat_toolkit.exe (single .exe, ~8-10MB)
# Sử dụng WebView2 runtime (pre-installed trên Windows 10/11)
```

### Build bằng Wails CLI (đầy đủ tính năng, khuyến nghị)
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails build -platform linux/amd64,windows/amd64
```

### Build trực tiếp `go build`
```bash
go build -tags "desktop,production" -ldflags "-s -w" -o build/bin/kiem_sat_toolkit .
```
⚠️ **Bắt buộc có `-tags "desktop,production"`** — nếu không có sẽ gặp lỗi
"Wails applications will not build without the correct build tags."

## Chạy thử (dev mode)

```bash
wails dev
# Mở cửa sổ desktop app + hot reload
```

## Tests

```bash
go test ./internal/modules/calc/ -v
# 11 tests PASS — verify án phí HS/DS, lãi, tuổi, thời hạn, thời hiệu, bồi thường
```

## Offline-first design

- **Không có deps ngoài**: SQLite đã được thay bằng JSON file store thuần Go — không cần libsqlite3 hệ thống.
- **Không gọi mạng**: mọi thao tác chạy local. Audit log ghi rõ mọi lần tính/tra cứu/OCR/che PII.
- **Portable data dir**: ưu tiên `./data` cùng thư mục binary (USB-stick friendly), fallback `~/.kiem_sat_toolkit`.
- **Single binary**: Go embed tất cả frontend (HTML/CSS/JS) + legal seed data vào binary.
- **Cross-platform**: cùng codebase, build cho Windows + Linux từ một máy.

## Bổ sung dữ liệu pháp luật

Sửa `internal/db/seed.go` để thêm văn bản mới vào `legalDocsSeed` slice, hoặc
viết tool import từ CSDL vbpl.vn → JSON snapshot → embed qua `go:embed`.

## Roadmap (chưa implement trong MVP)

- AI offline (LLM local — Qwen/Llama.cpp) để kiểm sát bản án tự động
- Đồng bộ nội bộ LAN giữa các máy VKS
- Ký số file xuất (PKI/VNPT-CA)
- Module khởi kiện công ích theo NR 205
- Module nghiệp vụ văn phòng (156 KPI, 9 mảng)
- Import DB pháp luật đầy đủ (hàng nghìn văn bản) qua go:embed JSON
- OpenVINO integration cho OCR/LLM inference (tăng tốc 4-10× trên CPU/NPU)

## Liên hệ / đóng góp

Báo lỗi, góp ý qua issue trên repo hoặc email. Mọi đóng góp tuân thủ
các quy định nghiệp vụ ngành Kiểm sát.
