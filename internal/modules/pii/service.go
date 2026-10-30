// Package pii masks personal information (PII) in Vietnamese legal
// documents before sharing or sending to AI services. Detection is
// regex-based — no network, no model — to remain 100% offline.
//
// Recognized PII categories (per Nghị định 13/2023/NĐ-CP on personal
// data protection):
//   - Họ tên người (heuristic)
//   - CCCD/CMND (12 / 9 digits)
//   - Số hộ chiếu
//   - Số điện thoại (Vietnamese mobile patterns)
//   - Email
//   - Ngày tháng năm sinh (dd/mm/yyyy, dd-mm-yyyy)
//   - Địa chỉ (số + tên đường + phường/xã, quận/huyện, tỉnh/TP)
//   - Số tài khoản ngân hàng (9-16 digits near "STK" / "tài khoản")
//   - Biển số xe (Vietnamese plate format)
//   - Mã số thuế (10 / 14 digits)
package pii

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"kiem_sat_toolkit/internal/audit"
)

type Service struct {
	audit *audit.Logger
}

func New(auditLog *audit.Logger) *Service {
	return &Service{audit: auditLog}
}

// PIIOptions cấu hình che thông tin.
type PIIOptions struct {
	ReplaceMode string   `json:"replaceMode"` // "mask" | "token" | "remove"
	KeepFirst   bool     `json:"keepFirst"`   // giữ ký tự đầu (Nguyễn V*** A**)
	Categories  []string `json:"categories"`  // subset to mask; empty = all
}

