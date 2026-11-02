# Models — LLM .gguf files

Đặt file model tại đây:

```
models/qwen3-0.6b-instruct-q4_k_m.gguf
```

Download từ HuggingFace:
```bash
wget https://huggingface.co/Qwen/Qwen3-0.6B-Instruct-GGUF/resolve/main/qwen3-0.6b-instruct-q4_k_m.gguf \
  -O models/qwen3-0.6b-instruct-q4_k_m.gguf
```

Spec:
- Context window: 2048
- Temperature: 0.0 - 0.5
- Size: ~400MB
- Format: GGUF Q4_K_M (quantized 4-bit)

Nếu không có file model, app vẫn chạy với STUB binding (mock responses).
Để dùng real model, build với: `go build -tags "llama,desktop,production"`
