// worker.go — Worker Pool ngầm sử dụng Goroutines.
//
// Tuân thủ yêu cầu kiến trúc v1.0:
// 1. KHÔNG chạy AI trên main thread — dispatch qua jobs chan → workers xử lý
// 2. Quản lý hàng đợi: jobs chan Job (input) + results chan Result (output)
// 3. Model được khởi tạo 1 LẦN DUY NHẤT trong RAM của worker, defer model.Free()
//
// Lifecycle:
//   wp := NewWorkerPool(modelPath, numWorkers)
//   wp.Start()           // khởi tạo model + spawn workers goroutines
//   wp.Submit(Job{...})  // đẩy job vào queue
//   result := <-wp.Results()  // nhận kết quả
//   wp.Stop()            // đóng channels + defer model.Free()
package ai

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Job — yêu cầu xử lý AI từ UI.
type Job struct {
	ID         string                 `json:"id"`
	Module     string                 `json:"module"`     // extractor|reporter|chatter|auditor|advisor
	Input      map[string]interface{} `json:"input"`
	SubmittedAt time.Time             `json:"submittedAt"`
}

// Result — kết quả từ worker.
type Result struct {
	JobID     string        `json:"jobId"`
	Module    string        `json:"module"`
	Output    string        `json:"output"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	WorkerID  int           `json:"workerId"`
	CompletedAt time.Time   `json:"completedAt"`
}

// WorkerPool — pool goroutines xử lý AI jobs.
type WorkerPool struct {
	cfg        Config
	jobs       chan Job
	results    chan Result
	model      Model
	workers    int
	wg         sync.WaitGroup
	stopOnce   sync.Once
	stopped    atomic.Bool
	jobsCount  atomic.Int64
	doneCount  atomic.Int64
	log        func(string) // callback để đẩy log lines ra UI
}

// NewWorkerPool — tạo pool (chưa start).
// modelPath: đường dẫn file .gguf
// workers: số goroutines (mặc định 2 — ko quá 4 vì model nhẹ 0.6B)
func NewWorkerPool(modelPath string, workers int, logFn func(string)) *WorkerPool {
	if workers <= 0 {
		workers = 2
	}
	if logFn == nil {
		logFn = func(s string) {}
	}
	return &WorkerPool{
		cfg:     DefaultConfig(modelPath),
		workers: workers,
		jobs:    make(chan Job, 16),     // buffered queue, 16 jobs chờ được
		results: make(chan Result, 16),
		log:     logFn,
	}
}

// Start — load model + spawn workers goroutines.
// Phải gọi trước khi Submit.
func (wp *WorkerPool) Start() error {
	wp.log("[worker] Loading LLM model from " + wp.cfg.ModelPath + " ...")
	model, err := NewModel(wp.cfg)
	if err != nil {
		wp.log("[worker] ERROR loading model: " + err.Error())
		return err
	}
	wp.model = model
	wp.log(fmt.Sprintf("[worker] Model loaded: %s, context=%d, temp=%.2f",
		model.Info().Name, model.Info().ContextSize, model.Info().Temperature))

	// Spawn workers
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.runWorker(i)
	}
	wp.log(fmt.Sprintf("[worker] Spawned %d worker goroutines", wp.workers))
	return nil
}

// runWorker — vòng lặp chính của 1 worker goroutine.
func (wp *WorkerPool) runWorker(id int) {
	defer wp.wg.Done()
	wp.log(fmt.Sprintf("[worker-%d] started, waiting for jobs...", id))
	for job := range wp.jobs {
		wp.log(fmt.Sprintf("[worker-%d] picked up job %s (module=%s)", id, job.ID, job.Module))
		start := time.Now()
		out, err := dispatchModule(wp.model, job)
		dur := time.Since(start)
		res := Result{
			JobID:       job.ID,
			Module:      job.Module,
			Output:      out,
			Duration:    dur,
			WorkerID:    id,
			CompletedAt: time.Now(),
		}
		if err != nil {
			res.Error = err.Error()
			wp.log(fmt.Sprintf("[worker-%d] job %s FAILED after %v: %s",
				id, job.ID, dur, err.Error()))
		} else {
			wp.log(fmt.Sprintf("[worker-%d] job %s completed in %v", id, job.ID, dur))
		}
		wp.doneCount.Add(1)
		wp.results <- res
	}
	wp.log(fmt.Sprintf("[worker-%d] stopped (jobs channel closed)", id))
}

// Submit — đẩy job vào queue (non-blocking, không chặn UI).
// Trả về error nếu queue đầy hoặc pool đã stop.
func (wp *WorkerPool) Submit(job Job) error {
	if wp.stopped.Load() {
		return fmt.Errorf("worker pool stopped")
	}
	wp.jobsCount.Add(1)
	select {
	case wp.jobs <- job:
		wp.log(fmt.Sprintf("[ui] submitted job %s (queue=%d/%d)",
			job.ID, len(wp.jobs), cap(wp.jobs)))
		return nil
	default:
		return fmt.Errorf("job queue full (capacity=%d)", cap(wp.jobs))
	}
}

// Results — channel để UI nhận kết quả (range hoặc select).
func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

// Stop — đóng jobs channel → workers thoát vòng lặp → defer model.Free().
// Idempotent — gọi nhiều lần OK.
func (wp *WorkerPool) Stop() {
	wp.stopOnce.Do(func() {
		wp.stopped.Store(true)
		wp.log("[worker] Stopping worker pool...")
		close(wp.jobs)
		wp.wg.Wait()        // chờ tất cả workers finish
		close(wp.results)
		if wp.model != nil {
			wp.log("[worker] Freeing model memory (defer model.Free())")
			wp.model.Free() // GIẢI PHÓNG RAM — tránh memory leak
			wp.model = nil
		}
		wp.log("[worker] Worker pool stopped")
	})
}

// Stats — metrics cho UI dashboard.
type Stats struct {
	JobsSubmitted int64 `json:"jobsSubmitted"`
	JobsCompleted int64 `json:"jobsCompleted"`
	QueueLen      int   `json:"queueLen"`
	Workers       int   `json:"workers"`
	ModelInfo     ModelInfo `json:"modelInfo"`
}

func (wp *WorkerPool) Stats() Stats {
	return Stats{
		JobsSubmitted: wp.jobsCount.Load(),
		JobsCompleted: wp.doneCount.Load(),
		QueueLen:      len(wp.jobs),
		Workers:       wp.workers,
		ModelInfo: func() ModelInfo {
			if wp.model != nil {
				return wp.model.Info()
			}
			return ModelInfo{Loaded: false}
		}(),
	}
}

// LogChannel — variant: log lines được push qua channel thay vì callback.
// Dùng khi UI cần stream log real-time 60FPS mà không poll.
func (wp *WorkerPool) StartWithLogChannel(logCh chan<- string) error {
	wp.log = func(s string) {
		select {
		case logCh <- s:
		default: // drop if consumer slow
		}
	}
	return wp.Start()
}
