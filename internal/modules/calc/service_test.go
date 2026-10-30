package calc

import (
	"testing"
)

// TestAnPhiHS_basic verifies the core HS fee formula: 100k/bị cáo × số chịu phí
func TestAnPhiHS_basic(t *testing.T) {
	s := &Service{}
	out := s.AnPhiHS(AnPhiHSInput{
		LoaiAn:       "so_tham",
		SoNguoiBiXet: 3,
	}, "tester")
	want := 300_000.0
	if out.AnPhiCoBan != want {
		t.Errorf("AnPhiCoBan: got %.0f, want %.0f", out.AnPhiCoBan, want)
	}
}

// TestAnPhiHS_mien_phi verifies miễn án phí for exempt defendants
func TestAnPhiHS_mien_phi(t *testing.T) {
	s := &Service{}
	out := s.AnPhiHS(AnPhiHSInput{
		LoaiAn:         "so_tham",
		SoNguoiBiXet:   4,
		SoNguoiDuocFree: 2,
	}, "tester")
	// 4 - 2 = 2 chịu phí × 100k = 200k
	want := 200_000.0
	if out.AnPhiCoBan != want {
		t.Errorf("with miễn: got %.0f, want %.0f", out.AnPhiCoBan, want)
	}
}

// TestAnPhiHS_to_chuc_cap verifies tổ chức cap at 5tr
func TestAnPhiHS_to_chuc_cap(t *testing.T) {
	s := &Service{}
	out := s.AnPhiHS(AnPhiHSInput{
		LoaiAn:       "so_tham",
		SoNguoiBiXet: 100, // 100 × 100k = 10tr > 5tr cap
	}, "tester")
	want := 5_000_000.0
	if out.AnPhiCoBan != want {
		t.Errorf("cap: got %.0f, want %.0f", out.AnPhiCoBan, want)
	}
}

// TestAnPhiHS_phuc_tham verifies phúc thẩm = 50% sơ thẩm
func TestAnPhiHS_phuc_tham(t *testing.T) {
	s := &Service{}
	out := s.AnPhiHS(AnPhiHSInput{
		LoaiAn:       "phuc_tham",
		SoNguoiBiXet: 2,
	}, "tester")
	want := 100_000.0 // 2 × 100k × 50% = 100k
	if out.AnPhiCoBan != want {
		t.Errorf("phúc thẩm: got %.0f, want %.0f", out.AnPhiCoBan, want)
	}
}

// TestAnPhiDS_bac_thang verifies bậc thang DS fee
func TestAnPhiDS_bac_thang(t *testing.T) {
	s := &Service{}
	// 50tr → 5% × 50tr = 2.5tr
	out := s.AnPhiDS(AnPhiDSInput{
		LoaiAn:       "so_tham",
		GiaTranTuyen: 50_000_000,
	}, "tester")
	want := 2_500_000.0
	if out.AnPhiCoBan != want {
		t.Errorf("bậc 1: got %.0f, want %.0f", out.AnPhiCoBan, want)
	}

	// 200tr → 100tr×5% + 100tr×4% = 5tr + 4tr = 9tr
	out = s.AnPhiDS(AnPhiDSInput{
		LoaiAn:       "so_tham",
		GiaTranTuyen: 200_000_000,
	}, "tester")
	want = 9_000_000.0
	if out.AnPhiCoBan != want {
		t.Errorf("bậc 2: got %.0f, want %.0f", out.AnPhiCoBan, want)
	}

	// 1.5 tỷ → 100tr×5% + 300tr×4% + 600tr×3% + 500tr×2% = 5 + 12 + 18 + 10 = 45tr
	out = s.AnPhiDS(AnPhiDSInput{
		LoaiAn:       "so_tham",
		GiaTranTuyen: 1_500_000_000,
	}, "tester")
	want = 45_000_000.0
	if out.AnPhiCoBan != want {
		t.Errorf("bậc 4: got %.0f, want %.0f", out.AnPhiCoBan, want)
	}
}

