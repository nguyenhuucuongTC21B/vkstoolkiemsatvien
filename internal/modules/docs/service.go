// Package docs generates, fills, and exports prosecution documents
// from the bundled templates (Quyết định phân công KSV, Thông báo thụ lý,
// Quyết định khởi kiện, Kháng cáo, v.v.). Templates live in legal.sqlite
// as HTML; placeholders {{...}} are replaced with case data.
package docs

import (
        "encoding/json"
        "fmt"
        "os"
        "path/filepath"
        "regexp"
        "strings"
        "time"

        "kiem_sat_toolkit/internal/audit"
        "kiem_sat_toolkit/internal/db"
        "kiem_sat_toolkit/internal/docxgen"
        "kiem_sat_toolkit/internal/store"
)

type Service struct {
        db    *db.Manager
        audit *audit.Logger
}

func New(dbMgr *db.Manager, auditLog *audit.Logger) *Service {
        return &Service{db: dbMgr, audit: auditLog}
}

// FillRequest — yêu cầu điền một biểu mẫu.
type FillRequest struct {
        TemplateID int                    `json:"templateId"`
        CaseID     string                 `json:"caseId"`
        Fields     map[string]interface{} `json:"fields"`
        OutputFmt  string                 `json:"outputFmt"` // "html" | "md" | "txt"
}

// FillResult — kết quả.
type FillResult struct {
        HTML       string `json:"html"`
        OutputPath string `json:"outputPath"`
        Warnings   []string `json:"warnings"`
}

// Fill loads the template, replaces placeholders, writes to disk.
// Output format can be "html", "md", "txt", or "docx" (real Word file).
func (s *Service) Fill(req FillRequest, actor string) (*FillResult, error) {
        if req.OutputFmt == "" {
                req.OutputFmt = "docx" // default to Word .docx
        }
        tpl, err := s.loadTemplate(req.TemplateID)
        if err != nil {
                return nil, fmt.Errorf("load template: %w", err)
        }

        // Merge with case data if case_id is provided
        fields := req.Fields
        if req.CaseID != "" {
                // Pull case fields lazily — we don't have a direct reference to the
                // case service; the frontend should pre-fill the fields map.
        }

        html := applyPlaceholders(tpl, fields)

        outDir := filepath.Join(s.dataDir(), "exports")
        _ = os.MkdirAll(outDir, 0755)
        ts := time.Now().Format("20060102_150405")

        if req.OutputFmt == "docx" {
                // Generate real .docx file using docxgen
                doc := docxgen.New()
                htmlToDocx(doc, html, fields)
                outPath := filepath.Join(outDir, fmt.Sprintf("doc_%d_%s.docx", req.TemplateID, ts))
                if err := doc.Save(outPath); err != nil {
                        return nil, fmt.Errorf("save docx: %w", err)
                }
                if s.audit != nil {
                        d, _ := json.Marshal(map[string]interface{}{
                                "tpl": req.TemplateID, "case": req.CaseID, "out": outPath, "fmt": "docx",
                        })
                        s.audit.Log("docs", "fill", actor, string(d), "restricted")
                }
                return &FillResult{HTML: html, OutputPath: outPath}, nil
        }

        // Otherwise write plain text/md/html
        outPath := filepath.Join(outDir, fmt.Sprintf("doc_%d_%s.%s", req.TemplateID, ts, req.OutputFmt))
        _ = os.WriteFile(outPath, []byte(html), 0644)

        if s.audit != nil {
                d, _ := json.Marshal(map[string]interface{}{
                        "tpl": req.TemplateID, "case": req.CaseID, "out": outPath,
                })
                s.audit.Log("docs", "fill", actor, string(d), "restricted")
        }
        return &FillResult{HTML: html, OutputPath: outPath}, nil
}

