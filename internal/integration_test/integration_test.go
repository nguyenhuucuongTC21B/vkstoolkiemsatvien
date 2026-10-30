// Package integration_test runs end-to-end tests for the whole
// registry: init store → seed → call each module service → verify.
// Purpose: catch wiring bugs that unit tests miss.
package integration_test

import (
        "os"
        "path/filepath"
        "strings"
        "testing"

        "kiem_sat_toolkit/internal/audit"
        "kiem_sat_toolkit/internal/db"
        "kiem_sat_toolkit/internal/modules/calc"
        "kiem_sat_toolkit/internal/modules/casefile"
        "kiem_sat_toolkit/internal/modules/docs"
        "kiem_sat_toolkit/internal/modules/evidence"
        "kiem_sat_toolkit/internal/modules/file"
        "kiem_sat_toolkit/internal/modules/legal"
        "kiem_sat_toolkit/internal/modules/ocr"
        "kiem_sat_toolkit/internal/modules/pii"
        "kiem_sat_toolkit/internal/modules/spell"
)

// setup creates a temp data dir, inits Manager + audit.Logger,
// returns services wired exactly as production.
func setup(t *testing.T) (string, *db.Manager, func()) {
        t.Helper()
        tmp, err := os.MkdirTemp("", "kst_test_*")
        if err != nil {
                t.Fatalf("mkdtemp: %v", err)
        }
        mgr, err := db.NewManager(tmp)
        if err != nil {
                os.RemoveAll(tmp)
                t.Fatalf("NewManager: %v", err)
        }
        cleanup := func() {
                mgr.Close()
                os.RemoveAll(tmp)
        }
        return tmp, mgr, cleanup
}

// ─────────────── store + seed ───────────────

func TestStoreSeedsLegalDocs(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        rows := mgr.Legal.Table("legal_docs").All()
        if len(rows) < 10 {
                t.Errorf("legal_docs seed: got %d rows, want >= 10", len(rows))
        }
        r := mgr.Legal.Table("legal_docs").First(func(r map[string]interface{}) bool {
                return r["so_hieu"] == "326/2016/UBTVQH14"
        })
        if r == nil {
                t.Errorf("NR 326/2016 missing from seed")
        }
}

func TestStoreSeedsBieuMau(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        if n := mgr.Legal.Table("bieu_mau").Count(); n < 5 {
                t.Errorf("bieu_mau seed: got %d, want >= 5", n)
        }
}

func TestStoreSeedsSpellRules(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        if n := mgr.App.Table("spell_rules").Count(); n < 10 {
                t.Errorf("spell_rules seed: got %d, want >= 10", n)
        }
}

func TestStoreSeedsSettings(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        if n := mgr.App.Table("settings").Count(); n < 5 {
                t.Errorf("settings seed: got %d, want >= 5", n)
        }
}

func TestStoreCRUD(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        tbl := mgr.App.Table("test_table")
        tbl.Insert(map[string]interface{}{"id": 1, "name": "alice"})
        tbl.Insert(map[string]interface{}{"id": 2, "name": "bob"})
        if got := tbl.Count(); got != 2 {
                t.Errorf("count: got %d, want 2", got)
        }
        tbl.Update(
                func(r map[string]interface{}) bool { return r["id"].(float64) == 1 },
                map[string]interface{}{"name": "alice_updated"},
        )
        r := tbl.First(func(r map[string]interface{}) bool { return r["id"].(float64) == 1 })
        if r["name"] != "alice_updated" {
                t.Errorf("update did not persist; got %v", r["name"])
        }
        n, _ := tbl.Delete(func(r map[string]interface{}) bool { return r["id"].(float64) == 2 })
        if n != 1 {
                t.Errorf("delete: n=%d, want 1", n)
        }
        if got := tbl.Count(); got != 1 {
                t.Errorf("count after delete: got %d, want 1", got)
        }
}

// ─────────────── audit ───────────────

func TestAuditLogWrites(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        al.Log("test_module", "test_action", "tester", map[string]string{"k": "v"}, "normal")
        al.Logf("test_module", "fmt_action", "tester", "msg %d", 42)
        if n := mgr.App.Table("audit_log").Count(); n != 2 {
                t.Errorf("audit_log: got %d rows, want 2", n)
        }
}

// ─────────────── legal ───────────────

func TestLegalSearch(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := legal.New(mgr, al)

        docs := svc.Search("326/2016", "", 50)
        if len(docs) != 1 {
                t.Errorf("search 326/2016: got %d docs, want 1", len(docs))
        }
        if docs[0].TrichYeu == "" {
                t.Errorf("search returned doc with empty trich_yeu")
        }

        docs = svc.Search("án phí", "", 50)
        if len(docs) < 1 {
                t.Errorf("search 'án phí': got %d, want >= 1", len(docs))
        }

        docs = svc.Search("", "hinh_su", 100)
        if len(docs) < 2 {
                t.Errorf("list hinh_su: got %d, want >= 2", len(docs))
        }
}

