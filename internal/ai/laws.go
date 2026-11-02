// laws.go — load laws.json làm ngữ cảnh cho Module 5 LEGAL ADVISOR.
//
// laws.json là file cấu trúc cứng (hard-coded) chứa trích đoạn BLHS
// với chương/điều/tội/khung hình phạt. AI dựa vào đây để đề xuất điều áp dụng.
//
// File được embed vào binary qua go:embed (offline, không cần network).
package ai

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed data/laws.json
var lawsFS embed.FS

// LawsData — schema của laws.json.
type LawsData struct {
	Version string     `json:"version"`
	Updated string     `json:"updated"`
	Chuong  []Chuong   `json:"chuong"`
}

// Chuong — một chương BLHS.
type Chuong struct {
	So      int    `json:"so"`
	Ten     string `json:"ten"`
	Dieu    []Dieu `json:"dieu"`
}

// Dieu — một điều luật.
type Dieu struct {
	So             int      `json:"so"`
	Ten            string   `json:"ten"`
	Khoan          []Khoan  `json:"khoan"`
}

// Khoan — một khoản trong điều.
type Khoan struct {
	So        int    `json:"so"`
	NoiDung   string `json:"noiDung"`
	KhungPhat string `json:"khungPhat"` // "06 tháng - 03 năm tù"
}

var (
	lawsCache     *LawsData
	lawsCacheJSON string
	lawsOnce      sync.Once
	lawsLoadErr   error
)

// LoadLaws — load + cache laws.json (1 lần duy nhất, thread-safe).
func LoadLaws() (*LawsData, error) {
	lawsOnce.Do(func() {
		data, err := lawsFS.ReadFile("data/laws.json")
		if err != nil {
			lawsLoadErr = fmt.Errorf("read laws.json: %w", err)
			return
		}
		var laws LawsData
		if err := json.Unmarshal(data, &laws); err != nil {
			lawsLoadErr = fmt.Errorf("parse laws.json: %w", err)
			return
		}
		lawsCache = &laws
		lawsCacheJSON = string(data)
	})
	return lawsCache, lawsLoadErr
}

// cachedLawsJSON — trả JSON string để inject vào prompt AI (Module 5).
func cachedLawsJSON() string {
	_, _ = LoadLaws()
	return lawsCacheJSON
}

// SearchLaws — tra cứu điều luật theo từ khóa (cho UI tra cứu trực tiếp).
func SearchLaws(keyword string, limit int) []Dieu {
	if limit <= 0 {
		limit = 20
	}
	laws, err := LoadLaws()
	if err != nil {
		return nil
	}
	kw := normalizeVietnamese(keyword)
	var results []Dieu
	for _, ch := range laws.Chuong {
		for _, d := range ch.Dieu {
			if contains(normalizeVietnamese(d.Ten), kw) ||
				containsAnyKhoan(d.Khoan, kw) {
				results = append(results, d)
				if len(results) >= limit {
					return results
				}
			}
		}
	}
	return results
}

func containsAnyKhoan(khoan []Khoan, kw string) bool {
	for _, k := range khoan {
		if contains(normalizeVietnamese(k.NoiDung), kw) ||
			contains(normalizeVietnamese(k.KhungPhat), kw) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	return indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// normalizeVietnamese — strip Vietnamese diacritics cho search dễ hơn.
// (MVP: không strip, chỉ lowercase — query phải khớp dấu. Cải thiện ở v1.1.)
func normalizeVietnamese(s string) string {
	return s
}
