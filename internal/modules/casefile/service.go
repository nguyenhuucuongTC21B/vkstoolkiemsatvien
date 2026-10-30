// Package casefile manages prosecutor case files — the unit of work
// for everything else in the toolkit. A case binds a case number,
// type (HS/DS/HC/THADS/THAHS/KN-TC/KKCI), stage (thu_ly → dieu_tra →
// truy_to → xet_xu → thi_hanh), assigned prosecutor, deadlines, and
// metadata. All other modules (calc, evidence, ocr, pii) reference
// back to a case_id for audit trail continuity.
package casefile

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"kiem_sat_toolkit/internal/audit"
	"kiem_sat_toolkit/internal/db"
	"kiem_sat_toolkit/internal/store"
)

type Service struct {
	db    *db.Manager
	audit *audit.Logger
}

func New(dbMgr *db.Manager, auditLog *audit.Logger) *Service {
	return &Service{db: dbMgr, audit: auditLog}
}

// Case là một hồ sơ vụ án.
type Case struct {
	ID           string                 `json:"id"`
	CaseNumber   string                 `json:"caseNumber"`
	CaseType     string                 `json:"caseType"`
	Stage        string                 `json:"stage"`
	AssignedKSV  string                 `json:"assignedKsv"`
	Parties      []string               `json:"parties"`
	ReceivedDate string                 `json:"receivedDate"`
	Deadline     string                 `json:"deadline"`
	Status       string                 `json:"status"`
	Metadata     map[string]interface{} `json:"metadata"`
	Notes        string                 `json:"notes"`
	CreatedAt    string                 `json:"createdAt"`
	UpdatedAt    string                 `json:"updatedAt"`
}

func rowToCase(r store.Row) Case {
	var parties []string
	_ = json.Unmarshal([]byte(store.FieldStr(r, "parties")), &parties)
	var meta map[string]interface{}
	_ = json.Unmarshal([]byte(store.FieldStr(r, "metadata")), &meta)
	return Case{
		ID:           store.FieldStr(r, "id"),
		CaseNumber:   store.FieldStr(r, "case_number"),
		CaseType:     store.FieldStr(r, "case_type"),
		Stage:        store.FieldStr(r, "stage"),
		AssignedKSV:  store.FieldStr(r, "assigned_ksv"),
		Parties:      parties,
		ReceivedDate: store.FieldStr(r, "received_date"),
		Deadline:     store.FieldStr(r, "deadline"),
		Status:       store.FieldStr(r, "status"),
		Metadata:     meta,
		Notes:        store.FieldStr(r, "notes"),
		CreatedAt:    store.FieldStr(r, "created_at"),
		UpdatedAt:    store.FieldStr(r, "updated_at"),
	}
}

func caseToRow(c Case) store.Row {
	parties, _ := json.Marshal(c.Parties)
	meta, _ := json.Marshal(c.Metadata)
	if c.Status == "" {
		c.Status = "mo"
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	if c.CreatedAt == "" {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	return store.Row{
		"id":            c.ID,
		"case_number":   c.CaseNumber,
		"case_type":     c.CaseType,
		"stage":         c.Stage,
		"assigned_ksv":  c.AssignedKSV,
		"parties":       string(parties),
		"received_date": c.ReceivedDate,
		"deadline":      c.Deadline,
		"status":        c.Status,
		"metadata":      string(meta),
		"notes":         c.Notes,
		"created_at":    c.CreatedAt,
		"updated_at":    c.UpdatedAt,
	}
}

// Create thêm một vụ án mới.
func (s *Service) Create(c Case, actor string) (*Case, error) {
	if c.ID == "" {
		c.ID = newID("case")
	}
	if c.Status == "" {
		c.Status = "mo"
	}
	if err := s.db.App.Table("cases").Insert(caseToRow(c)); err != nil {
		return nil, fmt.Errorf("insert case: %w", err)
	}
	if s.audit != nil {
		s.audit.Logf("case", "create", actor, "case=%s type=%s", c.ID, c.CaseType)
	}
	return &c, nil
}

// Get retrieves a case by ID.
func (s *Service) Get(id string) *Case {
	r := s.db.App.Table("cases").First(func(r store.Row) bool {
		return store.FieldStr(r, "id") == id
	})
	if r == nil {
		return nil
	}
	c := rowToCase(r)
	return &c
}

// List trả về danh sách vụ án theo bộ lọc.
func (s *Service) List(caseType, status string, limit int) []Case {
	if limit <= 0 {
		limit = 100
	}
	rows := s.db.App.Table("cases").Filter(func(r store.Row) bool {
		if caseType != "" && store.FieldStr(r, "case_type") != caseType {
			return false
		}
		if status != "" && store.FieldStr(r, "status") != status {
			return false
		}
		return true
	})
	// Sort by updated_at DESC
	sort.SliceStable(rows, func(i, j int) bool {
		return store.FieldStr(rows[i], "updated_at") > store.FieldStr(rows[j], "updated_at")
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]Case, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToCase(r))
	}
	return out
}

// Update cập nhật vụ án (toàn bộ trường).
func (s *Service) Update(c Case, actor string) error {
	patch := caseToRow(c)
	if err := s.db.App.Table("cases").Update(
		func(r store.Row) bool { return store.FieldStr(r, "id") == c.ID }, patch); err != nil {
		return err
	}
	if s.audit != nil {
		s.audit.Logf("case", "update", actor, "case=%s", c.ID)
	}
	return nil
}

// SetStatus chỉ thay đổi trạng thái.
func (s *Service) SetStatus(id, status, actor string) error {
	patch := store.Row{
		"status":     status,
		"updated_at": time.Now().Format("2006-01-02 15:04:05"),
	}
	if err := s.db.App.Table("cases").Update(
		func(r store.Row) bool { return store.FieldStr(r, "id") == id }, patch); err != nil {
		return err
	}
	if s.audit != nil {
		s.audit.Logf("case", "status_change", actor, "case=%s -> %s", id, status)
	}
	return nil
}

// Delete — soft delete (set status='dong') — never actually DROP.
func (s *Service) Delete(id, actor string) error {
	return s.SetStatus(id, "dong", actor)
}

// UpcomingDeadlines returns cases whose deadline is within N days.
func (s *Service) UpcomingDeadlines(days int) []Case {
	now := time.Now()
	cutoff := now.AddDate(0, 0, days).Format("2006-01-02")
	rows := s.db.App.Table("cases").Filter(func(r store.Row) bool {
		if store.FieldStr(r, "status") != "mo" {
			return false
		}
		dl := store.FieldStr(r, "deadline")
		if dl == "" {
			return false
		}
		// Compare as YYYY-MM-DD strings (lexicographic = chronological for ISO dates)
		return dl >= now.Format("2006-01-02") && dl <= cutoff
	})
	sort.SliceStable(rows, func(i, j int) bool {
		return store.FieldStr(rows[i], "deadline") < store.FieldStr(rows[j], "deadline")
	})
	out := make([]Case, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToCase(r))
	}
	return out
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