// PIIMatch một vị trí PII phát hiện được.
type PIIMatch struct {
	Category string `json:"category"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Value    string `json:"value"` // redacted form (already replaced)
}

// MaskResult kết quả che PII trên một văn bản.
type MaskResult struct {
	Text       string         `json:"text"`
	Matches    []PIIMatch     `json:"matches"`
	CountByCat map[string]int `json:"countByCat"`
	TotalCount int            `json:"totalCount"`
	Warnings   []string       `json:"warnings"`
}

// rawMatch là kết quả khớp pattern chưa qua lọc overlap.
type rawMatch struct {
	start, end int
	category   string
	original   string
}

// Mask scans text and returns a redacted copy + match list.
func (s *Service) Mask(text string, opts PIIOptions, actor string) MaskResult {
	if opts.ReplaceMode == "" {
		opts.ReplaceMode = "mask"
	}
	patterns := allPatterns()
	if len(opts.Categories) > 0 {
		want := map[string]bool{}
		for _, c := range opts.Categories {
			want[c] = true
		}
		var filtered []pattern
		for _, p := range patterns {
			if want[p.category] {
				filtered = append(filtered, p)
			}
		}
		patterns = filtered
	}

	// Collect all matches
	var raws []rawMatch
	for _, p := range patterns {
		for _, m := range p.re.FindAllStringIndex(text, -1) {
			if len(m) < 2 {
				continue
			}
			raws = append(raws, rawMatch{
				start:    m[0],
				end:      m[1],
				category: p.category,
				original: text[m[0]:m[1]],
			})
		}
	}

	// Sort by start
	sort.Slice(raws, func(i, j int) bool { return raws[i].start < raws[j].start })

	// Filter overlaps
	accepted := make([]rawMatch, 0, len(raws))
	lastEnd := -1
	for _, r := range raws {
		if r.start < lastEnd {
			continue
		}
		accepted = append(accepted, r)
		lastEnd = r.end
	}

	// Build replaced text (work backwards to keep indices valid)
	out := text
	matches := make([]PIIMatch, 0, len(accepted))
	countByCat := map[string]int{}
	for i := len(accepted) - 1; i >= 0; i-- {
		r := accepted[i]
		var replacement string
		switch opts.ReplaceMode {
		case "remove":
			replacement = ""
		case "token":
			replacement = fmt.Sprintf("[%s]", strings.ToUpper(r.category))
		default: // mask
			replacement = maskValue(r.original, opts.KeepFirst)
		}
		out = out[:r.start] + replacement + out[r.end:]
		matches = append(matches, PIIMatch{
			Category: r.category,
			Start:    r.start,
			End:      r.start + len(replacement),
			Value:    replacement,
		})
		countByCat[r.category]++
	}
	// Reverse matches so they're in document order
	for i, j := 0, len(matches)-1; i < j; i, j = i+1, j-1 {
		matches[i], matches[j] = matches[j], matches[i]
	}

	res := MaskResult{
		Text:       out,
		Matches:    matches,
		CountByCat: countByCat,
		TotalCount: len(matches),
	}

	if s.audit != nil {
		d, _ := json.Marshal(map[string]interface{}{
			"categories": countByCat, "total": res.TotalCount, "mode": opts.ReplaceMode,
		})
		s.audit.Log("pii", "mask", actor, string(d), "restricted")
	}
	return res
}

// MaskFile reads, masks, and writes the result to <name>.masked.txt.
// Returns the path of the masked file.
func (s *Service) MaskFile(inputPath string, opts PIIOptions, actor string) (string, MaskResult, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return "", MaskResult{}, err
	}
	res := s.Mask(string(data), opts, actor)
	outPath := inputPath + ".masked.txt"
	if err := os.WriteFile(outPath, []byte(res.Text), 0644); err != nil {
		return "", res, err
	}
	return outPath, res, nil
}

func maskValue(orig string, keepFirst bool) string {
	chars := []rune(orig)
	if len(chars) == 0 {
		return ""
	}
	if !keepFirst {
		return strings.Repeat("*", len(chars))
	}
	if len(chars) == 1 {
		return "*"
	}
	return string(chars[0]) + strings.Repeat("*", len(chars)-1)
}

// ─────────────── patterns ───────────────

type pattern struct {
	category string
	re       *regexp.Regexp
}

func allPatterns() []pattern {
	return []pattern{
		// CCCD 12 digits
		{"cccd", regexp.MustCompile(`\b\d{12}\b`)},
		// CMND 9 digits
		{"cmnd", regexp.MustCompile(`\b\d{9}\b`)},
		// Passport VN + 1 letter + 7 digits
		{"passport", regexp.MustCompile(`\b[A-Z]\d{7}\b`)},
		// Vietnamese mobile: 0xx-xxxx-xxx (10 digits starting 0)
		{"phone", regexp.MustCompile(`\b0\d{9,10}\b`)},
		// Email
		{"email", regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)},
		// Date dd/mm/yyyy or dd-mm-yyyy
		{"dob", regexp.MustCompile(`\b\d{1,2}[/-]\d{1,2}[/-]\d{2,4}\b`)},
		// Bank account (mention keyword + digits 9-16)
		{"bank_account", regexp.MustCompile(`(?i)(?:stk|tài khoản|số tk)[:\s]+(\d{9,16})\b`)},
		// License plate Vietnamese
		{"plate", regexp.MustCompile(`\b\d{2}[A-Z]{1,2}-\d{3,5}\b`)},
		// Tax code (MST) 10 or 14 digits
		{"tax_code", regexp.MustCompile(`(?i)(?:mst|mã số thuế)[:\s]+(\d{10}(?:\d{4})?)\b`)},
		// Address heuristic: số + (đường/phố/khu/phường/xã...)
		{"address", regexp.MustCompile(`(?i)\b(?:số|ngõ|ngách|khu|tổ|ấp|thôn)\s+\d+[A-Za-z]?(?:\s+\S+){0,8}?\s+(?:đường|phố|phường|xã|thị trấn|quận|huyện|thành phố|tỉnh)\b`)},
		// Vietnamese full name heuristic
		{"name", regexp.MustCompile(`(?i)\b(?:ông|bà|anh|chị|bé|em|giám đốc|trưởng phòng|kiểm sát viên|thẩm phán|luật sư)\s+([A-ZÀ-Ỵ][a-zà-ỹ]+(?:\s+[A-ZÀ-Ỵ][a-zà-ỹ]+){1,4})\b`)},
	}
}
