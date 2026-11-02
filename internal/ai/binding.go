// Package ai — binding interface cho LLM model.
//
// Có 2 implementation:
//   - binding.go (default, no build tag): STUB trả về mock JSON hợp lệ.
//     Dùng cho development/test mà không cần libllama cài sẵn.
//   - binding_llama.go (build tag "llama"): REAL implementation dùng
//     github.com/go-skynet/go-llama.cpp để load file .gguf thật.
//
// Build với real binding:
//   go build -tags "llama,desktop,production" -o ks_tools .
// (cần libllama.a build từ llama.cpp, xem README_AI.md)
package ai

// Model — abstract interface cho LLM binding.
// Implementations: stubModel (default) hoặc llamaModel (build tag llama).
type Model interface {
	// Predict chạy inference với prompt, trả về text output.
	// maxTokens giới hạn độ dài output. temp = 0.0..0.5.
	Predict(prompt string, maxTokens int, temp float32) (string, error)

	// Free giải phóng RAM — phải gọi defer khi shutdown.
	Free()

	// Info trả về metadata của model (tên, context size, v.v.).
	Info() ModelInfo
}

// ModelInfo — metadata của LLM model.
type ModelInfo struct {
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	ContextSize  int     `json:"contextSize"`
	Temperature  float32 `json:"temperature"`
	Loaded       bool    `json:"loaded"`
	Backend      string  `json:"backend"` // "stub" | "llama.cpp"
}

// Config — cấu hình load model.
type Config struct {
	ModelPath    string  `json:"modelPath"`
	ContextSize  int     `json:"contextSize"`
	Temperature  float32 `json:"temperature"`
	MaxTokens    int     `json:"maxTokens"`
	Threads      int     `json:"threads"` // số CPU threads cho inference
}

// DefaultConfig — cấu hình mặc định theo spec v1.0.
func DefaultConfig(modelPath string) Config {
	return Config{
		ModelPath:   modelPath,
		ContextSize: 2048,
		Temperature: 0.3, // 0.0-0.5 range, default 0.3
		MaxTokens:   512,
		Threads:     4,
	}
}