// TestAnPhiDS_vo_gia verifies vô giá = 300k
func TestAnPhiDS_vo_gia(t *testing.T) {
	s := &Service{}
	out := s.AnPhiDS(AnPhiDSInput{
		LoaiAn:  "so_tham",
		LaVoGia: true,
	}, "tester")
	want := 300_000.0
	if out.AnPhiCoBan != want {
		t.Errorf("vô giá: got %.0f, want %.0f", out.AnPhiCoBan, want)
	}
}

// TestLai verifies interest formula
func TestLai(t *testing.T) {
	s := &Service{}
	// Vốn 100tr, 365 ngày, 9%/năm → lãi = 100tr × 0.09 × 365/365 = 9tr
	out := s.Lai(LaiInput{
		Von:        100_000_000,
		LaiSuatNam: 9.0,
		TuNgay:     "2024-01-01",
		DenNgay:    "2024-12-31", // leap year, 365 days
	}, "tester")
	// Note: 2024 is leap year; 2024-01-01 to 2024-12-31 = 365 days (not 364)
	want := 9_000_000.0
	tolerance := 1_000.0 // allow tiny rounding from day count
	diff := out.TienLaiTrongHan - want
	if diff < -tolerance || diff > tolerance {
		t.Errorf("lãi trong hạn: got %.0f, want %.0f (diff %.0f)", out.TienLaiTrongHan, want, diff)
	}
}

// TestTuoi verifies age calculation
func TestTuoi(t *testing.T) {
	s := &Service{}
	out := s.Tuoi(TuoiInput{
		NgaySinh:        "1990-05-20",
		NgayXayRaSuViec: "2026-10-30",
	}, "tester")
	if out.Tuoi != 36 {
		t.Errorf("tuổi: got %d, want 36", out.Tuoi)
	}
	if !out.DaNghiAm70 && out.Tuoi >= 60 {
		// ok — expected
	}
	// Under 16 test
	out2 := s.Tuoi(TuoiInput{
		NgaySinh:        "2015-01-01",
		NgayXayRaSuViec: "2026-01-01",
	}, "tester")
	if !out2.DaNghiAm {
		t.Errorf("under 16 should be DaNghiAm; got tuổi=%d", out2.Tuoi)
	}
}

// TestThoiHan verifies deadline calculation
func TestThoiHan(t *testing.T) {
	s := &Service{}
	// 4 tháng tạm giam = 120 ngày (theo điều 35 BLDS: 1 tháng = 30 ngày)
	out := s.ThoiHan(ThoiHanInput{
		Loai:       "tam_giam",
		NgayBatDau: "2026-01-15",
		SoThang:    4,
	}, "tester")
	want := "2026-05-15"
	if out.NgayKetThuc != want {
		t.Errorf("4 tháng TKG: got %s, want %s", out.NgayKetThuc, want)
	}
	if out.SoNgayThucTe != 120 {
		t.Errorf("số ngày: got %d, want 120", out.SoNgayThucTe)
	}
}

// TestThoiHieuTNHS verifies statute of limitations
func TestThoiHieuTNHS(t *testing.T) {
	s := &Service{}
	out := s.ThoiHieuTNHS(ThoiHieuInput{
		NgayXayRaSuViec: "2024-01-01",
		LoaiToiPham:     "it_nghiem_trong", // 5 năm
	}, "tester")
	if out.SoNamThoiHieu != 5 {
		t.Errorf("số năm: got %d, want 5", out.SoNamThoiHieu)
	}
	want := "2029-01-01"
	if out.NgayHetThoiHieu != want {
		t.Errorf("hết hạn: got %s, want %s", out.NgayHetThoiHieu, want)
	}
}

// TestBoiThuongNT verifies damages summation
func TestBoiThuongNT(t *testing.T) {
	s := &Service{}
	out := s.BoiThuongNT(BoiThuongInput{
		TaiSanMat:     10_000_000,
		ChiPhiYTe:     5_000_000,
		ThuNhapMat:    3_000_000,
		ChiPhiTangLe:  20_000_000,
	}, "tester")
	want := 38_000_000.0
	if out.TongTienBoiThuong != want {
		t.Errorf("bồi thường: got %.0f, want %.0f", out.TongTienBoiThuong, want)
	}
}
