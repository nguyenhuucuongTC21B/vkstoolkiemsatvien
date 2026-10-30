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
        MarkedText string       `json:"markedText"` // text with «old→new» markers around issues
        OutputPath string       `json:"outputPath"`
        FixedText  string       `json:"fixedText"`  // NEW v0.2.0: text with all fixes applied
        Stats      SpellStats   `json:"stats"`
}

// SpellStats — thống kê về văn bản đã sửa.
type SpellStats struct {
        OriginalChars int `json:"originalChars"`
        FixedChars    int `json:"fixedChars"`
        IssuesFixed   int `json:"issuesFixed"`
        IssuesSkipped int `json:"issuesSkipped"`
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

        // 4. Double punctuation: ",,", ".." (3+), "!!", "??", ";;"
        // RE2 (Go regex) không hỗ trợ backreference \1, nên liệt kê từng cặp.
        reDP := regexp.MustCompile(`(?:,,|\.\.\.+|;;|!!|\?\?)`)
        for _, m := range reDP.FindAllStringIndex(text, -1) {
                orig := text[m[0]:m[1]]
                // Suggestion: rút gọn về 1 ký tự duy nhất
                suggest := string(orig[0])
                issues = append(issues, SpellIssue{
                        Type: "punct", Original: orig,
                        Suggestion: suggest, Start: m[0], End: m[1],
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

        // Sort issues by start position (ascending)
        for i := 1; i < len(issues); i++ {
                for j := i; j > 0 && issues[j].Start < issues[j-1].Start; j-- {
                        issues[j], issues[j-1] = issues[j-1], issues[j]
                }
        }
        // Drop overlapping issues (keep first)
        if len(issues) > 1 {
                filtered := []SpellIssue{issues[0]}
                lastEnd := issues[0].End
                for i := 1; i < len(issues); i++ {
                        if issues[i].Start >= lastEnd {
                                filtered = append(filtered, issues[i])
                                lastEnd = issues[i].End
                        }
                }
                issues = filtered
        }

        // Build marked text — work backwards to preserve indices
        marked := text
        for i := len(issues) - 1; i >= 0; i-- {
                iss := issues[i]
                marked = marked[:iss.Start] + "«" + iss.Original + "→" + iss.Suggestion + "»" + marked[iss.End:]
        }

        // Build fixed text — apply all suggestions
        fixed := text
        for i := len(issues) - 1; i >= 0; i-- {
                iss := issues[i]
                fixed = fixed[:iss.Start] + iss.Suggestion + fixed[iss.End:]
        }

        res := SpellResult{
                Issues: issues, CountByCat: countByCat, TotalCount: len(issues),
                MarkedText: marked,
                FixedText:  fixed,
                Stats: SpellStats{
                        OriginalChars: len(text),
                        FixedChars:    len(fixed),
                        IssuesFixed:   len(issues),
                        IssuesSkipped: 0,
                },
        }
        if s.audit != nil {
                d, _ := json.Marshal(map[string]interface{}{
                        "count": res.TotalCount, "cats": countByCat, "fixed_chars": len(fixed),
                })
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
