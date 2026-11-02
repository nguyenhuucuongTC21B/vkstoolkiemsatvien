// binding_llama.go — REAL LLM implementation using github.com/go-skynet/go-llama.cpp.
//
// Build với tag "llama" để enable:
//   go build -tags "llama,desktop,production" -o ks_tools .
//
// Yêu cầu hệ thống:
//   - libllama.a build từ llama.cpp (xem scripts/build-libllama.sh)
//   - Model file: ./models/qwen3-0.6b-instruct-q4_k_m.gguf
//   - CGO_ENABLED=1, gcc/clang, libstdc++
//
//go:build llama

package ai

import (
	"fmt"
	"os"

	llama "github.com/go-skynet/go-llama.cpp"
)

// llamaModel — wrapper thật cho llama.cpp Go binding.
type llamaModel struct {
	cfg   Config
	model *llama.LLama
}

// NewModel — constructor khi build với -tags llama.
func NewModel(cfg Config) (Model, error) {
	if _, err := os.Stat(cfg.ModelPath); err != nil {
		return nil, fmt.Errorf("model file not found: %s — %w", cfg.ModelPath, err)
	}

	opts := []llama.ModelOption{
		llama.SetContext(cfg.ContextSize),
		llama.SetThreads(cfg.Threads),
		llama.SetEmbeddings(false),
	}
	if cfg.Temperature > 0 {
		opts = append(opts, llama.SetTemperature(cfg.Temperature))
	}

	model, err := llama.New(cfg.ModelPath, opts...)
	if err != nil {
		return nil, fmt.Errorf("load llama model: %w", err)
	}

	return &llamaModel{cfg: cfg, model: model}, nil
}

func (m *llamaModel) Predict(prompt string, maxTokens int, temp float32) (string, error) {
	// Apply temperature override if specified
	if temp > 0 && temp <= 0.5 {
		// go-llama.cpp Predict uses options set at model load time;
		// per-call temp not supported in current API. Document this.
	}
	opts := []llama.PredictOption{
		llama.SetTokens(maxTokens),
		llama.SetTopK(40),
		llama.SetTopP(0.9),
		llama.SetStop([]string{"###", "[/INST]", "</s>"}),
	}
	out, err := m.model.Predict(prompt, opts...)
	if err != nil {
		return "", fmt.Errorf("predict: %w", err)
	}
	return out, nil
}

func (m *llamaModel) Free() {
	if m.model != nil {
		m.model.Free()
		m.model = nil
	}
}

func (m *llamaModel) Info() ModelInfo {
	return ModelInfo{
		Name:        "qwen3-0.6b-instruct-q4_k_m (REAL llama.cpp)",
		Path:        m.cfg.ModelPath,
		ContextSize: m.cfg.ContextSize,
		Temperature: m.cfg.Temperature,
		Loaded:      true,
		Backend:     "llama.cpp",
	}
}
