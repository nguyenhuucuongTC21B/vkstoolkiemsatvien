// Package evidence manages labeled evidence markers attached to case
// files. Each marker points to a file + (page | timestamp) range and
// carries a free-form caption + color tag. Used both for trial
// presentation (chế độ trình chiếu 2 màn hình) and for cross-referencing
// documents during the investigation phase.
package evidence

import (
	"encoding/json"
	"fmt"
	"path/filepath"
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

// Evidence là một mục chứng cứ được gắn nhãn.
type Evidence struct {
	ID         string   `json:"id"`
	CaseID     string   `json:"caseId"`
	Label      string   `json:"label"`
	LabelColor string   `json:"labelColor"`
	FilePath   string   `json:"filePath"`
	FileType   string   `json:"fileType"`
	PageMarker string   `json:"pageMarker"`
	Caption    string   `json:"caption"`
	Tags       []string `json:"tags"`
	CreatedAt  string   `json:"createdAt"`
}

// LabelTree — cây nhãn nhiều cấp phục vụ trình chiếu tại phiên tòa.
type LabelTree struct {
	Category string     `json:"category"`
	Color    string     `json:"color"`
	Items    []Evidence `json:"items"`
}

func rowToEvidence(r store.Row) Evidence {
	var tags []string
	_ = json.Unmarshal([]byte(store.FieldStr(r, "tags")), &tags)
	return Evidence{
		ID:         store.FieldStr(r, "id"),
		CaseID:     store.FieldStr(r, "case_id"),
		Label:      store.FieldStr(r, "label"),
		LabelColor: store.FieldStr(r, "label_color"),
		FilePath:   store.FieldStr(r, "file_path"),
		FileType:   store.FieldStr(r, "file_type"),
		PageMarker: store.FieldStr(r, "page_marker"),
		Caption:    store.FieldStr(r, "caption"),
		Tags:       tags,
		CreatedAt:  store.FieldStr(r, "created_at"),
	}
}

// Add thêm một chứng cứ mới.
func (s *Service) Add(e Evidence, actor string) (*Evidence, error) {
	if e.ID == "" {
		e.ID = newID("ev")
	}
	if e.CreatedAt == "" {
		e.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	}
	if e.FileType == "" {
		e.FileType = guessFileType(e.FilePath)
	}
	tags, _ := json.Marshal(e.Tags)
	row := store.Row{
		"id":          e.ID,
		"case_id":     e.CaseID,
		"label":       e.Label,
		"label_color": e.LabelColor,
		"file_path":   e.FilePath,
		"file_type":   e.FileType,
		"page_marker": e.PageMarker,
		"caption":     e.Caption,
		"tags":        string(tags),
		"created_at":  e.CreatedAt,
	}
	if err := s.db.App.Table("evidence").Insert(row); err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}
	if s.audit != nil {
		s.audit.Logf("evidence", "add", actor, "ev=%s case=%s", e.ID, e.CaseID)
	}
	return &e, nil
}

// ListByCase liệt kê toàn bộ chứng cứ theo vụ án.
func (s *Service) ListByCase(caseID string) []Evidence {
	rows := s.db.App.Table("evidence").Filter(func(r store.Row) bool {
		return store.FieldStr(r, "case_id") == caseID
	})
	out := make([]Evidence, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToEvidence(r))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

// TreeByCase builds a hierarchical view grouped by label prefix.
func (s *Service) TreeByCase(caseID string) []LabelTree {
	items := s.ListByCase(caseID)
	groups := map[string][]Evidence{}
	colors := map[string]string{}
	order := []string{}
	for _, it := range items {
		cat := labelCategory(it.Label)
		if _, ok := groups[cat]; !ok {
			order = append(order, cat)
		}
		groups[cat] = append(groups[cat], it)
		if colors[cat] == "" {
			colors[cat] = it.LabelColor
		}
	}
	out := make([]LabelTree, 0, len(order))
	for _, cat := range order {
		out = append(out, LabelTree{Category: cat, Color: colors[cat], Items: groups[cat]})
	}
	return out
}

// Update modifies an existing evidence item.
func (s *Service) Update(e Evidence, actor string) error {
	tags, _ := json.Marshal(e.Tags)
	patch := store.Row{
		"label":       e.Label,
		"label_color": e.LabelColor,
		"file_path":   e.FilePath,
		"file_type":   e.FileType,
		"page_marker": e.PageMarker,
		"caption":     e.Caption,
		"tags":        string(tags),
	}
	if err := s.db.App.Table("evidence").Update(
		func(r store.Row) bool { return store.FieldStr(r, "id") == e.ID }, patch); err != nil {
		return err
	}
	if s.audit != nil {
		s.audit.Logf("evidence", "update", actor, "ev=%s", e.ID)
	}
	return nil
}

// Delete removes an evidence item (the underlying file is NOT touched).
func (s *Service) Delete(id, actor string) error {
	_, err := s.db.App.Table("evidence").Delete(func(r store.Row) bool {
		return store.FieldStr(r, "id") == id
	})
	if err != nil {
		return err
	}
	if s.audit != nil {
		s.audit.Logf("evidence", "delete", actor, "ev=%s", id)
	}
	return nil
}

// PresentationOrder returns the evidence items in the order set by
// the prosecutor for trial presentation. MVP: alphabetical by label.
func (s *Service) PresentationOrder(caseID string) []Evidence {
	return s.ListByCase(caseID)
}

func labelCategory(label string) string {
	for i, c := range label {
		if c == '-' {
			return label[:i]
		}
	}
	return label
}

func guessFileType(p string) string {
	ext := filepath.Ext(p)
	switch ext {
	case ".pdf":
		return "pdf"
	case ".png", ".jpg", ".jpeg", ".bmp", ".tif", ".tiff", ".webp":
		return "image"
	case ".mp4", ".mov", ".avi", ".mkv":
		return "video"
	case ".mp3", ".wav", ".m4a", ".ogg":
		return "audio"
	case ".docx", ".doc":
		return "docx"
	default:
		return "other"
	}
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
