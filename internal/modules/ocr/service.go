// Package ocr wraps the system `tesseract` binary for offline OCR of
// PDF (after rasterization) and image files. Output is returned to the
// frontend as plain text or Markdown; the original file is never
// modified.
//
// On Linux: requires `tesseract` + `vietnamese.traineddata` installed.
// On Windows: ships a portable tesseract.exe alongside the binary.
package ocr

import (
        "bytes"
        "encoding/json"
        "fmt"
        "os"
        "os/exec"
        "path/filepath"
        "strings"
        "time"

        "kiem_sat_toolkit/internal/db"
        "kiem_sat_toolkit/internal/audit"
)

type Service struct {
        db      *db.Manager
        audit   *audit.Logger
        dataDir string
}

func New(dbMgr *db.Manager, auditLog *audit.Logger, dataDir string) *Service {
        return &Service{db: dbMgr, audit: auditLog, dataDir: dataDir}
}

// OCROptions cấu hình cho một lần OCR.
type OCROptions struct {
        Lang       string `json:"lang"`       // "vie" | "vie+eng" | "eng"
        OutputFmt  string `json:"outputFmt"`  // "txt" | "md" | "json"
        DPI        int    `json:"dpi"`        // for PDF rasterization (default 300)
        Pages      string `json:"pages"`      // for PDF: "1-3,5" or "" = all
        PreserveLayout bool `json:"preserveLayout"`
}

// OCROutput kết quả.
type OCROutput struct {
        Text      string            `json:"text"`
        PageCount int               `json:"pageCount"`
        Pages     []string          `json:"pages"`
        Confidence float64          `json:"confidence"`
        OutputPath string           `json:"outputPath"`
        Lang      string            `json:"lang"`
        Duration  string            `json:"duration"`
        Warnings  []string          `json:"warnings"`
}

// Recognize runs OCR on an image or PDF file.
// For PDFs, we rasterize each page via `pdftoppm` (Poppler) — if not
// available, we fall back to ImageMagick `convert`.
func (s *Service) Recognize(filePath string, opts OCROptions, actor string) OCROutput {
        start := time.Now()
        if opts.Lang == "" {
                opts.Lang = "vie"
        }
        if opts.DPI == 0 {
                opts.DPI = 300
        }
        if opts.OutputFmt == "" {
                opts.OutputFmt = "txt"
        }

        out := OCROutput{Lang: opts.Lang}

        if _, err := os.Stat(filePath); err != nil {
                out.Warnings = append(out.Warnings, fmt.Sprintf("file not found: %s", filePath))
                return out
        }

        ext := strings.ToLower(filepath.Ext(filePath))
        var pages []string

        switch ext {
        case ".pdf":
                pages, out.Warnings = rasterizePDF(filePath, opts, s.dataDir)
        case ".png", ".jpg", ".jpeg", ".tif", ".tiff", ".bmp", ".webp":
                pages = []string{filePath}
        default:
                out.Warnings = append(out.Warnings, fmt.Sprintf("unsupported extension: %s", ext))
                return out
        }

        out.PageCount = len(pages)
        out.Pages = make([]string, 0, len(pages))

        var allText strings.Builder
        for i, p := range pages {
                text, conf, warn := runTesseract(p, opts.Lang)
                if warn != "" {
                        out.Warnings = append(out.Warnings, fmt.Sprintf("page %d: %s", i+1, warn))
                }
                if opts.PreserveLayout {
                        allText.WriteString(fmt.Sprintf("<!-- Page %d -->\n\n", i+1))
                }
                allText.WriteString(text)
                allText.WriteString("\n\n----\n\n")
                out.Pages = append(out.Pages, text)
                out.Confidence += conf
        }
        if out.PageCount > 0 {
                out.Confidence /= float64(out.PageCount)
        }

        out.Text = allText.String()

        // Write output file
        outFile := filepath.Join(s.dataDir, "exports",
                fmt.Sprintf("ocr_%s.%s", sanitizeBase(filePath), opts.OutputFmt))
        _ = os.MkdirAll(filepath.Dir(outFile), 0755)
        _ = os.WriteFile(outFile, []byte(out.Text), 0644)
        out.OutputPath = outFile

        out.Duration = time.Since(start).Round(time.Millisecond).String()

        if s.audit != nil {
                d, _ := json.Marshal(map[string]interface{}{
                        "file": filePath, "lang": opts.Lang, "pages": out.PageCount,
                        "confidence": out.Confidence,
                })
                s.audit.Log("ocr", "recognize", actor, string(d), "restricted")
        }
        return out
}

