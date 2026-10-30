// Package legal provides offline lookup of Vietnamese legal documents
// (Bộ luật, Luật, Nghị định, Thông tư, Nghị quyết, Quyết định) and
// án lệ từ TANDTC. Documents are stored in the JSON-backed store
// (internal/store) and seeded from internal/db/seed.go on first run.
package legal

import (
	"encoding/json"
	"fmt"
	"strings"

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

// Doc là một văn bản pháp luật.
type Doc struct {
	ID            int    `json:"id"`
	Loai          string `json:"loai"`
	SoHieu        string `json:"soHieu"`
	CoQuanBanHanh string `json:"coQuanBanHanh"`
	NgayBanHanh   string `json:"ngayBanHanh"`
	TrichYeu      string `json:"trichYeu"`
	TomTat        string `json:"tomTat"`
	NoiDung       string `json:"noiDung"`
	LinhVuc       string `json:"linhVuc"`
	ConHieuLuc    bool   `json:"conHieuLuc"`
}

// rowToDoc chuyển một store.Row thành Doc.
func rowToDoc(r store.Row) Doc {
	return Doc{
		ID:            store.FieldInt(r, "id"),
		Loai:          store.FieldStr(r, "loai"),
		SoHieu:        store.FieldStr(r, "so_hieu"),
		CoQuanBanHanh: store.FieldStr(r, "co_quan_ban_hanh"),
		NgayBanHanh:   store.FieldStr(r, "ngay_ban_hanh"),
		TrichYeu:      store.FieldStr(r, "trich_yeu"),
		TomTat:        store.FieldStr(r, "tom_tat"),
		NoiDung:       store.FieldStr(r, "noi_dung"),
		LinhVuc:       store.FieldStr(r, "linh_vuc"),
		ConHieuLuc:    store.FieldBool(r, "con_hieu_luc"),
	}
}

// Search tra cứu văn bản theo từ khóa (số hiệu, trích yếu, nội dung).
func (s *Service) Search(query, linhVuc string, limit int) []Doc {
	if limit <= 0 {
		limit = 50
	}
	q := strings.ToLower(strings.TrimSpace(query))
	rows := s.db.Legal.Table("legal_docs").Filter(func(r store.Row) bool {
		if linhVuc != "" && store.FieldStr(r, "linh_vuc") != linhVuc {
			return false
		}
		if q == "" {
			return true
		}
		soHieu := strings.ToLower(store.FieldStr(r, "so_hieu"))
		trichYeu := strings.ToLower(store.FieldStr(r, "trich_yeu"))
		tomTat := strings.ToLower(store.FieldStr(r, "tom_tat"))
		noiDung := strings.ToLower(store.FieldStr(r, "noi_dung"))
		return strings.Contains(soHieu, q) || strings.Contains(trichYeu, q) ||
			strings.Contains(tomTat, q) || strings.Contains(noiDung, q)
	})
	// Sort by ngay_ban_hanh DESC
	sorted := make([]store.Row, len(rows))
	copy(sorted, rows)
	// simple insertion sort by ngay_ban_hanh desc
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && store.FieldStr(sorted[j], "ngay_ban_hanh") > store.FieldStr(sorted[j-1], "ngay_ban_hanh"); j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	if len(sorted) > limit {
		sorted = sorted[:limit]
	}
	out := make([]Doc, 0, len(sorted))
	for _, r := range sorted {
		out = append(out, rowToDoc(r))
	}
	return out
}

// GetByID trả về toàn văn một văn bản.
func (s *Service) GetByID(id int) *Doc {
	r := s.db.Legal.Table("legal_docs").First(func(r store.Row) bool {
		return store.FieldInt(r, "id") == id
	})
	if r == nil {
		return nil
	}
	d := rowToDoc(r)
	if s.audit != nil {
		s.audit.Logf("legal.lookup", "view", "", "doc id=%d (%s)", id, d.SoHieu)
	}
	return &d
}

// ListAnLe liệt kê án lệ TANDTC.
func (s *Service) ListAnLe(limit int) []Doc {
	if limit <= 0 {
		limit = 100
	}
	rows := s.db.Legal.Table("legal_docs").Filter(func(r store.Row) bool {
		return store.FieldStr(r, "loai") == "Án lệ"
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]Doc, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToDoc(r))
	}
	return out
}

// ListByLinhVuc liệt kê văn bản theo lĩnh vực.
func (s *Service) ListByLinhVuc(linhVuc string, limit int) []Doc {
	if limit <= 0 {
		limit = 100
	}
	rows := s.db.Legal.Table("legal_docs").Filter(func(r store.Row) bool {
		return store.FieldStr(r, "linh_vuc") == linhVuc
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]Doc, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToDoc(r))
	}
	return out
}

// BieumauInfo — một biểu mẫu tố tụng.
type BieumauInfo struct {
	ID       int    `json:"id"`
	MaMau    string `json:"maMau"`
	TenMau   string `json:"tenMau"`
	LinhVuc string `json:"linhVuc"`
	GiaiDoan string `json:"giaiDoan"`
	FileHTML string `json:"fileHTML"`
	GhiChu   string `json:"ghiChu"`
}

func rowToBieumau(r store.Row) BieumauInfo {
	return BieumauInfo{
		ID:       store.FieldInt(r, "id"),
		MaMau:    store.FieldStr(r, "ma_mau"),
		TenMau:   store.FieldStr(r, "ten_mau"),
		LinhVuc:  store.FieldStr(r, "linh_vuc"),
		GiaiDoan: store.FieldStr(r, "giai_doan"),
		FileHTML: store.FieldStr(r, "file_html"),
		GhiChu:   store.FieldStr(r, "ghi_chu"),
	}
}

// ListBieuMau liệt kê biểu mẫu theo lĩnh vực và giai đoạn.
func (s *Service) ListBieuMau(linhVuc, giaiDoan string) []BieumauInfo {
	rows := s.db.Legal.Table("bieu_mau").Filter(func(r store.Row) bool {
		if linhVuc != "" && store.FieldStr(r, "linh_vuc") != linhVuc {
			return false
		}
		if giaiDoan != "" && store.FieldStr(r, "giai_doan") != giaiDoan {
			return false
		}
		return true
	})
	out := make([]BieumauInfo, 0, len(rows))
	for _, r := range rows {
		out = append(out, rowToBieumau(r))
	}
	return out
}

// Stats trả về số liệu thống kê CSDL luật theo lĩnh vực.
func (s *Service) Stats() map[string]int {
	stats := map[string]int{}
	rows := s.db.Legal.Table("legal_docs").All()
	for _, r := range rows {
		stats[store.FieldStr(r, "linh_vuc")]++
	}
	if s.audit != nil {
		d, _ := json.Marshal(stats)
		s.audit.Log("legal.stats", "view", "", string(d), "normal")
	}
	return stats
}

// Referenced returns text for human display in UI error messages.
func (s *Service) Referenced(linhVuc string) string {
	docs := s.ListByLinhVuc(linhVuc, 10)
	if len(docs) == 0 {
		return fmt.Sprintf("Chưa có văn bản trong lĩnh vực %s.", linhVuc)
	}
	parts := []string{}
	for _, d := range docs {
		parts = append(parts, fmt.Sprintf("%s %s (%s) — %s", d.Loai, d.SoHieu, d.NgayBanHanh, d.TrichYeu))
	}
	return strings.Join(parts, "\n")
}
