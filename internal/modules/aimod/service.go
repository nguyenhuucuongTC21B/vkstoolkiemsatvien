// Package aimod — Wails binding cho AI Worker Pool.
//
// Frontend gọi qua:
//   - go.aimod.Service.Submit(job) → jobID (non-blocking)
//   - go.aimod.Service.PollResults() → []Result (drain channel)
//   - go.aimod.Service.Stats() → worker pool metrics
//   - go.aimod.Service.GetLog() → []string (recent log lines, 60FPS polling)
package aimod

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"kiem_sat_toolkit/internal/ai"
	"kiem_sat_toolkit/internal/audit"
)

// Service — bound to frontend as `window.go.aimod.Service`.
type Service struct {
	pool      *ai.WorkerPool
	audit     *audit.Logger
	logBuffer *LogBuffer
}

// LogBuffer — ring buffer giữ N log lines gần nhất cho UI poll.
type LogBuffer struct {
	mu    sync.Mutex
	lines []string
	max   int
	head  int
	count int
}

func NewLogBuffer(max int) *LogBuffer {
	return &LogBuffer{lines: make([]string, max), max: max}
}

func (b *LogBuffer) Push(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines[b.head] = line
	b.head = (b.head + 1) % b.max
	if b.count < b.max {
		b.count++
	}
}

// Recent — trả N dòng log gần nhất theo thứ tự cũ→mới.
func (b *LogBuffer) Recent(n int) []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if n > b.count {
		n = b.count
	}
	out := make([]string, n)
	start := (b.head - n + b.max) % b.max
	for i := 0; i < n; i++ {
		out[i] = b.lines[(start+i)%b.max]
	}
	return out
}

// New — constructor. modelPath relative to dataDir.
func New(dataDir string, auditLog *audit.Logger) *Service {
	modelPath := findModel(dataDir)
	logBuf := NewLogBuffer(500)
	svc := &Service{
		audit:     auditLog,
		logBuffer: logBuf,
	}
	svc.pool = ai.NewWorkerPool(modelPath, 2, func(line string) {
		logBuf.Push(fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), line))
	})
	return svc
}

// findModel — tìm file .gguf.
func findModel(dataDir string) string {
	candidates := []string{
		filepath.Join(dataDir, "models", "qwen3-0.6b-instruct-q4_k_m.gguf"),
		filepath.Join("models", "qwen3-0.6b-instruct-q4_k_m.gguf"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(dataDir, "models", "qwen3-0.6b-instruct-q4_k_m.gguf")
}

// Start — khởi tạo worker pool (load model).
func (s *Service) Start() error {
	if err := s.pool.Start(); err != nil {
		s.logBuffer.Push("[ERROR] " + err.Error())
		return err
	}
	return nil
}

// Stop — shutdown worker pool + free model RAM. MUST call on app shutdown.
func (s *Service) Stop() {
	s.pool.Stop()
}

// ─────────── Bound methods cho frontend ───────────

// SubmitReq — payload cho SubmitJob.
type SubmitReq struct {
	Module string                 `json:"module"`
	Input  map[string]interface{} `json:"input"`
}

// SubmitJob — đẩy 1 job vào queue (non-blocking).
func (s *Service) SubmitJob(req SubmitReq) (string, error) {
	job := ai.Job{
		ID:          fmt.Sprintf("job_%d", time.Now().UnixNano()),
		Module:      req.Module,
		Input:       req.Input,
		SubmittedAt: time.Now(),
	}
	if err := s.pool.Submit(job); err != nil {
		return "", err
	}
	if s.audit != nil {
		s.audit.Logf("ai", "submit", "ui", "job=%s module=%s", job.ID, job.Module)
	}
	return job.ID, nil
}

// PollResults — drain results channel. UI gọi mỗi 16ms (60FPS).
func (s *Service) PollResults() []ai.Result {
	var results []ai.Result
	for {
		select {
		case r, ok := <-s.pool.Results():
			if !ok {
				return results
			}
			results = append(results, r)
		default:
			return results
		}
	}
}

// GetLog — trả 50 dòng log gần nhất.
func (s *Service) GetLog() []string {
	return s.logBuffer.Recent(50)
}

// Stats — metrics cho dashboard.
func (s *Service) Stats() ai.Stats {
	return s.pool.Stats()
}

// ListLaws — tra cứu luật trực tiếp từ laws.json (no AI).
func (s *Service) ListLaws(keyword string, limit int) []ai.Dieu {
	return ai.SearchLaws(keyword, limit)
}
