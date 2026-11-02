// binding_stub.go — STUB implementation (default, no build tag).
//
// Trả về mock JSON/output hợp lệ cho 5 module để dev/test mà không cần
// libllama cài sẵn. Khi user build với `-tags llama`, file này bị loại
// (build constraint) và binding_llama.go (real) được dùng thay.
//
//go:build !llama

package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// stubModel — giả lập LLM cho development.
type stubModel struct {
	cfg Config
}

// NewModel — constructor. Default = stub; real = binding_llama.go.
func NewModel(cfg Config) (Model, error) {
	return &stubModel{cfg: cfg}, nil
}

func (m *stubModel) Predict(prompt string, maxTokens int, temp float32) (string, error) {
	// Simulate inference latency (50-200ms) để worker pool thực sự chạy async
	time.Sleep(80 * time.Millisecond)

	// Detect module type from prompt prefix và trả về mock output phù hợp
	switch {
	case strings.Contains(prompt, "[EXTRACTOR]"):
		return mockExtractorOutput(prompt), nil
	case strings.Contains(prompt, "[REPORTER]"):
		return mockReporterOutput(prompt), nil
	case strings.Contains(prompt, "[CHATTER]"):
		return mockChatterOutput(prompt), nil
	case strings.Contains(prompt, "[AUDITOR]"):
		return mockAuditorOutput(prompt), nil
	case strings.Contains(prompt, "[ADVISOR]"):
		return mockAdvisorOutput(prompt), nil
	default:
		return "[stub] Mock response for prompt: " + truncate(prompt, 80), nil
	}
}

func (m *stubModel) Free() {
	// no-op for stub
}

func (m *stubModel) Info() ModelInfo {
	return ModelInfo{
		Name:        "qwen3-0.6b-instruct-q4_k_m (STUB)",
		Path:        m.cfg.ModelPath,
		ContextSize: m.cfg.ContextSize,
		Temperature: m.cfg.Temperature,
		Loaded:      true,
		Backend:     "stub",
	}
}

// ─────────── mock outputs cho 5 modules ───────────

func mockExtractorOutput(input string) string {
	// Return clean JSON struct that Go can Unmarshal
	out := map[string]interface{}{
		"nhan_than": map[string]interface{}{
			"ho_ten":      "Nguyễn Văn A",
			"mst":         "0123456789",
			"cccd":        "001234567890",
			"ngay_sinh":   "1990-05-20",
			"noi_sinh":    "Bắc Ninh",
			"nghe_nghiep": "Lái xe",
			"noi_cu_tru":  "Thôn X, xã Y, huyện Z, tỉnh Bắc Ninh",
		},
		"timeline": []map[string]interface{}{
			{"thoi_diem": "2026-09-15T20:00", "su_kien": "Bị can đến nhà nạn nhân"},
			{"thoi_diem": "2026-09-15T20:30", "su_kien": "Xảy ra ẩu đả"},
			{"thoi_diem": "2026-09-15T21:00", "su_kien": "Nạn nhân được đưa đi cấp cứu"},
		},
		"vat_chung": []map[string]interface{}{
			{"ten": "Dao phay", "so_luong": 1, "tinh_trang": "Có vết máu"},
			{"ten": "Điện thoại iPhone", "so_luong": 1, "tinh_trang": "Bị hư hỏng"},
		},
		"tom_tat_vu_an": "Bị can Nguyễn Văn A dùng dao phay đánh nạn nhân dẫn đến thương tích 32%.",
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b)
}