func TestLegalGetByID(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := legal.New(mgr, al)
        d := svc.GetByID(5)
        if d == nil {
                t.Fatalf("GetByID(5) returned nil")
        }
        if d.SoHieu != "326/2016/UBTVQH14" {
                t.Errorf("GetByID(5) wrong; got %s", d.SoHieu)
        }
}

func TestLegalStats(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := legal.New(mgr, al)
        stats := svc.Stats()
        if stats["hinh_su"] < 2 {
                t.Errorf("stats hinh_su: got %d, want >= 2", stats["hinh_su"])
        }
}

func TestLegalListBieuMau(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := legal.New(mgr, al)
        all := svc.ListBieuMau("", "")
        if len(all) < 5 {
                t.Errorf("ListBieuMau all: got %d, want >= 5", len(all))
        }
        dsOnly := svc.ListBieuMau("dan_su", "")
        for _, b := range dsOnly {
                if b.LinhVuc != "dan_su" {
                        t.Errorf("filter leaked: %v", b.LinhVuc)
                }
        }
}

// ─────────────── spell ───────────────

func TestSpellCheck(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := spell.New(mgr, al)

        text := "ông thi han an  truong hop nhung  cua,,  a^ a( o^"
        res := svc.Check(text, "tester")

        if res.TotalCount == 0 {
                t.Errorf("spell check: no issues found")
        }
        if res.CountByCat["telex_typo"] == 0 {
                t.Errorf("expected telex_typo > 0; got %v", res.CountByCat)
        }
        if res.CountByCat["spacing"] == 0 {
                t.Errorf("expected spacing > 0; got %v", res.CountByCat)
        }
        if res.CountByCat["punct"] == 0 {
                t.Errorf("expected punct > 0; got %v", res.CountByCat)
        }
        if res.CountByCat["term_norm"] == 0 {
                t.Errorf("expected term_norm > 0; got %v", res.CountByCat)
        }
        if !strings.Contains(res.MarkedText, "«") {
                t.Errorf("marked text missing markers")
        }
}

// ─────────────── PII ───────────────

func TestPIIMask(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := pii.New(al)

        text := "CCCD 012345678901, phone 0912345678, email test@gmail.com, DOB 15/03/1990"
        res := svc.Mask(text, pii.PIIOptions{ReplaceMode: "mask"}, "tester")

        if res.TotalCount < 3 {
                t.Errorf("PII: expected >= 3 matches, got %d", res.TotalCount)
        }
        if strings.Contains(res.Text, "012345678901") {
                t.Errorf("CCCD not masked")
        }
        if strings.Contains(res.Text, "test@gmail.com") {
                t.Errorf("email not masked")
        }
}

// ─────────────── casefile ───────────────

func TestCaseCreateListUpdate(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := casefile.New(mgr, al)

        c := casefile.Case{
                CaseNumber:   "QĐ 12/2026/TLPT-DS",
                CaseType:     "DS",
                Stage:        "thu_ly",
                AssignedKSV:  "Nguyễn Văn A",
                ReceivedDate: "2026-10-30",
                Deadline:     "2026-12-30",
                Parties:      []string{"Nguyễn B", "Công ty X"},
        }
        created, err := svc.Create(c, "tester")
        if err != nil {
                t.Fatalf("create: %v", err)
        }
        if created.ID == "" {
                t.Errorf("create did not assign ID")
        }

        got := svc.Get(created.ID)
        if got == nil {
                t.Fatalf("Get returned nil")
        }
        if got.CaseNumber != "QĐ 12/2026/TLPT-DS" {
                t.Errorf("Get wrong case_number: %s", got.CaseNumber)
        }
        if len(got.Parties) != 2 {
                t.Errorf("parties not persisted: got %d", len(got.Parties))
        }

        list := svc.List("DS", "", 100)
        if len(list) != 1 {
                t.Errorf("List: got %d, want 1", len(list))
        }

        if err := svc.SetStatus(created.ID, "dong", "tester"); err != nil {
                t.Fatalf("SetStatus: %v", err)
        }
        got = svc.Get(created.ID)
        if got.Status != "dong" {
                t.Errorf("SetStatus did not persist; got %s", got.Status)
        }
}

// ─────────────── evidence ───────────────

