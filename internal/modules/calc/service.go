// Package calc implements the prosecutor-office calculation tools:
//   - An phí hình sự / dân sự (court fees)
//   - Tiền lãi trong hạn / quá hạn / chậm trả (interest)
//   - Tuổi & thời hạn tạm giữ / tạm giam (age & procedural deadlines)
//   - Thời hiệu truy cứu TNHS (statute of limitations)
//   - Bồi thường thiệt hại ngoài hợp đồng (damages)
//
// All formulas follow Bộ luật Dân sự 2015, BLTTHS 2015 (sửa đổi bởi
// Luật 128/2025), Nghị quyết 326/2016/UBTVQH14 (án phí) and
// Thông tư liên tịch 01/2016/TTLT-TANDTC-VKSNDTC-BCA-BTP.
package calc

import (
        "encoding/json"
        "fmt"
        "math"
        "time"

        "kiem_sat_toolkit/internal/audit"
        "kiem_sat_toolkit/internal/db"
        "kiem_sat_toolkit/internal/store"
)

// Service is bound to the frontend as `window.go.modules.Calc`.
type Service struct {
        db    *db.Manager
        audit *audit.Logger
}

func New(dbMgr *db.Manager, auditLog *audit.Logger) *Service {
        return &Service{db: dbMgr, audit: auditLog}
}

// ─────────────────────────── ÁN PHÍ ───────────────────────────

// AnPhiHSInput đầu vào cho tính án phí hình sự.
type AnPhiHSInput struct {
        LoaiAn         string  `json:"loaiAn"`         // "so_tham" | "phuc_tham" | "giai_doi_phuc_tham" | "toc_hinh"
        SoNguoiBiXet   int     `json:"soNguoiBiXet"`   // số bị cáo
        SoNguoiDuocFree int    `json:"soNguoiDuocFree"` // số được miễn (KN-TC, người dưới 16t, khiếm thị...)
        TienBoiThuong  float64 `json:"tienBoiThuong"`  // nếu có bồi thường dân sự trong án HS
        TienAnPhiDaNop float64 `json:"tienAnPhiDaNop"`
}

// AnPhiHSOutput kết quả án phí HS.
type AnPhiHSOutput struct {
        AnPhiCoBan     float64 `json:"anPhiCoBan"`
        AnPhiBoSung    float64 `json:"anPhiBoSung"`
        AnPhiPhaiNop   float64 `json:"anPhiPhaiNop"`
        TienHoanLai    float64 `json:"tienHoanLai"`
        TienPhaiBoSung float64 `json:"tienPhaiBoSung"`
        GhiChu         string  `json:"ghiChu"`
        ChiTiet        string  `json:"chiTiet"`
}

