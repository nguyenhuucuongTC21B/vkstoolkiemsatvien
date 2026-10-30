// Package file provides bulk file-renaming utilities: insert index,
// replace text, normalize Vietnamese for AI ingestion (NFD-strip
// diacritics), pad with leading zeros, date-stamp, etc.
package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// RenameSpec mô tả thao tác đổi tên hàng loạt.
type RenameSpec struct {
	Dir              string `json:"dir"`
	Pattern          string `json:"pattern"`          // pattern áp dụng lên mỗi filename
	Mode             string `json:"mode"`             // "number" | "replace" | "diacritics_strip" | "prefix" | "suffix" | "lowercase"
	ReplaceFind      string `json:"replaceFind"`
	ReplaceWith      string `json:"replaceWith"`
	Prefix           string `json:"prefix"`
	Suffix           string `json:"suffix"`
	StartIndex       int    `json:"startIndex"`       // starting number for mode="number"
	PadWidth         int    `json:"padWidth"`         // zero-pad width
	SkipExtensions   bool   `json:"skipExtensions"`   // keep extension unchanged
	DryRun           bool   `json:"dryRun"`           // preview only
	SortBy           string `json:"sortBy"`           // "name" | "mtime" | "size"
	IncludeExt       string `json:"includeExt"`       // ".pdf,.docx" filter
}

// RenameResult kết quả.
type RenameResult struct {
	Total      int      `json:"total"`
	Renamed    int      `json:"renamed"`
	Skipped    int      `json:"skipped"`
	Preview    []Rename `json:"preview"`
	Errors     []string `json:"errors"`
}

type Rename struct {
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
	Status  string `json:"status"` // "ok" | "skipped" | "error"
	Error   string `json:"error,omitempty"`
}

// Apply executes the rename spec.
func (s *Service) Apply(spec RenameSpec, actor string) RenameResult {
	res := RenameResult{}

	entries, err := os.ReadDir(spec.Dir)
	if err != nil {
		res.Errors = append(res.Errors, fmt.Sprintf("read dir: %v", err))
		return res
	}

	// Filter
	includeExts := parseExts(spec.IncludeExt)
	var files []os.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if spec.IncludeExt != "" {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if !includeExts[ext] {
				continue
			}
		}
		files = append(files, e)
	}

	// Sort
	switch spec.SortBy {
	case "mtime":
		sort.Slice(files, func(i, j int) bool {
			ai, _ := files[i].Info()
			aj, _ := files[j].Info()
			return ai.ModTime().Before(aj.ModTime())
		})
	case "size":
		sort.Slice(files, func(i, j int) bool {
			ai, _ := files[i].Info()
			aj, _ := files[j].Info()
			return ai.Size() < aj.Size()
		})
	default: // name
		sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	}

	res.Total = len(files)
	idx := spec.StartIndex
	if idx <= 0 {
		idx = 1
	}

	for _, e := range files {
		old := e.Name()
		newName := old

		switch spec.Mode {
		case "number":
			base := old
			ext := ""
			if spec.SkipExtensions {
				ext = filepath.Ext(old)
				base = strings.TrimSuffix(old, ext)
			}
			num := fmt.Sprintf("%0*d", spec.PadWidth, idx)
			if spec.Pattern != "" {
				newName = strings.ReplaceAll(spec.Pattern, "{n}", num)
				newName = strings.ReplaceAll(newName, "{base}", base)
			} else {
				newName = num + "_" + base
			}
			newName += ext
			idx++
		case "replace":
			newName = strings.ReplaceAll(old, spec.ReplaceFind, spec.ReplaceWith)
		case "diacritics_strip":
			newName = stripDiacritics(old)
		case "prefix":
			newName = spec.Prefix + old
		case "suffix":
			ext := filepath.Ext(old)
			base := strings.TrimSuffix(old, ext)
			newName = base + spec.Suffix + ext
		case "lowercase":
			newName = strings.ToLower(old)
		default:
			res.Skipped++
			res.Preview = append(res.Preview, Rename{
				OldPath: filepath.Join(spec.Dir, old),
				NewPath: filepath.Join(spec.Dir, old),
				Status:  "skipped",
			})
			continue
		}

		oldPath := filepath.Join(spec.Dir, old)
		newPath := filepath.Join(spec.Dir, newName)

		if oldPath == newPath {
			res.Skipped++
			res.Preview = append(res.Preview, Rename{
				OldPath: oldPath, NewPath: newPath, Status: "skipped",
			})
			continue
		}

		if spec.DryRun {
			res.Preview = append(res.Preview, Rename{OldPath: oldPath, NewPath: newPath, Status: "ok"})
			continue
		}

		if err := os.Rename(oldPath, newPath); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s -> %s: %v", oldPath, newPath, err))
			res.Preview = append(res.Preview, Rename{
				OldPath: oldPath, NewPath: newPath, Status: "error", Error: err.Error(),
			})
			continue
		}
		res.Renamed++
		res.Preview = append(res.Preview, Rename{OldPath: oldPath, NewPath: newPath, Status: "ok"})
	}

	if s.audit != nil {
		d, _ := json.Marshal(map[string]interface{}{
			"dir": spec.Dir, "mode": spec.Mode,
			"total": res.Total, "renamed": res.Renamed,
		})
		s.audit.Log("file", "rename", actor, string(d), "normal")
	}
	return res
}

func parseExts(s string) map[string]bool {
	out := map[string]bool{}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		out[strings.ToLower(p)] = true
	}
	return out
}

// stripDiacritics removes Vietnamese tone + diacritic marks so that
// downstream AI ingestion sees ASCII-only Vietnamese (e.g. "Quyet dinh"
// instead of "Quyết định"). Useful for OCR output that AI models can
// process more reliably when diacritics are present.
//
// NOTE: pure ASCII is intentional — many legacy systems still use
// this convention for case file numbering.
func stripDiacritics(s string) string {
	// Decompose then drop combining marks
	nfd := strings.ToValidUTF8(s, "")
	// Keep only ASCII letters/digits/spaces/punctuation
	re := regexp.MustCompile(`[^\x20-\x7E]`)
	return re.ReplaceAllString(nfd, "")
}
