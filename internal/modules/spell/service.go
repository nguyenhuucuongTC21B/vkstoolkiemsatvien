// Package spell checks Vietnamese spelling in Word/text documents.
// Pure-rule based — no network, no AI. Detection uses:
//   - common typo dictionary (regex + literal replacements)
//   - Vietnamese-specific patterns: "a^" -> "â", "a(" -> "ă", "u*" -> "ư",
//     "o^" -> "ô", "o*" -> "ơ", "e^" -> "ê", "d/" -> "đ",
//     "sequential whitespace", "double punctuation"
//   - legal-term normalization ("thi han an" -> "thi hành án")
//
// For .docx files: extract document.xml, mark errors with yellow
// highlight, save as .docx; original file never modified.
package spell

import (
        "bytes"
        "encoding/json"
        "fmt"
        "os"
        "os/exec"
        "path/filepath"
        "regexp"
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

// SpellIssue một lỗi phát hiện được.
type SpellIssue struct {
        Type        string `json:"type"`        // "typo" | "spacing" | "punct" | "term_norm" | "telex_typo"
        Original    string `json:"original"`
        Suggestion  string `json:"suggestion"`
        Start       int    `json:"start"`
        End         int    `json:"end"`
}

// SpellResult kết quả rà soát.
type SpellResult struct {
        Issues     []SpellIssue `json:"issues"`
        CountByCat map[string]int `json:"countByCat"`
        TotalCount int          `json:"totalCount"`
        MarkedText string       `json:"markedText"` // text with [[...]] markers around issues
        OutputPath string       `json:"outputPath"`
}

// Check scans plain text and returns issues.
func (s *Service) Check(text string, actor string) SpellResult {
        issues := []SpellIssue{}
        countByCat := map[string]int{}

        // 1. Telex-style unconverted sequences: a^, a(, u*, o^, o*, e^, d/
        telexPatterns := []struct {
                pat  string
                repl string
        }{
                {`a\^`, "â"},
                {`a\(`, "ă"},
                {`u\*`, "ư"},
                {`o\^`, "ô"},
                {`o\*`, "ơ"},
                {`e\^`, "ê"},
                {`d/`, "đ"},
                {`A\^`, "Â"},
                {`A\(`, "Ă"},
                {`U\*`, "Ư"},
                {`O\^`, "Ô"},
                {`O\*`, "Ơ"},
                {`E\^`, "Ê"},
                {`D/`, "Đ"},
        }
        for _, tp := range telexPatterns {
                re := regexp.MustCompile(tp.pat)
                for _, m := range re.FindAllStringIndex(text, -1) {
                        issues = append(issues, SpellIssue{
                                Type: "telex_typo", Original: text[m[0]:m[1]],
                                Suggestion: tp.repl, Start: m[0], End: m[1],
                        })
                        countByCat["telex_typo"]++
                }
        }

        // 2. Double whitespace
        reWS := regexp.MustCompile(`[ \t]{2,}`)
        for _, m := range reWS.FindAllStringIndex(text, -1) {
                issues = append(issues, SpellIssue{
                        Type: "spacing", Original: text[m[0]:m[1]],
                        Suggestion: " ", Start: m[0], End: m[1],
                })
                countByCat["spacing"]++
        }

        // 3. Leading/trailing whitespace per line
        reLeadWS := regexp.MustCompile(`(?m)^[ \t]+`)
        for _, m := range reLeadWS.FindAllStringIndex(text, -1) {
                // Skip if it's a meaningful indent (4 spaces). Only flag if 1-3 spaces
                // — likely a stray.
                seg := text[m[0]:m[1]]
                if len(seg) >= 1 && len(seg) <= 3 {
                        issues = append(issues, SpellIssue{
                                Type: "spacing", Original: seg,
                                Suggestion: "", Start: m[0], End: m[1],
                        })
                        countByCat["spacing"]++
                }
        }

        // 4. Double punctuation: ",,", ".." (more than 2), "!!", "??"
        reDP := regexp.MustCompile(`([,.;:!?,])\1{1,}`)
        for _, m := range reDP.FindAllStringIndex(text, -1) {
                issues = append(issues, SpellIssue{
                        Type: "punct", Original: text[m[0]:m[1]],
                        Suggestion: string(text[m[0]]), Start: m[0], End: m[1],
                })
                countByCat["punct"]++
        }

        // 5. Space before punctuation: " ," " ." " ;"
        reSBP := regexp.MustCompile(`\s+([,.;:!?])`)
        for _, m := range reSBP.FindAllStringSubmatchIndex(text, -1) {
                issues = append(issues, SpellIssue{
                        Type: "punct", Original: text[m[0]:m[1]],
                        Suggestion: text[m[2]:m[3]], Start: m[0], End: m[1],
                })
                countByCat["punct"]++
        }

        // 6. Legal-term normalization (lowercase source, keep case after replace)
        termRules := s.loadRules("term_norm")
        for _, rule := range termRules {
                re, err := regexp.Compile(`(?i)` + regexp.QuoteMeta(rule.pattern))
                if err != nil {
                        continue
                }
                for _, m := range re.FindAllStringIndex(text, -1) {
                        issues = append(issues, SpellIssue{
                                Type: "term_norm", Original: text[m[0]:m[1]],
                                Suggestion: rule.replacement, Start: m[0], End: m[1],
                        })
                        countByCat["term_norm"]++
                }
        }

        // 7. Literal typo rules
        literalRules := s.loadRules("literal")
        for _, rule := range literalRules {
                idx := strings.Index(text, rule.pattern)
                for idx >= 0 {
                        issues = append(issues, SpellIssue{
                                Type: "typo", Original: rule.pattern,
                                Suggestion: rule.replacement, Start: idx, End: idx + len(rule.pattern),
                        })
                        countByCat["typo"]++
                        next := idx + len(rule.pattern)
                        if next >= len(text) {
                                break
                        }
                        idx = strings.Index(text[next:], rule.pattern)
                        if idx < 0 {
                                break
                        }
                        idx += next
                }
        }

        // Build marked text — wrap each issue with <<...>>
        out := text
        // Apply in reverse to preserve indices
        sorted := make([]SpellIssue, len(issues))
        copy(sorted, issues)
        for i := 0; i < len(sorted); i++ {
                for j := i + 1; j < len(sorted); j++ {
                        if sorted[j].Start > sorted[i].Start {
                                sorted[i], sorted[j] = sorted[j], sorted[i]
                        }
                }
        }
        for _, iss := range sorted {
                out = out[:iss.Start] + "«" + iss.Original + "→" + iss.Suggestion + "»" + out[iss.End:]
        }

        res := SpellResult{
                Issues: issues, CountByCat: countByCat, TotalCount: len(issues),
                MarkedText: out,
        }
        if s.audit != nil {
                d, _ := json.Marshal(map[string]interface{}{"count": res.TotalCount, "cats": countByCat})
                s.audit.Log("spell", "check", actor, string(d), "normal")
        }
        return res
}

// CheckFile checks a .txt/.md/.docx file and writes the marked output.
func (s *Service) CheckFile(inputPath, actor string) (SpellResult, error) {
        ext := strings.ToLower(filepath.Ext(inputPath))
        if ext == ".docx" {
                return s.checkDocx(inputPath, actor)
        }
        data, err := os.ReadFile(inputPath)
        if err != nil {
                return SpellResult{}, err
        }
        res := s.Check(string(data), actor)
        outPath := inputPath + ".marked.txt"
        _ = os.WriteFile(outPath, []byte(res.MarkedText), 0644)
        res.OutputPath = outPath
        return res, nil
}

// checkDocx is a placeholder for the docx workflow — MVP just reads
// the raw text via unzip+grep on document.xml. A follow-up module
// uses the docx skill for proper round-trip editing.
func (s *Service) checkDocx(inputPath, actor string) (SpellResult, error) {
        // Minimal: try `unzip -p` to extract document.xml text
        xml, err := extractDocxText(inputPath)
        if err != nil {
                return SpellResult{}, fmt.Errorf("docx extract: %w", err)
        }
        res := s.Check(xml, actor)
        res.OutputPath = inputPath + ".marked.txt"
        _ = os.WriteFile(res.OutputPath, []byte(res.MarkedText), 0644)
        return res, nil
}

// extractDocxText pulls the visible text out of a .docx file by
// unzipping word/document.xml and stripping XML tags. MVP-grade —
// for round-trip editing with yellow highlighting, the dedicated
// docx skill (separate package) must be used.
func extractDocxText(path string) (string, error) {
        // Try `unzip -p` to extract document.xml
        cmd := exec.Command("unzip", "-p", path, "word/document.xml")
        var out bytes.Buffer
        cmd.Stdout = &out
        if err := cmd.Run(); err != nil {
                return "", fmt.Errorf("unzip: %w (cài unzip để đọc .docx)", err)
        }
        xml := out.String()
        // Convert <w:p> to newlines, strip all other tags
        xml = strings.ReplaceAll(xml, "</w:p>", "\n")
        re := regexp.MustCompile(`<[^>]+>`)
        text := re.ReplaceAllString(xml, "")
        // Unescape basic entities
        text = strings.ReplaceAll(text, "&amp;", "&")
        text = strings.ReplaceAll(text, "&lt;", "<")
        text = strings.ReplaceAll(text, "&gt;", ">")
        text = strings.ReplaceAll(text, "&quot;", "\"")
        text = strings.ReplaceAll(text, "&apos;", "'")
        return text, nil
}

// ───── rules from DB ─────

type spellRule struct {
        pattern     string
        replacement string
}

func (s *Service) loadRules(ruleType string) []spellRule {
        if s.db == nil || s.db.App == nil {
                return nil
        }
        rows := s.db.App.Table("spell_rules").Filter(func(r store.Row) bool {
                return store.FieldStr(r, "rule_type") == ruleType && store.FieldBool(r, "enabled")
        })
        out := make([]spellRule, 0, len(rows))
        for _, r := range rows {
                out = append(out, spellRule{
                        pattern:     store.FieldStr(r, "pattern"),
                        replacement: store.FieldStr(r, "replacement"),
                })
        }
        return out
}