// AnPhiHS tính án phí hình sự theo NR 326/2016/UBTVQH14 điều 4, 5, 6.
// Mức: 100.000đ / bị cáo sơ thẩm (không quá 5.000.000đ/tổ chức),
// phúc thẩm 50% sơ thẩm; miễn án phí cho người được miễn (điều 10).
func (s *Service) AnPhiHS(in AnPhiHSInput, actor string) AnPhiHSOutput {
        const phiMotBiCao = 100_000.0
        const phiToChucToiDa = 5_000_000.0
        const lePhiPhucThamTiLe = 0.5

        soBiXet := in.SoNguoiBiXet
        if soBiXet < 0 {
                soBiXet = 0
        }
        soFree := in.SoNguoiDuocFree
        if soFree > soBiXet {
                soFree = soBiXet
        }
        soChiuPhi := soBiXet - soFree

        var phiCoBan, phiBoSung float64

        switch in.LoaiAn {
        case "so_tham":
                phiCoBan = float64(soChiuPhi) * phiMotBiCao
                if phiCoBan > phiToChucToiDa {
                        phiCoBan = phiToChucToiDa
                }
        case "phuc_tham":
                // Phúc thẩm: 50% án phí sơ thẩm đã nộp (theo điều 5 NR 326)
                // Nếu chưa nộp sơ thẩm → tính bằng 50% của mức sơ thẩm.
                base := float64(soChiuPhi) * phiMotBiCao
                if base > phiToChucToiDa {
                        base = phiToChucToiDa
                }
                phiCoBan = base * lePhiPhucThamTiLe
        case "giai_doi_phuc_tham":
                // Giai đoạn phúc thẩm khi giải quyết kháng cáo/giám đốc thẩm
                // không phải nộp lại nếu án sơ thẩm đã nộp đủ.
                phiCoBan = 0
        case "toc_hinh":
                // Tố cẩm hình sự: án phí = 1/2 sơ thẩm (điều 6 NR 326).
                base := float64(soChiuPhi) * phiMotBiCao
                phiCoBan = base * 0.5
        default:
                phiCoBan = 0
        }

        // Án phí bổ sung: nếu án có bồi thường dân sự → 1% giá trị bồi thường
        // (theo điều 9 NR 326 trích dụng cho phần dân sự trong án hình sự)
        if in.TienBoiThuong > 0 {
                phiBoSung = math.Max(in.TienBoiThuong*0.01, 100_000)
        }

        phaiNop := phiCoBan + phiBoSung
        daNop := in.TienAnPhiDaNop
        var hoanLai, phaiBoSung float64
        if daNop > phaiNop {
                hoanLai = daNop - phaiNop
        } else {
                phaiBoSung = phaiNop - daNop
        }

        out := AnPhiHSOutput{
                AnPhiCoBan:     phiCoBan,
                AnPhiBoSung:    phiBoSung,
                AnPhiPhaiNop:   phaiNop,
                TienHoanLai:    hoanLai,
                TienPhaiBoSung: phaiBoSung,
                ChiTiet: fmt.Sprintf(
                        "Sơ thẩm: %d bị cáo − %d miễn = %d chịu phí × 100.000đ = %.0fđ;"+
                                " bổ sung 1%%×BT %.0fđ = %.0fđ; đã nộp %.0fđ.",
                        soBiXet, soFree, soChiuPhi, phiCoBan,
                        in.TienBoiThuong, phiBoSung, daNop,
                ),
                GhiChu: "Áp dụng Nghị quyết 326/2016/UBTVQH14 (điều 4, 5, 6, 9, 10).",
        }

        s.persistCalc("an_phi_hs", in, out, actor)
        return out
}

// AnPhiDSInput đầu vào cho án phí dân sự sơ thẩm.
type AnPhiDSInput struct {
        LoaiAn         string  `json:"loaiAn"`         // "so_tham" | "phuc_tham" | "giai_doi" | "sotradon"
        GiaTranTuyen   float64 `json:"giaTranTuyen"`   // giá trị tranh chấp (VND), 0 nếu vô giá
        LaVoGia        bool    `json:"laVoGia"`        // tranh chấp không có giá (ly hôn, nhận con...)
        TienAnPhiDaNop float64 `json:"tienAnPhiDaNop"`
}

// AnPhiDSOutput kết quả án phí DS.
type AnPhiDSOutput struct {
        AnPhiCoBan     float64 `json:"anPhiCoBan"`
        TienHoanLai    float64 `json:"tienHoanLai"`
        TienPhaiBoSung float64 `json:"tienPhaiBoSung"`
        PhanTramApDung float64 `json:"phanTramApDung"`
        BacApDung      string  `json:"bacApDung"`
        ChiTiet        string  `json:"chiTiet"`
        GhiChu         string  `json:"ghiChu"`
}

