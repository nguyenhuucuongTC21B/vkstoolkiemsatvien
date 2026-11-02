// modules.go — 5 Module nghiệp vụ AI cho ngành Kiểm sát.
//
// Mỗi module nhận Job.Input (map fields) + Model, build prompt, gọi Predict,
// trả về output string. Trả error nếu input thiếu field bắt buộc.
package ai

import (
        "encoding/json"
        "fmt"
        "strings"
)

// dispatchModule — router từ module name → handler.
func dispatchModule(model Model, job Job) (string, error) {
        switch job.Module {
        case "extractor":
                return runExtractor(model, job.Input)
        case "reporter":
                return runReporter(model, job.Input)
        case "chatter":
                return runChatter(model, job.Input)
        case "auditor":
                return runAuditor(model, job.Input)
        case "advisor":
                return runAdvisor(model, job.Input)
        default:
                return "", fmt.Errorf("unknown module: %s", job.Module)
        }
}

// ─────────────────── MODULE 1: EXTRACTOR ───────────────────
// Bóc tách dữ liệu từ văn bản thô OCR.
// Input: {"text": "văn bản lời khai/Quyết định điều tra..."}
// Output: JSON sạch với nhan_than, timeline, vat_chung, tom_tat_vu_an.
func runExtractor(model Model, input map[string]interface{}) (string, error) {
        text, _ := input["text"].(string)
        if text == "" {
                return "", fmt.Errorf("field 'text' required")
        }

        prompt := fmt.Sprintf(`[EXTRACTOR]
Bạn là trợ lý bóc tách dữ liệu cho Viện Kiểm sát. Đọc văn bản thô OCR từ biên bản lời khai hoặc quyết định điều tra dưới đây và trích xuất thành JSON.

Văn bản:
"""
%s
"""

Yêu cầu đầu ra BẮT BUỘC là JSON hợp lệ theo đúng schema sau (không được thêm chú thích, không markdown):
{
  "nhan_than": {
    "ho_ten": "...", "mst": "...", "cccd": "...", "ngay_sinh": "YYYY-MM-DD",
    "noi_sinh": "...", "nghe_nghiep": "...", "noi_cu_tru": "..."
  },
  "timeline": [{"thoi_diem": "YYYY-MM-DDTHH:MM", "su_kien": "..."}],
  "vat_chung": [{"ten": "...", "so_luong": 0, "tinh_trang": "..."}],
  "tom_tat_vu_an": "..."
}
Chỉ trả về JSON, không thêm gì khác.`, truncate(text, 1500))

        out, err := model.Predict(prompt, 512, 0.0) // temp=0 cho deterministic
        if err != nil {
                return "", err
        }

        // Verify output là JSON hợp lệ — nếu không, wrap lại
        cleaned := cleanJSONOutput(out)
        if !isJSON(cleaned) {
                return "", fmt.Errorf("AI output is not valid JSON: %s", truncate(cleaned, 200))
        }
        return cleaned, nil
}

// ─────────────────── MODULE 2: REPORTER ───────────────────
// Dự thảo văn bản báo cáo/cáo trạng với 3 phần chuẩn.
// Input: {"case_info": "..."} hoặc {"extractor_output": "<JSON>"}
func runReporter(model Model, input map[string]interface{}) (string, error) {
        caseInfo, _ := input["case_info"].(string)
        if caseInfo == "" {
                // Allow passing extractor_output JSON instead
                if ext, ok := input["extractor_output"].(string); ok && ext != "" {
                        caseInfo = ext
                }
        }
        if caseInfo == "" {
                return "", fmt.Errorf("field 'case_info' or 'extractor_output' required")
        }

        prompt := fmt.Sprintf(`[REPORTER]
Bạn là Kiểm sát viên kinh nghiệm. Viết BÁO CÁO ĐỀ XUẤT GIẢI QUYẾT VỤ ÁN dựa trên thông tin vụ án dưới đây.

Thông tin vụ án:
"""
%s
"""

Bố cục BẮT BUỘC gồm 3 phần:
I. TỔNG QUAN VỤ ÁN: Mô tả sự việc, thời gian, địa điểm, các bên liên quan.
II. ĐIỂM NHẤN VỤ ÁN: Các chi tiết quan trọng, hành vi phạm tội, vật chứng, mức độ thương hại.
III. ĐÁNH GIÁ RỦI RO / ĐỀ XUẤT: Rủi ro (bỏ trốn, tiêu hủy chứng cứ), đề xuất biện pháp ngăn chặn, tội danh, khung hình phạt.

Văn phong hành chính, ngắn gọn, chính xác. Không thêm ý kiến cá nhân.`, truncate(caseInfo, 1500))

        out, err := model.Predict(prompt, 1024, 0.3)
        if err != nil {
                return "", err
        }
        return strings.TrimSpace(out), nil
}