func mockReporterOutput(input string) string {
	return `CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM
Độc lập - Tự do - Hạnh phúc

BÁO CÁO ĐỀ XUẤT GIẢI QUYẾT VỤ ÁN

I. TỔNG QUAN VỤ ÁN:
Vụ án xảy ra vào ngày 15/09/2026 tại thôn X, xã Y, huyện Z, tỉnh Bắc Ninh.
Bị can Nguyễn Văn A, sinh năm 1990, nơi cư trú tại Bắc Ninh.
Bị can bị bắt quả tang hành vi cố ý gây thương tích nạn nhân là ông B bằng dao phay.

II. ĐIỂM NHẤN VỤ ÁN:
- Bị can có lỗi do mâu thuẫn cá nhân từ trước.
- Hành vi phạm tội được thực hiện công khai, có nhiều nhân chứng trực tiếp chứng kiến.
- Vật chứng thu giữ: 01 dao phay có vết máu, 01 điện thoại iPhone hư hỏng của nạn nhân.
- Nạn nhân bị thương tích 32% theo kết luận giám định pháp y.

III. ĐÁNH GIÁ RỦI RO / ĐỀ XUẤT:
- Rủi ro: Bị can có khả năng bỏ trốn (không có nơi cư trú cố định trong thời gian dài).
- Đề xuất:
  + Áp dụng biện pháp ngăn chặn: Tạm giam 4 tháng theo Điều 119 BLTTHS.
  + Truy cứu trách nhiệm hình sự về tội "Cố ý gây thương tích" theo Điều 134 BLHS.
  + Khung hình phạt dự kiến: 03 - 07 năm tù (khoản 4 Điều 134 BLHS).
  + Tiếp tục thu thập thêm lời khai nhân chứng để củng cố hồ sơ.`
}

func mockChatterOutput(input string) string {
	return `Trả lời: Dựa trên bút lục vụ án, bị can Nguyễn Văn A đã đến nhà nạn nhân vào lúc 20:00 ngày 15/09/2026.

Chi tiết này nằm ở: Đoạn 2, dòng 5-8 của Biên bản lời khai bị can số 01/BB-LK ngày 16/09/2026.

Cụ thể: "Tôi đến nhà anh B khoảng 8 giờ tối, lúc đó anh B đang xem TV..." (trích lời khai bị can).`
}

func mockAuditorOutput(input string) string {
	// Detect nếu có dấu hiệu mâu thuẫn
	if strings.Contains(input, "20:00") && strings.Contains(input, "21:00") {
		return `MAU THUẬN:
1. Thời gian: Lời khai bị can A cho rằng sự việc xảy ra lúc 20:00, nhưng lời khai nhân chứng B cho rằng lúc 21:00.
   → Cần đối chiếu thêm camera an ninh hoặc hóa đơn điện thoại.
2. Hành vi: Bị can khai "không có dao", nhưng nhân chứng thấy "bị can cầm dao phay".
   → Mâu thuẫn về vật chứng — cần giám định dấu vân tay trên dao thu được.

Đề xuất: Tiếp tục lấy lời khai lần 2 đối với bị can để làm rõ mâu thuẫn.`
	}
	return "NHẤT QUÁN"
}

func mockAdvisorOutput(input string) string {
	return `PHÂN TÍCH PHÁP LÝ:

1. Hành vi phạm tội bóc tách được: Cố ý gây thương tích hoặc gây tổn hại sức khỏe của người khác.
2. Mã chương/điều BLHS áp dụng:
   - Chương XVIII: Các tội xâm phạm tính mạng, sức khỏe, danh dự, nhân phẩm của con người.
   - Điều 134: Tội cố ý gây thương tích hoặc gây tổn hại sức khỏe của người khác.
3. Khung hình phạt dự kiến:
   - Khoản 1 (thương tích ≤ 11%): Phạt tù từ 06 tháng đến 03 năm.
   - Khoản 2 (có vũ khí): Phạt tù từ 01 năm đến 05 năm.
   - Khoản 4 (thương tích ≥ 31%): Phạt tù từ 03 năm đến 07 năm.
   - Khoản 5 (≥ 61% hoặc tử vong): Phạt tù từ 05 năm đến 15 năm.
4. Đề xuất áp dụng: Khoản 4 Điều 134 BLHS — phạt tù 03-07 năm.
5. Án phí: Theo NR 326/2016/UBTVQH14, án phí hình sự sơ thẩm = 100.000đ/bị cáo.`
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// prevent unused import warning when stub is built
var _ = fmt.Sprintf