// AnPhiDS tính án phí dân sự sơ thẩm theo điều 7 NR 326/2016/UBTVQH14.
// Bậc thang (% lũy tiến theo từng phần giá trị tranh chấp):
//
//      ≤ 100tr: 5% (min 200k, max 5tr)
//      100tr–400tr: 4%
//      400tr–1 tỷ: 3%
//      > 1 tỷ: 2% (các phần vượt giới hạn)
func (s *Service) AnPhiDS(in AnPhiDSInput, actor string) AnPhiDSOutput {
        const (
                bac1ToiDa = 100_000_000.0
                bac2ToiDa = 400_000_000.0
                bac3ToiDa = 1_000_000_000.0
                bac1Le    = 0.05
                bac2Le    = 0.04
                bac3Le    = 0.03
                bac4Le    = 0.02
                minPhi    = 200_000.0
                maxBac1   = 5_000_000.0
        )
        const lePhiPhucThamTiLe = 0.5

        if in.LaVoGia {
                // Án phí vô giá: 300.000đ (điều 8 NR 326)
                const voGia = 300_000.0
                phaiNop := voGia
                if in.LoaiAn == "phuc_tham" {
                        phaiNop = voGia * lePhiPhucThamTiLe
                }
                out := AnPhiDSOutput{
                        AnPhiCoBan:     phaiNop,
                        PhanTramApDung: 0,
                        BacApDung:      "Vô giá",
                        ChiTiet:        fmt.Sprintf("Tranh chấp không có giá: 300.000đ%s.", loaiSuffix(in.LoaiAn)),
                        GhiChu:         "Điều 8 NR 326/2016/UBTVQH14.",
                }
                setDaNop(&out, in.TienAnPhiDaNop)
                s.persistCalc("an_phi_ds_vo_gia", in, out, actor)
                return out
        }

        g := in.GiaTranTuyen
        if g < 0 {
                g = 0
        }
        var phi float64
        var chiTiet string

        // Tính theo bậc
        t1 := math.Min(g, bac1ToiDa)
        phi += t1 * bac1Le

        if g > bac1ToiDa {
                t2 := math.Min(g-bac1ToiDa, bac2ToiDa-bac1ToiDa)
                phi += t2 * bac2Le
        }
        if g > bac2ToiDa {
                t3 := math.Min(g-bac2ToiDa, bac3ToiDa-bac2ToiDa)
                phi += t3 * bac3Le
        }
        if g > bac3ToiDa {
                t4 := g - bac3ToiDa
                phi += t4 * bac4Le
        }

        if phi < minPhi {
                phi = minPhi
        }
        if phi > maxBac1 {
                // Bậc 1 cap
                if g <= bac1ToiDa {
                        phi = maxBac1
                }
        }

        if in.LoaiAn == "phuc_tham" {
                phi = phi * lePhiPhucThamTiLe
                chiTiet = fmt.Sprintf("Phúc thẩm = 50%% × án phí sơ thẩm = %.0fđ.", phi)
        } else if in.LoaiAn == "giai_doi" {
                // Giải quyết việc dân sự (sơ thẩm) — phí cố định 300.000đ (điều 8).
                phi = 300_000
                chiTiet = fmt.Sprintf("Giải quyết việc dân sự không có tranh chấp tài sản: 300.000đ.")
        } else {
                chiTiet = fmt.Sprintf(
                        "Giá tranh chấp %.0fđ: BT×[5%%(≤100tr) + 4%%(≤400tr) + 3%%(≤1tỷ) + 2%%(>1tỷ)] = %.0fđ.",
                        g, phi,
                )
        }

        out := AnPhiDSOutput{
                AnPhiCoBan:     phi,
                PhanTramApDung: bac1Le,
                BacApDung:      bacLabel(g),
                ChiTiet:        chiTiet,
                GhiChu:         "Áp dụng điều 7 NR 326/2016/UBTVQH14 (bậc thang lũy tiến theo từng phần).",
        }
        setDaNop(&out, in.TienAnPhiDaNop)
        s.persistCalc("an_phi_ds", in, out, actor)
        return out
}

func bacLabel(g float64) string {
        switch {
        case g <= 100_000_000:
                return "Bậc 1 (≤ 100 triệu)"
        case g <= 400_000_000:
                return "Bậc 2 (100tr – 400tr)"
        case g <= 1_000_000_000:
                return "Bậc 3 (400tr – 1 tỷ)"
        default:
                return "Bậc 4 (> 1 tỷ)"
        }
}

func loaiSuffix(loai string) string {
        switch loai {
        case "phuc_tham":
                return " × 50% (phúc thẩm)"
        default:
                return ""
        }
}

