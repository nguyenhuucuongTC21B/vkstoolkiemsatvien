# AI Module — Build với Real LLM (llama.cpp)

Phần mềm v1.0 có 2 chế độ build:

## 1. Default (STUB) — không cần libllama

```bash
go build -tags "desktop,production" -o ks_tools .
# hoặc qua wails:
wails build -platform linux/amd64
```

Stub trả về mock JSON hợp lệ cho 5 module. Dùng cho:
- Dev/test UI mà không cần GPU/CPU mạnh
- Demo trên máy không có model .gguf
- CI/CD chạy tests

## 2. Real LLM — cần libllama + model file

### Bước 1: Build llama.cpp

```bash
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make  # Linux/macOS
# Hoặc Windows: cmake -B build && cmake --build build --config Release
```

Output: `libllama.a` (Linux) hoặc `llama.dll` (Windows).

### Bước 2: Install Go binding

```bash
go get github.com/go-skynet/go-llama.cpp
# Binding tự build libllama.a via CMake (CGO_ENABLED=1)
```

### Bước 3: Download model

```bash
mkdir -p models
# Tải qwen3-0.6b-instruct-q4_k_m.gguf từ HuggingFace
wget https://huggingface.co/Qwen/Qwen3-0.6B-Instruct-GGUF/resolve/main/qwen3-0.6b-instruct-q4_k_m.gguf \
  -O models/qwen3-0.6b-instruct-q4_k_m.gguf
```

Spec:
- Context window: 2048 tokens
- Temperature: 0.0 - 0.5
- Format: GGUF Q4_K_M (quantized 4-bit)
- Size: ~400MB

### Bước 4: Build với tag `llama`

```bash
CGO_ENABLED=1 go build -tags "llama,desktop,production" \
  -ldflags "-s -w" -o ks_tools .

# Hoặc qua Wails:
wails build -platform linux/amd64 -tags "llama"
```

### Bước 5: Run

```bash
./ks_tools
# App start → load model (400MB RAM) → workers spawn → ready for jobs
# Shutdown → defer model.Free() giải phóng RAM
```

## Architecture — Worker Pool

```
                    ┌─────────────────────────────┐
                    │   Main Thread (UI/60FPS)    │
                    │   - Wails window            │
                    │   - JS setInterval 16ms     │
                    │     ↳ PollResults()         │
                    │     ↳ GetLog()              │
                    └──────────┬──────────────────┘
                               │
                    Submit(job)│
                               ▼
                    ┌─────────────────────────────┐
                    │   jobs chan Job (buffered 16)│
                    └──────────┬──────────────────┘
                               │
                  ┌────────────┴────────────┐
                  ▼                         ▼
        ┌──────────────────┐      ┌──────────────────┐
        │ Worker-0 goroutine│      │ Worker-1 goroutine│
        │ (model shared)    │      │ (model shared)    │
        │                   │      │                   │
        │  dispatchModule() │      │  dispatchModule() │
        │   ├ extractor     │      │   ├ extractor     │
        │   ├ reporter      │      │   ├ reporter      │
        │   ├ chatter       │      │   ├ chatter       │
        │   ├ auditor       │      │   ├ auditor       │
        │   └ advisor       │      │   └ advisor       │
        └────────┬──────────┘      └────────┬──────────┘
                 │                          │
                 ▼                          ▼
        ┌─────────────────────────────────────────┐
        │      results chan Result (buffered 16)   │
        └────────────────────┬────────────────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  UI drain loop   │
                    │  range results   │
                    └──────────────────┘
```

## Memory Management

- Model load **1 lần duy nhất** khi app start (`WorkerPool.Start()`)
- Workers share cùng `model Model` instance (thread-safe trong go-llama.cpp)
- `defer model.Free()` gọi trong `WorkerPool.Stop()` khi app shutdown
- Nếu app crash, OS tự reclaim RAM (no leak risk)

## 5 Module AI

| Module | Input | Output |
|---|---|---|
| 1. EXTRACTOR | `text` (văn bản OCR) | JSON `{nhan_than, timeline, vat_chung, tom_tat_vu_an}` |
| 2. REPORTER | `case_info` hoặc `extractor_output` | Báo cáo 3 phần (Tổng quan, Điểm nhấn, Đánh giá rủi ro) |
| 3. CHATTER | `documents` + `question` | Câu trả lời ngắn + chỉ đoạn trong bút lục |
| 4. AUDITOR | `statement_a` + `statement_b` | `MAU THUẬN:\n1. ...\n2. ...` hoặc `NHẤT QUÁN` |
| 5. ADVISOR | `behavior` | `PHÂN TÍCH PHÁP LÝ: 1. Hành vi 2. Điều BLHS 3. Khung phạt 4. Đề xuất 5. Án phí` |

## laws.json (Module 5)

File `internal/ai/data/laws.json` embed vào binary:
- BLHS 2015 (sửa 2017)
- 5 chương mẫu: An ninh quốc gia, Xâm phạm tính mạng/sức khỏe, Xâm phạm sở hữu, Ma túy
- 6 điều tiêu biểu: 108 (gián điệp), 123 (giết người), 134 (cố ý gây thương tích), 139 (cướp), 173 (trộm cắp), 247 (sản xuất ma túy)
- Mỗi điều có 2-5 khoản với khung hình phạt cụ thể

## Troubleshooting

### Lỗi: "model file not found"
- File `models/qwen3-0.6b-instruct-q4_k_m.gguf` không có
- Fix: download từ HuggingFace (xem Bước 3)

### Lỗi: build fail với tag llama
- libllama.a chưa build hoặc không trong LD_LIBRARY_PATH
- Fix: `cd llama.cpp && make`, rồi set `CGO_LDFLAGS="-L/path/to/llama.cpp"`

### Lỗi: inference chậm
- Model 0.6B Q4_K_M cần ~1GB RAM, inference ~1-3s/prompt trên CPU 4-core
- Nếu cần nhanh hơn: dùng model nhỏ hơn (Q2_K) hoặc GPU build của llama.cpp

### Lỗi: output JSON không hợp lệ (Module 1)
- Model nhỏ (0.6B) đôi khi trả JSON sai format
- Stub implementation đã clean output; real binding cần post-process
- Code đã có `cleanJSONOutput()` + `isJSON()` check