// ─────────────────── MODULE 3: CHATTER ───────────────────
// Hỏi đáp hồ sơ vụ án dựa trên ngữ cảnh.
// Input: {"documents": "toàn văn chứng cứ", "question": "câu hỏi KSV"}
func runChatter(model Model, input map[string]interface{}) (string, error) {
        docs, _ := input["documents"].(string)
        question, _ := input["question"].(string)
        if docs == "" || question == "" {
                return "", fmt.Errorf("fields 'documents' and 'question' required")
        }

        prompt := fmt.Sprintf(`[CHATTER]
Bạn là trợ lý đọc bút lục vụ án cho Kiểm sát viên. Đọc tài liệu dưới đây làm ngữ cảnh, sau đó trả lời câu hỏi NGẮN GỌN, CHÍNH XÁC. Chỉ ra cụ thể chi tiết nằm ở đoạn nào trong tài liệu.

Tài liệu bút lục:
"""
%s
"""

Câu hỏi của Kiểm sát viên: %s

Trả lời:`, truncate(docs, 1800), truncate(question, 200))

        out, err := model.Predict(prompt, 256, 0.2)
        if err != nil {
                return "", err
        }
        return strings.TrimSpace(out), nil
}

// ─────────────────── MODULE 4: AUDITOR & CROSS-CHECKER ───────────────────
// So sánh đối chiếu 2 lời khai, phát hiện mâu thuẫn hoặc "NHẤT QUÁN".
// Input: {"statement_a": "lời khai bị can A", "statement_b": "lời khai nhân chứng B"}
func runAuditor(model Model, input map[string]interface{}) (string, error) {
        stmtA, _ := input["statement_a"].(string)
        stmtB, _ := input["statement_b"].(string)
        if stmtA == "" || stmtB == "" {
                return "", fmt.Errorf("fields 'statement_a' and 'statement_b' required")
        }

        prompt := fmt.Sprintf(`[AUDITOR]
Bạn là Kiểm sát viên rà soát mâu thuẫn lời khai. So sánh 2 lời khai dưới đây và phát hiện mâu thuẫn về: (1) thời gian, (2) địa điểm, (3) hành vi gây án, (4) danh tính người liên quan.

Lời khai A (bị can):
"""
%s
"""

Lời khai B (nhân chứng):
"""
%s
"""

QUY TẮC ĐẦU RA:
- Nếu 2 lời khai NHẤT QUÁN hoàn toàn, chỉ trả về đúng chữ "NHẤT QUÁN" (không thêm gì).
- Nếu có mâu thuẫn, liệt kê từng điểm mâu thuẫn theo format:
  MAU THUẬN:
  1. <loại mâu thuẫn>: <chi tiết A> vs <chi tiết B> → <đề xuất làm rõ>
  2. ...
Kết luận bằng "Đề xuất: ..." đề xuất bước tiếp theo.`, truncate(stmtA, 900), truncate(stmtB, 900))

        out, err := model.Predict(prompt, 384, 0.1)
        if err != nil {
                return "", err
        }
        return strings.TrimSpace(out), nil
}

// ─────────────────── MODULE 5: LEGAL ADVISOR ───────────────────
// Tra cứu luật offline từ laws.json + AI đề xuất điều/khung phạt.
// Input: {"behavior": "mô tả hành vi phạm tội bóc tách từ Module 1"}
func runAdvisor(model Model, input map[string]interface{}) (string, error) {
        behavior, _ := input["behavior"].(string)
        if behavior == "" {
                return "", fmt.Errorf("field 'behavior' required")
        }

        // Load laws.json (cached in package-level var, see laws.go)
        lawsJSON := cachedLawsJSON()

        prompt := fmt.Sprintf(`[ADVISOR]
Bạn là chuyên gia pháp lý hình sự Việt Nam. Dựa trên cơ sở dữ liệu luật dưới đây và hành vi phạm tội được mô tả, đề xuất:

Cơ sở dữ liệu BLHS (trích):
"""
%s
"""

Hành vi phạm tội cần phân tích: %s

Đầu ra BẮT BUỘC theo format:
PHÂN TÍCH PHÁP LÝ:
1. Hành vi phạm tội: <mô tả ngắn>
2. Mã chương/điều BLHS: <số chương, số điều>
3. Khung hình phạt dự kiến: <liệt kê các khoản + mức phạt>
4. Đề xuất áp dụng: <khoản nào, mức phạt cụ thể>
5. Án phí dự kiến: <theo NR 326/2016>`, truncate(lawsJSON, 1200), truncate(behavior, 400))

        out, err := model.Predict(prompt, 512, 0.2)
        if err != nil {
                return "", err
        }
        return strings.TrimSpace(out), nil
}

// ─────────────────── helpers ───────────────────

// cleanJSONOutput — strip markdown fences ```json ... ``` và text thừa.
func cleanJSONOutput(s string) string {
        s = strings.TrimSpace(s)
        // Remove leading ```json or ```
        if strings.HasPrefix(s, "```json") {
                s = strings.TrimPrefix(s, "```json")
        } else if strings.HasPrefix(s, "```") {
                s = strings.TrimPrefix(s, "```")
        }
        if strings.HasSuffix(s, "```") {
                s = strings.TrimSuffix(s, "```")
        }
        s = strings.TrimSpace(s)
        // Find first { and last } — trim everything outside
        if i := strings.Index(s, "{"); i > 0 {
                s = s[i:]
        }
        if i := strings.LastIndex(s, "}"); i >= 0 && i < len(s)-1 {
                s = s[:i+1]
        }
        return s
}

// isJSON — check if string is valid JSON.
func isJSON(s string) bool {
        var v interface{}
        return json.Unmarshal([]byte(s), &v) == nil
}