func setDaNop(o *AnPhiDSOutput, daNop float64) {
        if daNop > o.AnPhiCoBan {
                o.TienHoanLai = daNop - o.AnPhiCoBan
        } else {
                o.TienPhaiBoSung = o.AnPhiCoBan - daNop
        }
}

// ─────────────────────────── TIỀN LÃI ───────────────────────────

// LaiInput đầu vào tính tiền lãi.
type LaiInput struct {
        Von           float64 `json:"von"`           // vốn gốc (VND)
        LaiSuatNam    float64 `json:"laiSuatNam"`    // %/năm (vd 9.0)
        LaiSuatQuaHan float64 `json:"laiSuatQuaHan"` // %/năm lãi quá hạn (vd 150% = 13.5)
        TuNgay        string  `json:"tuNgay"`        // "2024-01-15"
        DenNgay       string  `json:"denNgay"`       // "2026-10-30"
        NgayDaoHan    string  `json:"ngayDaoHan"`    // ngày đáo hạn (để phân chia trong hạn/quá hạn)
        LaQuaHan      bool    `json:"laQuaHan"`
        DonViTinh     string  `json:"donViTinh"`     // "nam" | "thang" | "ngay"
}

// LaiOutput kết quả.
type LaiOutput struct {
        TienLaiTrongHan float64 `json:"tienLaiTrongHan"`
        TienLaiQuaHan   float64 `json:"tienLaiQuaHan"`
        TienLaiChamTra  float64 `json:"tienLaiChamTra"` // Lãi chậm trả = 150% lãi trong hạn (BLDS 306)
        TongTienLai     float64 `json:"tongTienLai"`
        SoNgay          int     `json:"soNgay"`
        ChiTiet         string  `json:"chiTiet"`
        GhiChu          string  `json:"ghiChu"`
}

// Lai tính lãi trong hạn, quá hạn, chậm trả theo BLDS 2015 điều 306, 468, 466.
// Công thức: Lãi = Vốn × lãi suất × thời gian (theo đơn vị tính tương ứng).
//
//      Tiền lãi chậm trả = 150% × tiền lãi trong hạn (điều 306 BLDS 2015)
//      trong thời gian chậm trả.
func (s *Service) Lai(in LaiInput, actor string) LaiOutput {
        tu, err1 := time.Parse("2006-01-02", in.TuNgay)
        den, err2 := time.Parse("2006-01-02", in.DenNgay)
        if err1 != nil || err2 != nil || den.Before(tu) {
                return LaiOutput{GhiChu: "Ngày không hợp lệ. Định dạng YYYY-MM-DD."}
        }
        soNgay := int(den.Sub(tu).Hours() / 24)

        if in.LaiSuatNam == 0 {
                in.LaiSuatNam = 9.0 // mặc định tham chiếu
        }
        if in.LaiSuatQuaHan == 0 {
                in.LaiSuatQuaHan = in.LaiSuatNam * 1.5
        }

        von := in.Von
        laiNamTrongHan := in.LaiSuatNam / 100.0
        laiNamQuaHan := in.LaiSuatQuaHan / 100.0

        // Phân chia trong hạn/quá hạn nếu có ngày đáo hạn
        var laiTrongHan, laiQuaHan float64
        if in.NgayDaoHan != "" {
                daoHan, err := time.Parse("2006-01-02", in.NgayDaoHan)
                if err == nil && daoHan.After(tu) && daoHan.Before(den) {
                        ngayTrongHan := int(daoHan.Sub(tu).Hours() / 24)
                        ngayQuaHan := soNgay - ngayTrongHan
                        laiTrongHan = von * laiNamTrongHan * float64(ngayTrongHan) / 365.0
                        laiQuaHan = von * laiNamQuaHan * float64(ngayQuaHan) / 365.0
                } else if !daoHan.Before(den) {
                        // còn trong hạn toàn bộ
                        laiTrongHan = von * laiNamTrongHan * float64(soNgay) / 365.0
                } else {
                        // đã quá hạn toàn bộ
                        laiQuaHan = von * laiNamQuaHan * float64(soNgay) / 365.0
                }
        } else if in.LaQuaHan {
                laiQuaHan = von * laiNamQuaHan * float64(soNgay) / 365.0
        } else {
                laiTrongHan = von * laiNamTrongHan * float64(soNgay) / 365.0
        }

        // Lãi chậm trả (điều 306) = 150% × lãi trong hạn trong thời gian chậm trả
        laiChamTra := 0.0
        if laiQuaHan > 0 {
                laiChamTra = von * laiNamTrongHan * 1.5 * float64(soNgay) / 365.0
        }

        tong := laiTrongHan + laiQuaHan
        out := LaiOutput{
                TienLaiTrongHan: laiTrongHan,
                TienLaiQuaHan:   laiQuaHan,
                TienLaiChamTra:  laiChamTra,
                TongTienLai:     tong,
                SoNgay:          soNgay,
                ChiTiet: fmt.Sprintf(
                        "Vốn %.0fđ, %d ngày, lãi trong hạn %.2f%%/năm = %.0fđ; lãi quá hạn %.2f%%/năm = %.0fđ.",
                        von, soNgay, in.LaiSuatNam, laiTrongHan, in.LaiSuatQuaHan, laiQuaHan,
                ),
                GhiChu: "BLDS 2015 điều 306, 468, 466; công thức Lãi = Vốn × r × t/365.",
        }
        s.persistCalc("lai", in, out, actor)
        return out
}