func TestEvidenceAddListTree(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        caseSvc := casefile.New(mgr, al)
        evSvc := evidence.New(mgr, al)

        c, _ := caseSvc.Create(casefile.Case{
                CaseNumber: "QĐ 55/2026",
                CaseType:   "HS",
                Stage:      "dieu_tra",
        }, "tester")

        evs := []evidence.Evidence{
                {CaseID: c.ID, Label: "CT-01", LabelColor: "#ff0000", FilePath: "/tmp/ct1.pdf", Caption: "Bản khai"},
                {CaseID: c.ID, Label: "CT-02", LabelColor: "#ff0000", FilePath: "/tmp/ct2.pdf", Caption: "Ảnh hiện trường"},
                {CaseID: c.ID, Label: "TL-01", LabelColor: "#00ff00", FilePath: "/tmp/tl1.docx", Caption: "Tờ trình"},
        }
        // Add returns a *Evidence with the assigned ID — must capture it
        createdIDs := make([]string, len(evs))
        for i, e := range evs {
                created, err := evSvc.Add(e, "tester")
                if err != nil {
                        t.Fatalf("Add: %v", err)
                }
                createdIDs[i] = created.ID
        }

        all := evSvc.ListByCase(c.ID)
        if len(all) != 3 {
                t.Errorf("ListByCase: got %d, want 3", len(all))
        }

        tree := evSvc.TreeByCase(c.ID)
        if len(tree) != 2 {
                t.Errorf("Tree: got %d categories, want 2 (CT + TL)", len(tree))
        }
        var ctFound bool
        for _, node := range tree {
                if node.Category == "CT" && len(node.Items) == 2 {
                        ctFound = true
                }
        }
        if !ctFound {
                t.Errorf("Tree missing CT category with 2 items")
        }

        // Delete using the captured ID
        if err := evSvc.Delete(createdIDs[0], "tester"); err != nil {
                t.Fatalf("Delete: %v", err)
        }
        if all = evSvc.ListByCase(c.ID); len(all) != 2 {
                t.Errorf("after delete: got %d, want 2", len(all))
        }
}

// ─────────────── docs ───────────────

func TestDocsListTemplates(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := docs.New(mgr, al)
        if all := svc.ListTemplates("", ""); len(all) < 5 {
                t.Errorf("ListTemplates: got %d, want >= 5", len(all))
        }
}

func TestDocsFill(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := docs.New(mgr, al)

        r, err := svc.Fill(docs.FillRequest{
                TemplateID: 3,
                Fields: map[string]interface{}{
                        "case_number": "QĐ 12/2026/TLPT-DS",
                        "ksv":         "Nguyễn Văn A",
                },
                OutputFmt: "html",
        }, "tester")
        if err != nil {
                t.Fatalf("Fill: %v", err)
        }
        if !strings.Contains(r.HTML, "QĐ 12/2026/TLPT-DS") {
                t.Errorf("HTML missing filled case_number")
        }
        if !strings.Contains(r.HTML, "Nguyễn Văn A") {
                t.Errorf("HTML missing filled KSV")
        }
        if r.OutputPath == "" {
                t.Errorf("OutputPath empty")
        }
        if _, err := os.Stat(r.OutputPath); err != nil {
                t.Errorf("output file not written: %v", err)
        }
}

// ─────────────── file (rename) ───────────────

func TestFileRenameDryRun(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := file.New(al)

        tmpDir := filepath.Join(os.TempDir(), "kst_file_test")
        os.MkdirAll(tmpDir, 0755)
        defer os.RemoveAll(tmpDir)
        for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
                os.WriteFile(filepath.Join(tmpDir, name), []byte("test"), 0644)
        }

        res := svc.Apply(file.RenameSpec{
                Dir:            tmpDir,
                Mode:           "number",
                StartIndex:     1,
                PadWidth:       3,
                SkipExtensions: true,
                DryRun:         true,
                SortBy:         "name",
        }, "tester")
        if res.Total != 3 {
                t.Errorf("rename total: got %d, want 3", res.Total)
        }
        if res.Renamed != 0 {
                t.Errorf("dry run should not rename; got %d", res.Renamed)
        }
        for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
                if _, err := os.Stat(filepath.Join(tmpDir, name)); err != nil {
                        t.Errorf("dry run modified files: %s missing", name)
                }
        }
}

// ─────────────── OCR (skip if tesseract not installed) ───────────────

func TestOCRAvailability(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := ocr.New(mgr, al, "/tmp")
        if !svc.IsAvailable() {
                t.Skip("tesseract not installed — skipping OCR runtime test")
        }
        langs := svc.AvailableLanguages()
        if len(langs) == 0 {
                t.Errorf("AvailableLanguages empty")
        }
        t.Logf("tesseract langs: %v", langs)
}

// ─────────────── calc history persistence ───────────────

func TestCalcHistoryPersisted(t *testing.T) {
        _, mgr, cleanup := setup(t)
        defer cleanup()
        al := audit.New(mgr.App)
        svc := calc.New(mgr, al)

        svc.AnPhiHS(calc.AnPhiHSInput{LoaiAn: "so_tham", SoNguoiBiXet: 3}, "tester")
        svc.AnPhiDS(calc.AnPhiDSInput{LoaiAn: "so_tham", GiaTranTuyen: 100_000_000}, "tester")
        svc.Lai(calc.LaiInput{Von: 1_000_000, LaiSuatNam: 9, TuNgay: "2024-01-01", DenNgay: "2024-12-31"}, "tester")

        hist := svc.History(100)
        if len(hist) != 3 {
                t.Errorf("history: got %d, want 3", len(hist))
        }
        for _, h := range hist {
                if h.Module == "" || h.Inputs == "" || h.Outputs == "" {
                        t.Errorf("history entry missing fields: %+v", h)
                }
        }
}
