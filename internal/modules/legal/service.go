// Package legal provides offline lookup of Vietnamese legal documents
// (Bộ luật, Luật, Nghị định, Thông tư, Nghị quyết, Quyết định) and
// án lệ từ TANDTC. Documents are stored in the JSON-backed store
// (internal/store) and seeded from internal/db/seed.go on first run.
package legal

import (
        "encoding/json"
        "fmt"
        "regexp"
        "sort"
        "strconv"
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

// Doc là một văn bản pháp luật (flat format, backward compat).
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

// DocFull — văn bản có cấu trúc Điều/Khoản/Điểm đầy đủ.
type DocFull struct {
        ID            int          `json:"id"`
        Loai          string       `json:"loai"`
        SoHieu        string       `json:"soHieu"`
        CoQuanBanHanh string       `json:"coQuanBanHanh"`
        NgayBanHanh   string       `json:"ngayBanHanh"`
        TrichYeu      string       `json:"trichYeu"`
        TomTat        string       `json:"tomTat"`
        LinhVuc       string       `json:"linhVuc"`
        ConHieuLuc    bool         `json:"conHieuLuc"`
        DieuList      []db.Dieu    `json:"dieuList"`
}

// Dieu — alias để frontend nhận type.
type Dieu = db.Dieu
type Khoan = db.Khoan
type Diem = db.Diem

// rowToDoc chuyển một store.Row thành Doc flat.
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

// structuredSeedByID returns one LegalDocStructured from db package.
func structuredSeedByID(id int) *db.LegalDocStructured {
        for i := range db.LegalDocsStructuredSeed {
                if db.LegalDocsStructuredSeed[i].ID == id {
                        return &db.LegalDocsStructuredSeed[i]
                }
        }
        return nil
}

// allStructuredSeed returns the full structured seed.
func allStructuredSeed() []db.LegalDocStructured {
        return db.LegalDocsStructuredSeed
}

// Search tra cứu văn bản theo từ khóa (số hiệu, trích yếu, nội dung, điều/khoản).
// Trả về danh sách Doc (flat format) cho danh sách kết quả.
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
        sorted := make([]store.Row, len(rows))
        copy(sorted, rows)
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

// GetByID trả về flat Doc (cho backward compat).
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

// GetFullByID returns the full structured document (with Dieu/Khoan/Diem).
// Falls back to flat NoiDung if no structured data available.
func (s *Service) GetFullByID(id int) *DocFull {
        // First, find in flat store to get metadata
        r := s.db.Legal.Table("legal_docs").First(func(r store.Row) bool {
                return store.FieldInt(r, "id") == id
        })
        if r == nil {
                return nil
        }
        // Then, find structured seed (which has the DieuList)
        structDoc := structuredSeedByID(id)
        if structDoc == nil {
                // No structured data — return DocFull with single Dieu containing flat NoiDung
                flat := rowToDoc(r)
                return &DocFull{
                        ID:            flat.ID,
                        Loai:          flat.Loai,
                        SoHieu:        flat.SoHieu,
                        CoQuanBanHanh: flat.CoQuanBanHanh,
                        NgayBanHanh:   flat.NgayBanHanh,
                        TrichYeu:      flat.TrichYeu,
                        TomTat:        flat.TomTat,
                        LinhVuc:       flat.LinhVuc,
                        ConHieuLuc:    flat.ConHieuLuc,
                        DieuList: []Dieu{{
                                SoDieu:  0,
                                TenDieu: "Toàn văn",
                                KhoanList: []Khoan{{SoKhoan: 1, NoiDung: flat.NoiDung}},
                        }},
                }
        }
        // Merge flat metadata + structured dieu list
        flat := rowToDoc(r)
        out := &DocFull{
                ID:            flat.ID,
                Loai:          flat.Loai,
                SoHieu:        flat.SoHieu,
                CoQuanBanHanh: structDoc.CoQuanBanHanh,
                NgayBanHanh:   structDoc.NgayBanHanh,
                TrichYeu:      structDoc.TrichYeu,
                TomTat:        structDoc.TomTat,
                LinhVuc:       structDoc.LinhVuc,
                ConHieuLuc:    structDoc.ConHieuLuc,
                DieuList:      structDoc.DieuList,
        }
        if s.audit != nil {
                s.audit.Logf("legal.lookup", "view_full", "", "doc id=%d (%s), %d điều",
                        id, out.SoHieu, len(out.DieuList))
        }
        return out
}

// SearchDieu — tìm theo số điều trong một văn bản cụ thể.
// Ví dụ: docID=1 (BLHS), soDieu=27 → trả về điều 27 "Thời hiệu TNHS".
func (s *Service) SearchDieu(soHieu string, soDieu int) *Dieu {
        structDoc := findStructuredBySoHieu(soHieu)
        if structDoc == nil {
                return nil
        }
        for i := range structDoc.DieuList {
                if structDoc.DieuList[i].SoDieu == soDieu {
                        return &structDoc.DieuList[i]
                }
        }
        return nil
}

// SearchByDieuNumber — tìm tất cả điều có số được chỉ định trong tất cả văn bản.
// Ví dụ: soDieu=27 → trả về điều 27 của BLHS (thời hiệu TNHS) + điều 27 của các văn bản khác nếu có.
func (s *Service) SearchByDieuNumber(soDieu int) []DieuSearchResult {
        results := []DieuSearchResult{}
        for _, doc := range allStructuredSeed() {
                for _, dieu := range doc.DieuList {
                        if dieu.SoDieu == soDieu {
                                results = append(results, DieuSearchResult{
                                        DocID:       doc.ID,
                                        DocSoHieu:   doc.SoHieu,
                                        DocTrichYeu: doc.TrichYeu,
                                        Dieu:        dieu,
                                })
                        }
                }
        }
        return results
}

// DieuSearchResult — kết quả tìm điều theo số.
type DieuSearchResult struct {
        DocID       int    `json:"docId"`
        DocSoHieu   string `json:"docSoHieu"`
        DocTrichYeu string `json:"docTrichYeu"`
        Dieu        Dieu   `json:"dieu"`
}

// SearchKhoan — tìm theo từ khóa trong nội dung khoản của tất cả văn bản.
// Trả về danh sách khoản khớp, kèm điều + văn bản gốc.
func (s *Service) SearchKhoan(keyword string, limit int) []KhoanSearchResult {
        if limit <= 0 {
                limit = 50
        }
        q := strings.ToLower(strings.TrimSpace(keyword))
        if q == "" {
                return nil
        }
        results := []KhoanSearchResult{}
        for _, doc := range allStructuredSeed() {
                for _, dieu := range doc.DieuList {
                        for _, khoan := range dieu.KhoanList {
                                if strings.Contains(strings.ToLower(khoan.NoiDung), q) {
                                        results = append(results, KhoanSearchResult{
                                                DocID:       doc.ID,
                                                DocSoHieu:   doc.SoHieu,
                                                DocTrichYeu: doc.TrichYeu,
                                                SoDieu:      dieu.SoDieu,
                                                TenDieu:     dieu.TenDieu,
                                                Khoan:       khoan,
                                        })
                                        if len(results) >= limit {
                                                return results
                                        }
                                }
                                // Also search in Diem
                                for _, diem := range khoan.DiemList {
                                        if strings.Contains(strings.ToLower(diem.NoiDung), q) {
                                                results = append(results, KhoanSearchResult{
                                                        DocID:       doc.ID,
                                                        DocSoHieu:   doc.SoHieu,
                                                        DocTrichYeu: doc.TrichYeu,
                                                        SoDieu:      dieu.SoDieu,
                                                        TenDieu:     dieu.TenDieu,
                                                        Khoan:       khoan,
                                                        DiemHit:     diem.SoDiem,
                                                })
                                                if len(results) >= limit {
                                                        return results
                                                }
                                        }
                                }
                        }
                }
        }
        return results
}

// KhoanSearchResult — kết quả tìm khoản theo từ khóa.
type KhoanSearchResult struct {
        DocID       int    `json:"docId"`
        DocSoHieu   string `json:"docSoHieu"`
        DocTrichYeu string `json:"docTrichYeu"`
        SoDieu      int    `json:"soDieu"`
        TenDieu     string `json:"tenDieu"`
        Khoan       Khoan  `json:"khoan"`
        DiemHit     string `json:"diemHit,omitempty"` // nếu match ở điểm nào
}

// SmartSearch — tự động detect loại query và trả kết quả phù hợp:
// - "điều 27" / "điều 27 BLHS" → searchByDieuNumber
// - "thời hiệu" / "án phí" → searchKhoan (full text)
// - "326/2016" → search theo số hiệu văn bản
// - number alone → assume điều number
type SmartSearchResult struct {
        Query     string              `json:"query"`
        QueryType string              `json:"queryType"` // "dieu_number" | "keyword" | "so_hieu" | "all"
        Docs      []Doc               `json:"docs"`      // danh sách văn bản (flat)
        DieuHits  []DieuSearchResult  `json:"dieuHits"`  // điều khớp số
        KhoanHits []KhoanSearchResult `json:"khoanHits"` // khoản khớp từ khóa
}

func (s *Service) SmartSearch(query, linhVuc string, limit int) SmartSearchResult {
        if limit <= 0 {
                limit = 50
        }
        q := strings.TrimSpace(query)
        result := SmartSearchResult{Query: q}

        // Detect pattern: "điều X" / "dieu X" / " Điều X "
        dieuRE := regexp.MustCompile(`(?i)điều\s+(\d+)`)
        if m := dieuRE.FindStringSubmatch(q); m != nil {
                soDieu, _ := strconv.Atoi(m[1])
                result.QueryType = "dieu_number"
                result.DieuHits = s.SearchByDieuNumber(soDieu)
                // Also do keyword search on remainder
                rem := strings.TrimSpace(dieuRE.ReplaceAllString(q, ""))
                if rem != "" {
                        result.KhoanHits = s.SearchKhoan(rem, limit)
                }
                return result
        }

        // Detect: pure number → assume điều
        if n, err := strconv.Atoi(q); err == nil {
                result.QueryType = "dieu_number"
                result.DieuHits = s.SearchByDieuNumber(n)
                return result
        }

        // Detect: so hieu (contains "/" and digits)
        soHieuRE := regexp.MustCompile(`^\d+/\d+`)
        if soHieuRE.MatchString(q) {
                result.QueryType = "so_hieu"
                result.Docs = s.Search(q, linhVuc, limit)
                return result
        }

        // Default: keyword search in both docs and khoans
        result.QueryType = "keyword"
        result.Docs = s.Search(q, linhVuc, limit)
        result.KhoanHits = s.SearchKhoan(q, limit)
        return result
}

func findStructuredBySoHieu(soHieu string) *db.LegalDocStructured {
        for i := range db.LegalDocsStructuredSeed {
                if db.LegalDocsStructuredSeed[i].SoHieu == soHieu {
                        return &db.LegalDocsStructuredSeed[i]
                }
        }
        return nil
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
        // Count structured conditions
        stats["_dieu_count"] = 0
        stats["_khoan_count"] = 0
        for _, doc := range allStructuredSeed() {
                stats["_dieu_count"] += len(doc.DieuList)
                for _, d := range doc.DieuList {
                        stats["_khoan_count"] += len(d.KhoanList)
                }
        }
        if s.audit != nil {
                d, _ := json.Marshal(stats)
                s.audit.Log("legal.stats", "view", "", string(d), "normal")
        }
        return stats
}

// Referenced returns text for human display.
func (s *Service) Referenced(linhVuc string) string {
        docs := s.ListByLinhVuc(linhVuc, 10)
        if len(docs) == 0 {
                return fmt.Sprintf("Chưa có văn bản trong lĩnh vực %s.", linhVuc)
        }
        parts := []string{}
        for _, d := range docs {
                parts = append(parts, fmt.Sprintf("%s %s (%s) — %s", d.Loai, d.SoHieu, d.NgayBanHanh, d.TrichYeu))
        }
        sort.Strings(parts)
        return strings.Join(parts, "\n")
}