// htmlToDocx converts simple HTML template content to docx document.
// Recognizes: <h1>, <h2>, <h3>, <p>, <ul><li>, <table><tr><td>.
// Plus applies field substitutions already done in html string.
func htmlToDocx(doc *docxgen.Document, html string, fields map[string]interface{}) {
        // Strip newlines and normalize
        html = strings.ReplaceAll(html, "\n", " ")
        html = strings.ReplaceAll(html, "\t", " ")

        // Very simple HTML parser — handle one tag at a time
        tags := regexp.MustCompile(`<[^>]+>`)
        parts := tags.Split(html, -1)
        allMatches := tags.FindAllString(html, -1)

        for i, text := range parts {
                text = strings.TrimSpace(text)
                // Decode basic entities
                text = strings.ReplaceAll(text, "&amp;", "&")
                text = strings.ReplaceAll(text, "&lt;", "<")
                text = strings.ReplaceAll(text, "&gt;", ">")
                text = strings.ReplaceAll(text, "&quot;", `"`)
                text = strings.ReplaceAll(text, "&apos;", "'")
                text = strings.ReplaceAll(text, "&nbsp;", " ")

                var tag string
                if i < len(allMatches) {
                        tag = allMatches[i]
                }

                if text == "" && !strings.HasPrefix(tag, "<br") {
                        continue
                }

                switch {
                case strings.HasPrefix(tag, "<h1"):
                        doc.AddHeading(text, 1)
                case strings.HasPrefix(tag, "<h2"):
                        doc.AddHeading(text, 2)
                case strings.HasPrefix(tag, "<h3"):
                        doc.AddHeading(text, 3)
                case strings.HasPrefix(tag, "<li"):
                        doc.AddListItem(text)
                case strings.HasPrefix(tag, "<br"):
                        if text != "" {
                                doc.AddParagraph(text)
                        } else {
                                doc.AddEmptyLine()
                        }
                default:
                        if text != "" {
                                // Auto-detect center for short title-like lines
                                if len(text) < 80 && !strings.Contains(text, ":") &&
                                        (strings.Contains(strings.ToUpper(text), "QUYẾT ĐỊNH") ||
                                                strings.Contains(strings.ToUpper(text), "THÔNG BÁO") ||
                                                strings.Contains(strings.ToUpper(text), "ĐƠN") ||
                                                strings.Contains(strings.ToUpper(text), "BIÊN BẢN")) {
                                        doc.AddCentered(text, true)
                                } else {
                                        doc.AddParagraph(text)
                                }
                        }
                }
        }
}

// loadTemplate pulls the HTML body of one bieu_mau row.
func (s *Service) loadTemplate(id int) (string, error) {
        r := s.db.Legal.Table("bieu_mau").First(func(r store.Row) bool {
                return store.FieldInt(r, "id") == id
        })
        if r == nil {
                return "", fmt.Errorf("template id=%d not found", id)
        }
        return store.FieldStr(r, "file_html"), nil
}

// applyPlaceholders replaces {{key}} with fields[key].
// Supports {{date}}, {{today}}, {{ksv}}, {{case_number}} out of the box.
var placeholderRE = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

func applyPlaceholders(tpl string, fields map[string]interface{}) string {
        // Add default fields
        if _, ok := fields["today"]; !ok {
                fields["today"] = time.Now().Format("02/01/2006")
        }
        if _, ok := fields["date"]; !ok {
                fields["date"] = time.Now().Format("02/01/2006")
        }

        return placeholderRE.ReplaceAllStringFunc(tpl, func(m string) string {
                key := strings.TrimSpace(m[2 : len(m)-2])
                if v, ok := fields[key]; ok {
                        return fmt.Sprintf("%v", v)
                }
                // Leave the placeholder in place so the user sees what's missing
                return m
        })
}

// ListTemplates exposes the bieu_mau table to the UI.
func (s *Service) ListTemplates(linhVuc, giaiDoan string) []TemplateInfo {
        rows := s.db.Legal.Table("bieu_mau").Filter(func(r store.Row) bool {
                if linhVuc != "" && store.FieldStr(r, "linh_vuc") != linhVuc {
                        return false
                }
                if giaiDoan != "" && store.FieldStr(r, "giai_doan") != giaiDoan {
                        return false
                }
                return true
        })
        out := make([]TemplateInfo, 0, len(rows))
        for _, r := range rows {
                out = append(out, TemplateInfo{
                        ID:       store.FieldInt(r, "id"),
                        MaMau:    store.FieldStr(r, "ma_mau"),
                        TenMau:   store.FieldStr(r, "ten_mau"),
                        LinhVuc:  store.FieldStr(r, "linh_vuc"),
                        GiaiDoan: store.FieldStr(r, "giai_doan"),
                        GhiChu:   store.FieldStr(r, "ghi_chu"),
                })
        }
        return out
}

// TemplateInfo — bản tóm tắt biểu mẫu.
type TemplateInfo struct {
        ID       int    `json:"id"`
        MaMau    string `json:"maMau"`
        TenMau   string `json:"tenMau"`
        LinhVuc string `json:"linhVuc"`
        GiaiDoan string `json:"giaiDoan"`
        GhiChu   string `json:"ghiChu"`
}

func (s *Service) dataDir() string {
        // We don't have direct access to the dataDir string; the app exports
        // an `ExportDir` method via the Wails binding that the frontend uses
        // for absolute paths. For internal use here, we use the user home
        // as a fallback.
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".kiem_sat_toolkit")
}