func runTesseract(imagePath, lang string) (string, float64, string) {
        // tesseract <image> stdout -l <lang> --psm 6
        cmd := exec.Command("tesseract", imagePath, "stdout",
                "-l", lang, "--psm", "6",
        )
        var stdout, stderr bytes.Buffer
        cmd.Stdout = &stdout
        cmd.Stderr = &stderr
        if err := cmd.Run(); err != nil {
                return "", 0, fmt.Sprintf("tesseract: %v (%s)", err, strings.TrimSpace(stderr.String()))
        }
        text := stdout.String()

        // Get confidence via tsv output
        confCmd := exec.Command("tesseract", imagePath, "stdout",
                "-l", lang, "--psm", "6", "tsv",
        )
        var confOut, confErr bytes.Buffer
        confCmd.Stdout = &confOut
        confCmd.Stderr = &confErr
        conf := 0.0
        if err := confCmd.Run(); err == nil {
                conf = avgConfFromTSV(confOut.String())
        }
        return text, conf, ""
}

func avgConfFromTSV(tsv string) float64 {
        lines := strings.Split(tsv, "\n")
        if len(lines) < 2 {
                return 0
        }
        sum := 0.0
        n := 0
        for _, line := range lines[1:] {
                fields := strings.Split(line, "\t")
                if len(fields) < 12 {
                        continue
                }
                // field 11 (0-indexed) = confidence
                var c float64
                _, err := fmt.Sscanf(fields[11], "%f", &c)
                if err != nil || c < 0 {
                        continue
                }
                sum += c
                n++
        }
        if n == 0 {
                return 0
        }
        return sum / float64(n)
}

func rasterizePDF(pdfPath string, opts OCROptions, dataDir string) ([]string, []string) {
        tmpDir := filepath.Join(dataDir, ".ocr_tmp", sanitizeBase(pdfPath))
        _ = os.MkdirAll(tmpDir, 0755)

        // Try pdftoppm first (Poppler)
        ppm := exec.Command("pdftoppm", "-r", fmt.Sprintf("%d", opts.DPI),
                "-png", pdfPath, filepath.Join(tmpDir, "page"))
        var ppmErr bytes.Buffer
        ppm.Stderr = &ppmErr
        if err := ppm.Run(); err == nil {
                matches, _ := filepath.Glob(filepath.Join(tmpDir, "page-*.png"))
                return matches, nil
        }

        // Fallback: ImageMagick
        im := exec.Command("convert", "-density", fmt.Sprintf("%d", opts.DPI),
                pdfPath, filepath.Join(tmpDir, "page.png"))
        var imErr bytes.Buffer
        im.Stderr = &imErr
        if err := im.Run(); err == nil {
                matches, _ := filepath.Glob(filepath.Join(tmpDir, "page-*.png"))
                if len(matches) == 0 {
                        matches, _ = filepath.Glob(filepath.Join(tmpDir, "page.png"))
                }
                return matches, nil
        }

        return nil, []string{
                fmt.Sprintf("pdftoppm: %s", strings.TrimSpace(ppmErr.String())),
                fmt.Sprintf("convert: %s", strings.TrimSpace(imErr.String())),
                "Cần cài Poppler (pdftoppm) hoặc ImageMagick (convert) để OCR PDF.",
        }
}

func sanitizeBase(p string) string {
        base := filepath.Base(p)
        base = strings.TrimSuffix(base, filepath.Ext(base))
        for _, r := range []string{" ", "/", "\\", ":"} {
                base = strings.ReplaceAll(base, r, "_")
        }
        return base
}

// AvailableLanguages returns the list of tesseract .traineddata files
// available on the system.
func (s *Service) AvailableLanguages() []string {
        out, err := exec.Command("tesseract", "--list-langs").Output()
        if err != nil {
                return []string{"vie", "eng"}
        }
        lines := strings.Split(string(out), "\n")
        var langs []string
        for i, l := range lines {
                if i == 0 || strings.TrimSpace(l) == "" {
                        continue
                }
                langs = append(langs, strings.TrimSpace(l))
        }
        return langs
}

// IsAvailable checks whether the tesseract binary is on PATH.
func (s *Service) IsAvailable() bool {
        _, err := exec.LookPath("tesseract")
        return err == nil
}