// ─────────────────── TUỔI & THỜI HẠN TKTG ───────────────────

// TuoiInput cho tính tuổi + thời hạn.
type TuoiInput struct {
        NgaySinh          string `json:"ngaySinh"`          // "1990-05-20"
        NgayXayRaSuViec   string `json:"ngayXayRaSuViec"`   // "2026-10-30"
}

type TuoiOutput struct {
        Tuoi       int    `json:"tuoi"`
        TuoiThang  int    `json:"tuoiThang"`
        TuoiNgay   int    `json:"tuoiNgay"`
        DaNghiAm   bool   `json:"daNghiAm"`   // dưới 16t
        DaNghiAm18 bool   `json:"daNghiAm18"` // dưới 18t
        DaNghiAm70 bool   `json:"daNghiAm70"` // nam ≥ 60t (theo quy định TKG)
        ChiTiet    string `json:"chiTiet"`
        GhiChu     string `json:"ghiChu"`
}

func (s *Service) Tuoi(in TuoiInput, actor string) TuoiOutput {
        ns, e1 := time.Parse("2006-01-02", in.NgaySinh)
        sv, e2 := time.Parse("2006-01-02", in.NgayXayRaSuViec)
        if e1 != nil || e2 != nil || sv.Before(ns) {
                return TuoiOutput{GhiChu: "Ngày không hợp lệ."}
        }

        years, months, days := tinhTuoiChiTiet(ns, sv)
        out := TuoiOutput{
                Tuoi:       years,
                TuoiThang:  months,
                TuoiNgay:   days,
                DaNghiAm:   years < 16,
                DaNghiAm18: years < 18,
                DaNghiAm70: years >= 60,
                ChiTiet: fmt.Sprintf("Từ %s đến %s: %d năm %d tháng %d ngày.",
                        in.NgaySinh, in.NgayXayRaSuViec, years, months, days),
                GhiChu: "Sinh năm 1999 trở đi chịu TNHS từ đủ 16 tuổi (BLHS 12).",
        }
        s.persistCalc("tuoi", in, out, actor)
        return out
}

func tinhTuoiChiTiet(ns, sv time.Time) (int, int, int) {
        y, m, d := sv.Date()
        y2, m2, d2 := ns.Date()
        years := y - y2
        months := int(m) - int(m2)
        days := d - d2
        if days < 0 {
                lastMonth := time.Date(y, m, 0, 0, 0, 0, 0, time.UTC)
                days += lastMonth.Day()
                months--
        }
        if months < 0 {
                months += 12
                years--
        }
        return years, months, days
}

// ThoiHanInput đầu vào cho tính thời hạn tố tụng.
type ThoiHanInput struct {
        Loai       string `json:"loai"`       // "tam_giu" | "tam_giam" | "tra_ho_so" | "xet_xu" | "khau_keo"
        NgayBatDau string `json:"ngayBatDau"`
        SoNgay     int    `json:"soNgay"`     // số ngày theo luật định
        SoThang    int    `json:"soThang"`    // số tháng theo luật định (ưu tiên soThang nếu > 0)
        GioBatDau  string `json:"gioBatDau"`  // "2026-10-30T08:00" cho tạm giữ tính theo giờ
        SoGio      int    `json:"soGio"`      // số giờ (cho tạm giữ 12h/24h/48h/72h)
}

type ThoiHanOutput struct {
        NgayBatDau   string `json:"ngayBatDau"`
        NgayKetThuc  string `json:"ngayKetThuc"`
        SoNgayThucTe int    `json:"soNgayThucTe"`
        TinhTheoGio  bool   `json:"tinhTheoGio"`
        GioKetThuc   string `json:"gioKetThuc"`
        ChiTiet      string `json:"chiTiet"`
        GhiChu       string `json:"ghiChu"`
}

// ThoiHan tính ngày kết thúc của các thời hạn tố tụng.
// Theo BLTTHS 2015 (sửa 128/2025): tạm giữ 12h → 24h → 48h → 72h → 4 tháng tạm giam
// (theo điều 123, 119, 120, 121, 122 BLTTHS).
func (s *Service) ThoiHan(in ThoiHanInput, actor string) ThoiHanOutput {
        if in.SoGio > 0 && in.GioBatDau != "" {
                t, err := time.Parse("2006-01-02T15:04", in.GioBatDau)
                if err != nil {
                        return ThoiHanOutput{GhiChu: "Định dạng giờ: YYYY-MM-DDTHH:MM"}
                }
                end := t.Add(time.Duration(in.SoGio) * time.Hour)
                out := ThoiHanOutput{
                        NgayBatDau:   t.Format("2006-01-02 15:04"),
                        NgayKetThuc:  end.Format("2006-01-02"),
                        SoNgayThucTe: int(end.Sub(t).Hours() / 24),
                        TinhTheoGio:  true,
                        GioKetThuc:   end.Format("2006-01-02 15:04"),
                        ChiTiet: fmt.Sprintf("Tạm giữ %d giờ từ %s → hết hạn %s.",
                                in.SoGio, t.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04")),
                        GhiChu: "BLTTHS 2015 điều 128 (sửa 128/2025) về tạm giữ theo giờ.",
                }
                s.persistCalc("thoihan_gio", in, out, actor)
                return out
        }

        t, err := time.Parse("2006-01-02", in.NgayBatDau)
        if err != nil {
                return ThoiHanOutput{GhiChu: "Ngày: YYYY-MM-DD"}
        }
        days := in.SoNgay
        if in.SoThang > 0 {
                days = in.SoThang * 30 // quy đổi theo điều 35 BLDS 2015: 1 tháng = 30 ngày
        }
        if days <= 0 {
                days = 1
        }
        end := t.AddDate(0, 0, days)
        out := ThoiHanOutput{
                NgayBatDau:   in.NgayBatDau,
                NgayKetThuc:  end.Format("2006-01-02"),
                SoNgayThucTe: days,
                ChiTiet: fmt.Sprintf("Từ %s cộng %d ngày → kết thúc %s (theo điều 35 BLDS: 1 tháng = 30 ngày).",
                        in.NgayBatDau, days, end.Format("2006-01-02")),
                GhiChu: "Áp dụng điều 35 BLDS 2015 cho quy đổi tháng→ngày.",
        }
        s.persistCalc("thoihan_ngay", in, out, actor)
        return out
}

// ─────────────────── THỜI HIỆU TNHS ───────────────────

// ThoiHieuInput đầu vào.
type ThoiHieuInput struct {
        NgayXayRaSuViec string `json:"ngayXayRaSuViec"`
        LoaiToiPham     string `json:"loaiToiPham"` // "it_nghiem_trong" | "nghiem_trong" | "rat_nghiem_trong" | "dac_biet_nghiem_trong"
        AnNhanDinh      bool   `json:"anNhanDinh"`
        NgayAnHieuLuc   string `json:"ngayAnHieuLuc"`
}

type ThoiHieuOutput struct {
        NgayXayRa         string `json:"ngayXayRa"`
        SoNamThoiHieu     int    `json:"soNamThoiHieu"`
        NgayHetThoiHieu   string `json:"ngayHetThoiHieu"`
        ConThoiHieu       bool   `json:"conThoiHieu"`
        ConBaoNhieuNgay   int    `json:"conBaoNhieuNgay"`
        ChiTiet           string `json:"chiTiet"`
        GhiChu            string `json:"ghiChu"`
}

// ThoiHieuTNHS tính thời hiệu truy cứu trách nhiệm hình sự.
// Theo BLHS 2015 (sửa 2017) điều 27:
//
//      Toi ít nghiêm trọng (không quá 3 năm): 5 năm
//      Toi nghiêm trọng (trên 3 - 7 năm): 10 năm
//      Toi rất nghiêm trọng (trên 7 - 15 năm): 15 năm
//      Toi đặc biệt nghiêm trọng (trên 15 năm, hoặc loại hình khác): 20 năm
func (s *Service) ThoiHieuTNHS(in ThoiHieuInput, actor string) ThoiHieuOutput {
        sv, err := time.Parse("2006-01-02", in.NgayXayRaSuViec)
        if err != nil {
                return ThoiHieuOutput{GhiChu: "Ngày: YYYY-MM-DD"}
        }

        var soNam int
        switch in.LoaiToiPham {
        case "it_nghiem_trong":
                soNam = 5
        case "nghiem_trong":
                soNam = 10
        case "rat_nghiem_trong":
                soNam = 15
        case "dac_biet_nghiem_trong":
                soNam = 20
        default:
                soNam = 5
        }

        batDau := sv
        if in.AnNhanDinh && in.NgayAnHieuLuc != "" {
                if ahl, err := time.Parse("2006-01-02", in.NgayAnHieuLuc); err == nil {
                        batDau = ahl
                }
        }

        het := batDau.AddDate(soNam, 0, 0)
        now := time.Now()
        con := now.Before(het)
        conNgay := 0
        if con {
                conNgay = int(het.Sub(now).Hours() / 24)
        }

        out := ThoiHieuOutput{
                NgayXayRa:       in.NgayXayRaSuViec,
                SoNamThoiHieu:   soNam,
                NgayHetThoiHieu: het.Format("2006-01-02"),
                ConThoiHieu:     con,
                ConBaoNhieuNgay: conNgay,
                ChiTiet: fmt.Sprintf(
                        "Loại tội: %s → thời hiệu %d năm; từ %s → hết hạn %s (còn %d ngày).",
                        in.LoaiToiPham, soNam, batDau.Format("2006-01-02"),
                        het.Format("2006-01-02"), conNgay,
                ),
                GhiChu: "BLHS 2015 (sửa 2017) điều 27 — thời hiệu truy cứu TNHS.",
        }
        s.persistCalc("thoi_hieu_tnhs", in, out, actor)
        return out
}

// ───────────── BỒI THƯỜNG THIỆT HẠI NGOÀI HỢP ĐỒNG ─────────────

// BoiThuongInput cho tính bồi thường thiệt hại ngoài hợp đồng.
type BoiThuongInput struct {
        TaiSanMat             float64 `json:"taiSanMat"`
        TaiSanLoiNhuan        float64 `json:"taiSanLoiNhuan"`
        TaiSanChiPhiuKhamPha  float64 `json:"taiSanChiPhiuKhamPha"`
        ChiPhiYTe             float64 `json:"chiPhiYTe"`
        ChiPhiYTeTuongLai     float64 `json:"chiPhiYTeTuongLai"`
        ThuNhapMat            float64 `json:"thuNhapMat"`
        ThuNhapMatTuongLai    float64 `json:"thuNhapMatTuongLai"`
        ChiPhiTangLe          float64 `json:"chiPhiTangLe"`
        TienCapDuong          float64 `json:"tienCapDuong"`
        TienBoiThuongDanhDu   float64 `json:"tienBoiThuongDanhDu"`
}

type BoiThuongOutput struct {
        TongThietHaiTaiSan   float64 `json:"tongThietHaiTaiSan"`
        TongThietHaiSucKhoe  float64 `json:"tongThietHaiSucKhoe"`
        TongThietHaiTinhMang float64 `json:"tongThietHaiTinhMang"`
        TienBoiThuongDanhDu  float64 `json:"tienBoiThuongDanhDu"`
        TongTienBoiThuong    float64 `json:"tongTienBoiThuong"`
        ChiTiet              string  `json:"chiTiet"`
        GhiChu               string  `json:"ghiChu"`
}

// BoiThuongNT tính bồi thường thiệt hại ngoài hợp đồng theo BLDS 2015 điều 589-609.
func (s *Service) BoiThuongNT(in BoiThuongInput, actor string) BoiThuongOutput {
        ts := in.TaiSanMat + in.TaiSanLoiNhuan + in.TaiSanChiPhiuKhamPha
        sk := in.ChiPhiYTe + in.ChiPhiYTeTuongLai + in.ThuNhapMat + in.ThuNhapMatTuongLai
        tm := in.ChiPhiTangLe + in.TienCapDuong
        tong := ts + sk + tm + in.TienBoiThuongDanhDu

        out := BoiThuongOutput{
                TongThietHaiTaiSan:   ts,
                TongThietHaiSucKhoe:  sk,
                TongThietHaiTinhMang: tm,
                TienBoiThuongDanhDu:  in.TienBoiThuongDanhDu,
                TongTienBoiThuong:    tong,
                ChiTiet: fmt.Sprintf(
                        "TS: %.0f + SK: %.0f + TM: %.0f + DD: %.0f = %.0fđ.",
                        ts, sk, tm, in.TienBoiThuongDanhDu, tong,
                ),
                GhiChu: "BLDS 2015 điều 589-609 (bồi thường thiệt hại ngoài hợp đồng).",
        }
        s.persistCalc("boi_thuong_nt", in, out, actor)
        return out
}

// ──────────────── HISTORY / PERSISTENCE ────────────────

func (s *Service) persistCalc(module string, in, out interface{}, actor string) {
        if s == nil || s.db == nil || s.db.App == nil {
                return
        }
        inJSON, _ := json.Marshal(in)
        outJSON, _ := json.Marshal(out)
        tbl := s.db.App.Table("calc_history")
        id := tbl.NextID()
        _ = tbl.Insert(store.Row{
                "id":      id,
                "ts":      time.Now().Format("2006-01-02 15:04:05"),
                "module":  module,
                "inputs":  string(inJSON),
                "outputs": string(outJSON),
        })
        if s.audit != nil {
                s.audit.Logf("calc."+module, "compute", actor, "computed %s", module)
        }
}

// HistoryItem — bản ghi lịch sử tính toán.
type HistoryItem struct {
        ID      int    `json:"id"`
        TS      string `json:"ts"`
        Module  string `json:"module"`
        Inputs  string `json:"inputs"`
        Outputs string `json:"outputs"`
}

func (s *Service) History(limit int) []HistoryItem {
        if limit <= 0 {
                limit = 50
        }
        rows := s.db.App.Table("calc_history").SortedCopy(func(a, b store.Row) bool {
                return store.FieldInt(a, "id") > store.FieldInt(b, "id")
        })
        if len(rows) > limit {
                rows = rows[:limit]
        }
        out := make([]HistoryItem, 0, len(rows))
        for _, r := range rows {
                out = append(out, HistoryItem{
                        ID:      store.FieldInt(r, "id"),
                        TS:      store.FieldStr(r, "ts"),
                        Module:  store.FieldStr(r, "module"),
                        Inputs:  store.FieldStr(r, "inputs"),
                        Outputs: store.FieldStr(r, "outputs"),
                })
        }
        return out
}
