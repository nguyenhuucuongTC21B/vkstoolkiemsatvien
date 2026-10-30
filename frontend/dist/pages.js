// ──────────── Dashboard ────────────
route("dashboard", async (page) => {
  const info = await wcall("go.main.App.GetAppInfo");
  let dbStats = "";
  const stats = await wcall("go.legal.Service.Stats");
  if (stats) {
    dbStats = Object.entries(stats).map(([k, v]) =>
      `<div class="stat"><div class="stat-num">${v}</div><div class="stat-label">${k}</div></div>`
    ).join("");
  }
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Phần mềm Công cụ ngành Kiểm sát</div>
      <p class="text-muted">Bộ công cụ nghiệp vụ hỗ trợ công tác Kiểm sát — chạy 100% ngoại tuyến,
        không cần cài đặt, single binary Windows/Linux.</p>
      <div class="spacer"></div>
      <div class="flex">
        <span class="badge badge-offline">100% Offline</span>
        <span class="text-muted">Phiên bản ${info?.version || "?"} · Build ${info?.buildDate || "?"}</span>
      </div>
    </div>

    <div class="card">
      <div class="card-title">Cơ sở dữ liệu văn bản pháp luật</div>
      <div class="stat-grid">${dbStats || "<p class='text-muted'>Chưa có dữ liệu.</p>"}</div>
    </div>

    <div class="grid-2">
      <div class="card">
        <div class="card-title">Module nghiệp vụ</div>
        <ul style="list-style:none; padding:0;">
          <li>✓ <b>Tính án phí</b> HS/DS — NR 326/2016/UBTVQH14</li>
          <li>✓ <b>Tiền lãi</b> trong hạn/quá hạn/chậm trả — BLDS 306</li>
          <li>✓ <b>Tuổi & thời hạn</b> TKTG — BLTTHS 128/2025</li>
          <li>✓ <b>Thời hiệu TNHS</b> — BLHS điều 27</li>
          <li>✓ <b>Bồi thường NTNĐ</b> — BLDS 589-609</li>
          <li>✓ <b>Tra cứu luật</b> offline (13 văn bản + 6 biểu mẫu seed)</li>
          <li>✓ <b>OCR PDF/ảnh</b> — Tesseract vie+eng</li>
          <li>✓ <b>Che PII</b> 10 loại regex</li>
          <li>✓ <b>Rà soát chính tả</b> Word/text</li>
          <li>✓ <b>Đổi tên file</b> hàng loạt</li>
          <li>✓ <b>Quản lý vụ án</b> + gắn nhãn chứng cứ</li>
          <li>✓ <b>Soạn thảo biểu mẫu</b> tố tụng</li>
        </ul>
      </div>
      <div class="card">
        <div class="card-title">Văn bản áp dụng</div>
        <ul style="list-style:none; padding:0;">
          <li>• BLHS 2015 (sửa 2017)</li>
          <li>• BLTTHS 2015 (sửa 128/2025)</li>
          <li>• BLDS 2015</li>
          <li>• BLTTDS 2015 (sửa 85/2025)</li>
          <li>• NR 326/2016/UBTVQH14 (án phí)</li>
          <li>• Luật 106/2025 (THA DS)</li>
          <li>• Luật 127/2025 (THA HS)</li>
          <li>• NĐ 151-152/2026 (THA DS)</li>
          <li>• QĐ 39/QĐ-VKSTC 2018 (THA HS)</li>
          <li>• QĐ 410/QĐ-VKSTC 2021 (biểu mẫu)</li>
          <li>• QĐ 50/QĐ-VKSTC 2026 (KN-TC)</li>
          <li>• NR 205/QH15 (khởi kiện công ích)</li>
        </ul>
      </div>
    </div>
  `;
});

// ──────────── Án phí ────────────
route("an-phi", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Tính án phí hình sự</div>
      <div class="form-grid">
        <div class="form-row">
          <label>Loại án</label>
          <select id="hs-loai">
            <option value="so_tham">Sơ thẩm</option>
            <option value="phuc_tham">Phúc thẩm</option>
            <option value="giai_doi_phuc_tham">Giải quyết kháng cáo phúc thẩm</option>
            <option value="toc_hinh">Tố cẩm hình sự</option>
          </select>
        </div>
        <div class="form-row">
          <label>Số bị cáo</label>
          <input id="hs-bicao" type="number" value="1" min="0">
        </div>
        <div class="form-row">
          <label>Số được miễn án phí</label>
          <input id="hs-free" type="number" value="0" min="0">
        </div>
        <div class="form-row">
          <label>Tiền bồi thường dân sự (VND)</label>
          <input id="hs-boithuong" type="number" value="0">
        </div>
        <div class="form-row">
          <label>Án phí đã nộp (VND)</label>
          <input id="hs-danop" type="number" value="0">
        </div>
      </div>
      <button class="btn" id="hs-calc">Tính án phí HS</button>
      <div id="hs-result"></div>
    </div>

    <div class="card">
      <div class="card-title">Tính án phí dân sự</div>
      <div class="form-grid">
        <div class="form-row">
          <label>Loại án</label>
          <select id="ds-loai">
            <option value="so_tham">Sơ thẩm</option>
            <option value="phuc_tham">Phúc thẩm</option>
            <option value="giai_doi">Giải quyết việc dân sự</option>
          </select>
        </div>
        <div class="form-row">
          <label>Giá tranh chấp (VND)</label>
          <input id="ds-gia" type="number" value="100000000">
        </div>
        <div class="form-row">
          <label>Tranh chấp vô giá?</label>
          <select id="ds-vogia">
            <option value="0">Không</option>
            <option value="1">Vô giá (ly hôn, nhận con...)</option>
          </select>
        </div>
        <div class="form-row">
          <label>Án phí đã nộp (VND)</label>
          <input id="ds-danop" type="number" value="0">
        </div>
      </div>
      <button class="btn" id="ds-calc">Tính án phí DS</button>
      <div id="ds-result"></div>
    </div>

    <div class="callout">
      <b>Cơ sở pháp lý:</b> Nghị quyết 326/2016/UBTVQH14 (điều 4, 5, 6, 7, 8, 9, 10).
      Án phí HS: 100.000đ/bị cáo sơ thẩm (≤5tr/tổ chức); phúc thẩm 50%.
      Án phí DS: bậc thang 5/4/3/2% theo giá trị (≤100tr/100-400tr/400tr-1 tỷ/>1 tỷ); vô giá 300.000đ.
    </div>
  `;

  page.querySelector("#hs-calc").onclick = async () => {
    const in_ = {
      loaiAn: page.querySelector("#hs-loai").value,
      soNguoiBiXet: +page.querySelector("#hs-bicao").value,
      soNguoiDuocFree: +page.querySelector("#hs-free").value,
      tienBoiThuong: +page.querySelector("#hs-boithuong").value,
      tienAnPhiDaNop: +page.querySelector("#hs-danop").value,
    };
    const r = await wcall("go.calc.Service.AnPhiHS", in_, "ui");
    page.querySelector("#hs-result").innerHTML = anPhiRender(r);
  };
  page.querySelector("#ds-calc").onclick = async () => {
    const in_ = {
      loaiAn: page.querySelector("#ds-loai").value,
      giaTranTuyen: +page.querySelector("#ds-gia").value,
      laVoGia: page.querySelector("#ds-vogia").value === "1",
      tienAnPhiDaNop: +page.querySelector("#ds-danop").value,
    };
    const r = await wcall("go.calc.Service.AnPhiDS", in_, "ui");
    page.querySelector("#ds-result").innerHTML = anPhiDSRender(r);
  };
});

function anPhiRender(r) {
  return `
    <div class="result-block">
      <div class="result-label">Án phí cơ bản</div>
      <div class="result-value-lg">${money(r.anPhiCoBan)}</div>
      <div class="spacer"></div>
      <div class="result-label">Án phí bổ sung (BT dân sự)</div>
      <div class="result-value">${money(r.anPhiBoSung)}</div>
      <div class="spacer"></div>
      <div class="result-label">Tổng phải nộp</div>
      <div class="result-value-lg">${money(r.anPhiPhaiNop)}</div>
      ${r.tienHoanLai > 0 ? `<div class="callout callout-success">Hoàn lại: ${money(r.tienHoanLai)}</div>` : ""}
      ${r.tienPhaiBoSung > 0 ? `<div class="callout callout-warning">Cần bổ sung: ${money(r.tienPhaiBoSung)}</div>` : ""}
      <div class="result-detail">${r.chiTiet}</div>
      <div class="result-ghichu">${r.ghiChu}</div>
    </div>
  `;
}

function anPhiDSRender(r) {
  return `
    <div class="result-block">
      <div class="result-label">Án phí ${r.bacApDung}</div>
      <div class="result-value-lg">${money(r.anPhiCoBan)}</div>
      ${r.tienHoanLai > 0 ? `<div class="callout callout-success">Hoàn lại: ${money(r.tienHoanLai)}</div>` : ""}
      ${r.tienPhaiBoSung > 0 ? `<div class="callout callout-warning">Bổ sung: ${money(r.tienPhaiBoSung)}</div>` : ""}
      <div class="result-detail">${r.chiTiet}</div>
      <div class="result-ghichu">${r.ghiChu}</div>
    </div>
  `;
}

// ──────────── Tiền lãi ────────────
route("lai", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Tính tiền lãi</div>
      <div class="form-grid">
        <div class="form-row">
          <label>Vốn gốc (VND)</label>
          <input id="l-von" type="number" value="100000000">
        </div>
        <div class="form-row">
          <label>Lãi suất trong hạn (%/năm)</label>
          <input id="l-tronghan" type="number" step="0.01" value="9.0">
        </div>
        <div class="form-row">
          <label>Lãi suất quá hạn (%/năm)</label>
          <input id="l-quahan" type="number" step="0.01" value="13.5">
        </div>
        <div class="form-row">
          <label>Ngày đáo hạn (YYYY-MM-DD, optional)</label>
          <input id="l-daohan" type="text" placeholder="2024-12-31">
        </div>
        <div class="form-row">
          <label>Từ ngày (YYYY-MM-DD)</label>
          <input id="l-tu" type="text" value="2024-01-01">
        </div>
        <div class="form-row">
          <label>Đến ngày (YYYY-MM-DD)</label>
          <input id="l-den" type="text" value="2026-10-30">
        </div>
      </div>
      <button class="btn" id="l-calc">Tính lãi</button>
      <div id="l-result"></div>
    </div>
    <div class="callout">
      <b>BLDS 2015 điều 306:</b> Tiền lãi chậm trả = 150% × tiền lãi trong hạn trong thời gian chậm trả.<br>
      Công thức: Lãi = Vốn × r × t/365.
    </div>
  `;
  page.querySelector("#l-calc").onclick = async () => {
    const in_ = {
      von: +page.querySelector("#l-von").value,
      laiSuatNam: +page.querySelector("#l-tronghan").value,
      laiSuatQuaHan: +page.querySelector("#l-quahan").value,
      tuNgay: page.querySelector("#l-tu").value,
      denNgay: page.querySelector("#l-den").value,
      ngayDaoHan: page.querySelector("#l-daohan").value,
    };
    const r = await wcall("go.calc.Service.Lai", in_, "ui");
    page.querySelector("#l-result").innerHTML = `
      <div class="result-block">
        <div class="result-label">Số ngày</div><div class="result-value">${r.soNgay} ngày</div>
        <div class="result-label mt-4">Lãi trong hạn</div><div class="result-value">${money(r.tienLaiTrongHan)}</div>
        <div class="result-label mt-4">Lãi quá hạn</div><div class="result-value">${money(r.tienLaiQuaHan)}</div>
        <div class="result-label mt-4">Lãi chậm trả (150%)</div><div class="result-value">${money(r.tienLaiChamTra)}</div>
        <div class="result-label mt-4">Tổng tiền lãi</div><div class="result-value-lg">${money(r.tongTienLai)}</div>
        <div class="result-detail">${r.chiTiet}</div>
        <div class="result-ghichu">${r.ghiChu}</div>
      </div>`;
  };
});

// ──────────── Tuổi & thời hạn ────────────
route("tuoi", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Tính tuổi</div>
      <div class="form-grid">
        <div class="form-row"><label>Ngày sinh (YYYY-MM-DD)</label><input id="t-sinh" type="text" value="1990-05-20"></div>
        <div class="form-row"><label>Ngày xảy ra sự việc</label><input id="t-sv" type="text" value="2026-10-30"></div>
      </div>
      <button class="btn" id="t-calc">Tính tuổi</button>
      <div id="t-result"></div>
    </div>

    <div class="card">
      <div class="card-title">Tính thời hạn tố tụng</div>
      <div class="form-grid">
        <div class="form-row"><label>Loại</label>
          <select id="h-loai">
            <option value="tam_giu">Tạm giữ (theo giờ)</option>
            <option value="tam_giam">Tạm giam (theo tháng)</option>
            <option value="tra_ho_so">Trả hồ sơ</option>
            <option value="xet_xu">Xét xử</option>
          </select>
        </div>
        <div class="form-row"><label>Ngày bắt đầu (YYYY-MM-DD)</label><input id="h-tu" type="text" value="2026-01-15"></div>
        <div class="form-row"><label>Số tháng</label><input id="h-thang" type="number" value="4"></div>
        <div class="form-row"><label>HOẶC số ngày</label><input id="h-ngay" type="number" value="0"></div>
        <div class="form-row"><label>HOẶC số giờ (tạm giữ)</label><input id="h-gio" type="number" value="0"></div>
        <div class="form-row"><label>Giờ bắt đầu (nếu theo giờ)</label><input id="h-giobatdau" type="text" placeholder="2026-10-30T08:00"></div>
      </div>
      <button class="btn" id="h-calc">Tính thời hạn</button>
      <div id="h-result"></div>
    </div>
    <div class="callout">
      <b>BLTTHS 2015 (sửa 128/2025):</b> Tạm giữ 12h → gia hạn 24h → 48h → 72h; tạm giam 4 tháng (điều 119).
      <br><b>BLDS 2015 điều 35:</b> 1 tháng = 30 ngày.
    </div>
  `;
  page.querySelector("#t-calc").onclick = async () => {
    const in_ = {
      ngaySinh: page.querySelector("#t-sinh").value,
      ngayXayRaSuViec: page.querySelector("#t-sv").value,
    };
    const r = await wcall("go.calc.Service.Tuoi", in_, "ui");
    page.querySelector("#t-result").innerHTML = `
      <div class="result-block">
        <div class="result-label">Tuổi</div><div class="result-value-lg">${r.tuoi} năm ${r.tuoiThang} tháng ${r.tuoiNgay} ngày</div>
        ${r.daNghiAm ? `<div class="callout callout-warning">Dưới 16 tuổi — không chịu TNHS</div>` : ""}
        ${r.daNghiAm18 ? `<div class="callout callout-warning">Dưới 18 tuổi — áp dụng chế định người dưới 18 tuổi</div>` : ""}
        <div class="result-detail">${r.chiTiet}</div>
        <div class="result-ghichu">${r.ghiChu}</div>
      </div>`;
  };
  page.querySelector("#h-calc").onclick = async () => {
    const in_ = {
      loai: page.querySelector("#h-loai").value,
      ngayBatDau: page.querySelector("#h-tu").value,
      soThang: +page.querySelector("#h-thang").value,
      soNgay: +page.querySelector("#h-ngay").value,
      soGio: +page.querySelector("#h-gio").value,
      gioBatDau: page.querySelector("#h-giobatdau").value,
    };
    const r = await wcall("go.calc.Service.ThoiHan", in_, "ui");
    page.querySelector("#h-result").innerHTML = `
      <div class="result-block">
        <div class="result-label">Ngày bắt đầu</div><div class="result-value">${r.ngayBatDau}</div>
        <div class="result-label mt-4">Ngày/giờ kết thúc</div><div class="result-value-lg">${r.tinhTheoGio ? r.gioKetThuc : r.ngayKetThuc}</div>
        <div class="result-label mt-4">Số ngày thực tế</div><div class="result-value">${r.soNgayThucTe} ngày</div>
        <div class="result-detail">${r.chiTiet}</div>
        <div class="result-ghichu">${r.ghiChu}</div>
      </div>`;
  };
});

// ──────────── Thời hiệu TNHS ────────────
route("thoi-hieu", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Tính thời hiệu truy cứu trách nhiệm hình sự</div>
      <div class="form-grid">
        <div class="form-row"><label>Ngày xảy ra sự việc</label><input id="th-sv" type="text" value="2020-01-15"></div>
        <div class="form-row"><label>Loại tội phạm</label>
          <select id="th-loai">
            <option value="it_nghiem_trong">Ít nghiêm trọng (≤3 năm) — 5 năm</option>
            <option value="nghiem_trong">Nghiêm trọng (3-7 năm) — 10 năm</option>
            <option value="rat_nghiem_trong">Rất nghiêm trọng (7-15 năm) — 15 năm</option>
            <option value="dac_biet_nghiem_trong">Đặc biệt nghiêm trọng (>15 năm) — 20 năm</option>
          </select>
        </div>
      </div>
      <button class="btn" id="th-calc">Tính thời hiệu</button>
      <div id="th-result"></div>
    </div>
    <div class="callout">
      <b>BLHS 2015 (sửa 2017) điều 27:</b> Thời hiệu truy cứu TNHS tính từ ngày xảy ra tội phạm;
      nếu án đã có hiệu lực thì tính từ ngày án có hiệu lực.
    </div>
  `;
  page.querySelector("#th-calc").onclick = async () => {
    const in_ = {
      ngayXayRaSuViec: page.querySelector("#th-sv").value,
      loaiToiPham: page.querySelector("#th-loai").value,
    };
    const r = await wcall("go.calc.Service.ThoiHieuTNHS", in_, "ui");
    page.querySelector("#th-result").innerHTML = `
      <div class="result-block">
        <div class="result-label">Thời hiệu</div><div class="result-value-lg">${r.soNamThoiHieu} năm</div>
        <div class="result-label mt-4">Ngày hết hạn</div><div class="result-value">${r.ngayHetThoiHieu}</div>
        ${r.conThoiHieu
          ? `<div class="callout callout-success">Còn thời hiệu — ${r.conBaoNhieuNgay} ngày nữa</div>`
          : `<div class="callout callout-error">Đã hết thời hiệu</div>`}
        <div class="result-detail">${r.chiTiet}</div>
        <div class="result-ghichu">${r.ghiChu}</div>
      </div>`;
  };
});

// ──────────── Bồi thường NTNĐ ────────────
route("boi-thuong", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Tính bồi thường thiệt hại ngoài hợp đồng</div>
      <h4>Tài sản</h4>
      <div class="form-grid-3">
        <div class="form-row"><label>Tài sản mất</label><input id="bt-tsmat" type="number" value="0"></div>
        <div class="form-row"><label>Lợi nhuận mất</label><input id="bt-ln" type="number" value="0"></div>
        <div class="form-row"><label>Chi phí ngăn chặn</label><input id="bt-cp" type="number" value="0"></div>
      </div>
      <h4 class="mt-4">Sức khỏe</h4>
      <div class="form-grid-3">
        <div class="form-row"><label>Chi phí y tế</label><input id="bt-yte" type="number" value="0"></div>
        <div class="form-row"><label>Chi phí y tế tương lai</label><input id="bt-ytetl" type="number" value="0"></div>
        <div class="form-row"><label>Thu nhập mất</label><input id="bt-tnm" type="number" value="0"></div>
        <div class="form-row"><label>Thu nhập mất (TL)</label><input id="bt-tnmtl" type="number" value="0"></div>
      </div>
      <h4 class="mt-4">Tính mạng & danh dự</h4>
      <div class="form-grid-3">
        <div class="form-row"><label>Chi phí tang lễ</label><input id="bt-tangle" type="number" value="0"></div>
        <div class="form-row"><label>Tiền cấp dưỡng</label><input id="bt-cd" type="number" value="0"></div>
        <div class="form-row"><label>BT danh dự</label><input id="bt-dd" type="number" value="0"></div>
      </div>
      <button class="btn mt-4" id="bt-calc">Tính bồi thường</button>
      <div id="bt-result"></div>
    </div>
    <div class="callout">
      <b>BLDS 2015 điều 589-609:</b> Bồi thường thiệt hại ngoài hợp đồng — TS + SK + TM + DD.
    </div>
  `;
  page.querySelector("#bt-calc").onclick = async () => {
    const in_ = {
      taiSanMat: +page.querySelector("#bt-tsmat").value,
      taiSanLoiNhuan: +page.querySelector("#bt-ln").value,
      taiSanChiPhiuKhamPha: +page.querySelector("#bt-cp").value,
      chiPhiYTe: +page.querySelector("#bt-yte").value,
      chiPhiYTeTuongLai: +page.querySelector("#bt-ytetl").value,
      thuNhapMat: +page.querySelector("#bt-tnm").value,
      thuNhapMatTuongLai: +page.querySelector("#bt-tnmtl").value,
      chiPhiTangLe: +page.querySelector("#bt-tangle").value,
      tienCapDuong: +page.querySelector("#bt-cd").value,
      tienBoiThuongDanhDu: +page.querySelector("#bt-dd").value,
    };
    const r = await wcall("go.calc.Service.BoiThuongNT", in_, "ui");
    page.querySelector("#bt-result").innerHTML = `
      <div class="result-block">
        <div class="result-label">Tổng TS</div><div class="result-value">${money(r.tongThietHaiTaiSan)}</div>
        <div class="result-label mt-4">Tổng SK</div><div class="result-value">${money(r.tongThietHaiSucKhoe)}</div>
        <div class="result-label mt-4">Tổng TM</div><div class="result-value">${money(r.tongThietHaiTinhMang)}</div>
        <div class="result-label mt-4">Danh dự</div><div class="result-value">${money(r.tienBoiThuongDanhDu)}</div>
        <div class="result-label mt-4">Tổng bồi thường</div><div class="result-value-lg">${money(r.tongTienBoiThuong)}</div>
        <div class="result-detail">${r.chiTiet}</div>
        <div class="result-ghichu">${r.ghiChu}</div>
      </div>`;
  };
});

// ──────────── Legal lookup ────────────
route("legal", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Tra cứu văn bản pháp luật (offline) — Smart Search</div>
      <p class="text-muted mb-4">Hỗ trợ: <b>"điều 27"</b> (tìm theo số điều), <b>"326/2016"</b> (số hiệu), <b>"án phí"</b> / <b>"thời hiệu"</b> (từ khóa nội dung)</p>
      <div class="form-grid">
        <div class="form-row"><label>Từ khóa</label>
          <input id="lg-q" type="text" placeholder='vd: điều 27, 326/2016, "án phí", thời hiệu' value="điều 27">
        </div>
        <div class="form-row"><label>Lĩnh vực</label>
          <select id="lg-lv">
            <option value="">Tất cả</option>
            <option value="hinh_su">Hình sự</option>
            <option value="dan_su">Dân sự</option>
            <option value="thi_hanh_an">Thi hành án</option>
            <option value="khieu_nai">Khiếu nại tố cáo</option>
            <option value="khac">Khác</option>
          </select>
        </div>
      </div>
      <button class="btn" id="lg-search">🔍 Tìm kiếm</button>
      <div id="lg-results" class="mt-4"></div>
    </div>

    <div class="card">
      <div class="card-title">Cơ sở dữ liệu luật hiện có</div>
      <div id="lg-list"></div>
    </div>
  `;

  // List all legal docs on initial load
  const allDocs = await wcall("go.legal.Service.Search", "", "", 100);
  const listHtml = (allDocs || []).map(d => `
    <div class="card" style="margin-bottom:8px; padding:12px;">
      <span class="badge" style="background:#3586ae; color:white;">${d.loai}</span>
      <b style="margin-left:8px">${d.soHieu}</b>
      <span class="text-muted" style="margin-left:8px">(${d.ngayBanHanh})</span>
      <div class="mt-2">${d.trichYeu}</div>
      <button class="btn btn-sm mt-2" data-id="${d.id}">📖 Xem cấu trúc Điều/Khoản</button>
    </div>
  `).join("");
  page.querySelector("#lg-list").innerHTML = listHtml || "<p class='text-muted'>Chưa có văn bản.</p>";

  // Wire "Xem cấu trúc" buttons
  page.querySelectorAll("#lg-list button[data-id]").forEach(b => {
    b.onclick = async () => {
      const id = +b.dataset.id;
      const doc = await wcall("go.legal.Service.GetFullByID", id);
      if (!doc) { alert("Không load được"); return; }
      renderLegalDocModal(doc);
    };
  });

  // Search button
  page.querySelector("#lg-search").onclick = async () => {
    const q = page.querySelector("#lg-q").value;
    const lv = page.querySelector("#lg-lv").value;
    const result = await wcall("go.legal.Service.SmartSearch", q, lv, 50);
    renderSmartSearchResults(page.querySelector("#lg-results"), result);
  };

  // Auto-trigger search on load
  page.querySelector("#lg-search").click();
});

function renderSmartSearchResults(container, result) {
  if (!result) {
    container.innerHTML = `<div class="callout callout-warning">Không có kết quả.</div>`;
    return;
  }
  let html = `<div class="callout">Loại query: <b>${result.queryType}</b></div>`;

  // Dieu hits
  if (result.dieuHits && result.dieuHits.length > 0) {
    html += `<h3>Điều khớp (${result.dieuHits.length})</h3>`;
    for (const hit of result.dieuHits) {
      html += renderDieuCard(hit);
    }
  }

  // Khoan hits
  if (result.khoanHits && result.khoanHits.length > 0) {
    html += `<h3>Khoản/Điểm khớp (${result.khoanHits.length})</h3>`;
    for (const hit of result.khoanHits) {
      html += renderKhoanCard(hit);
    }
  }

  // Docs (flat)
  if (result.docs && result.docs.length > 0) {
    html += `<h3>Văn bản khớp (${result.docs.length})</h3>`;
    for (const d of result.docs) {
      html += `
        <div class="card" style="margin-bottom:8px; padding:12px;">
          <span class="badge" style="background:#3586ae; color:white;">${d.loai}</span>
          <b style="margin-left:8px">${d.soHieu}</b>
          <span class="text-muted" style="margin-left:8px">(${d.ngayBanHanh})</span>
          <div class="mt-2">${d.trichYeu}</div>
          <div class="text-muted mt-2">${d.tomTat}</div>
          <button class="btn btn-sm mt-2" onclick="showLegalDoc(${d.id})">📖 Xem cấu trúc</button>
        </div>
      `;
    }
  }

  if (!html.includes("khớp")) {
    html = `<div class="callout callout-warning">Không tìm thấy kết quả cho "<b>${result.query}</b>".</div>`;
  }
  container.innerHTML = html;
}

function renderDieuCard(hit) {
  return `
    <div class="card" style="margin-bottom:8px; padding:12px; border-left: 4px solid #3586ae;">
      <div class="text-muted">${hit.docTrichYeu} — ${hit.docSoHieu}</div>
      <h4 style="margin: 4px 0;">Điều ${hit.dieu.soDieu}: ${hit.dieu.tenDieu}</h4>
      ${renderKhoanList(hit.dieu.khoanList)}
    </div>
  `;
}

function renderKhoanCard(hit) {
  return `
    <div class="card" style="margin-bottom:8px; padding:12px; border-left: 4px solid #c66646;">
      <div class="text-muted">${hit.docTrichYeu} — ${hit.docSoHieu}</div>
      <div class="text-muted">Điều ${hit.soDieu}: ${hit.tenDieu}</div>
      ${renderKhoanList([hit.khoan])}
    </div>
  `;
}

function renderKhoanList(khoanList) {
  if (!khoanList) return "";
  let html = "";
  for (const k of khoanList) {
    html += `<p style="margin: 6px 0;"><b>${k.soKhoan}.</b> ${k.noiDung}</p>`;
    if (k.diemList && k.diemList.length > 0) {
      for (const d of k.diemList) {
        html += `<p style="margin: 4px 0 4px 24px;"><b>${d.soDiem})</b> ${d.noiDung}</p>`;
      }
    }
  }
  return html;
}

function showLegalDoc(id) {
  wcall("go.legal.Service.GetFullByID", id).then(doc => {
    if (doc) renderLegalDocModal(doc);
  });
}

window.showLegalDoc = showLegalDoc;

function renderLegalDocModal(doc) {
  // Render to a modal overlay
  let modal = document.getElementById("legal-modal");
  if (!modal) {
    modal = document.createElement("div");
    modal.id = "legal-modal";
    modal.className = "modal";
    modal.innerHTML = `
      <div class="modal-backdrop"></div>
      <div class="modal-content" style="max-width: 900px;">
        <button class="modal-close" onclick="document.getElementById('legal-modal').hidden=true">&times;</button>
        <div id="legal-modal-body" style="padding: 32px; overflow-y:auto; max-height:92vh;"></div>
      </div>
    `;
    modal.querySelector(".modal-backdrop").onclick = () => modal.hidden = true;
    document.body.appendChild(modal);
  }
  let html = `
    <h2 style="margin-bottom: 8px;">${doc.loai} ${doc.soHieu}</h2>
    <p class="text-muted">${doc.coQuanBanHanh} — ${doc.ngayBanHanh}</p>
    <h3 style="margin-top: 16px;">${doc.trichYeu}</h3>
    <p>${doc.tomTat}</p>
    <hr>
    <h3 style="margin-top: 16px;">Nội dung (${doc.dieuList.length} điều)</h3>
  `;
  for (const dieu of doc.dieuList) {
    html += `
      <div style="margin: 16px 0; padding: 12px; background: #f4f4f5; border-left: 4px solid #3586ae; border-radius: 4px;">
        <h4 style="margin: 0 0 8px;">Điều ${dieu.soDieu}: ${dieu.tenDieu}</h4>
        ${renderKhoanList(dieu.khoanList)}
      </div>
    `;
  }
  modal.querySelector("#legal-modal-body").innerHTML = html;
  modal.hidden = false;
}

// ──────────── Spell check ────────────
route("spell", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Rà soát chính tả văn bản</div>
      <div class="form-row">
        <label>Văn bản cần kiểm tra</label>
        <textarea id="sp-text" rows="8" style="font-family:monospace;">Ông thi han an truong hop nhung cua  nguoi  ,  chi  thi han,.,. a^ a( o^</textarea>
      </div>
      <div class="flex mt-2">
        <button class="btn" id="sp-check">🔍 Rà soát</button>
        <button class="btn btn-ghost" id="sp-copy-fixed">📋 Copy văn bản đã sửa</button>
      </div>
      <div id="sp-result" class="mt-4"></div>
    </div>
    <div class="callout">
      <b>Phát hiện:</b> telex chưa chuyển (a^ â, a( ă, u* ư, o^ ô, o* ơ, e^ ê, d/ đ),
      double whitespace, double punct (,, .. !! ??), space trước punct, term_norm
      (thi han an → thi hành án), literal typos.
    </div>
  `;
  page.querySelector("#sp-check").onclick = async () => {
    const text = page.querySelector("#sp-text").value;
    const r = await wcall("go.spell.Service.Check", text, "ui");
    if (!r) return;
    const cats = Object.entries(r.countByCat || {}).map(([k, v]) =>
      `<span class="badge" style="background:#e9ebec; color:#3e525c; margin-right: 6px;">${k}: ${v}</span>`
    ).join("");
    page.querySelector("#sp-result").innerHTML = `
      <div class="result-block">
        <div class="result-label">Phát hiện ${r.totalCount} lỗi</div>
        <div class="mt-2">${cats}</div>
        <div class="result-detail" style="white-space:pre-wrap; max-height:200px; overflow-y:auto; margin-top: 12px;">
          <b>Văn bản đánh dấu:</b><br>${escapeHtml(r.markedText)}
        </div>
        <div class="result-detail" style="white-space:pre-wrap; max-height:300px; overflow-y:auto; margin-top: 16px; background: #eaf3ed; border-left: 4px solid #437453;">
          <b>✓ Văn bản đã sửa hoàn thiện:</b><br>${escapeHtml(r.fixedText)}
        </div>
        <div class="text-muted mt-2">
          ${r.stats.issuesFixed} lỗi đã sửa · ${r.stats.originalChars} → ${r.stats.fixedChars} ký tự
        </div>
      </div>`;
  };
  page.querySelector("#sp-copy-fixed").onclick = async () => {
    const text = page.querySelector("#sp-text").value;
    const r = await wcall("go.spell.Service.Check", text, "ui");
    if (!r) return;
    navigator.clipboard.writeText(r.fixedText).then(() => {
      alert("Đã copy văn bản đã sửa vào clipboard!");
    });
  };
});

function escapeHtml(s) {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

// ──────────── PII ────────────
route("pii", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Che thông tin cá nhân (PII)</div>
      <div class="form-grid">
        <div class="form-row"><label>Chế độ thay thế</label>
          <select id="pi-mode">
            <option value="mask">Mask (*** )</option>
            <option value="token">Token ([CCCD])</option>
            <option value="remove">Remove (xóa hoàn toàn)</option>
          </select>
        </div>
        <div class="form-row"><label>Giữ ký tự đầu?</label>
          <select id="pi-keep"><option value="0">Không</option><option value="1">Có (N*****)</option></select>
        </div>
      </div>
      <div class="form-row">
        <label>Văn bản</label>
        <textarea id="pi-text" rows="6">Ông Nguyễn Văn An, CCCD 012345678901, SDT 0912345678, email an@gmail.com, sinh 15/03/1990, STK 9876543210 tại BIDV, MST 0123456789</textarea>
      </div>
      <button class="btn" id="pi-mask">Che thông tin</button>
      <div id="pi-result" class="mt-4"></div>
    </div>
  `;
  page.querySelector("#pi-mask").onclick = async () => {
    const in_ = {
      replaceMode: page.querySelector("#pi-mode").value,
      keepFirst: page.querySelector("#pi-keep").value === "1",
    };
    const r = await wcall("go.pii.Service.Mask", page.querySelector("#pi-text").value, in_, "ui");
    const cats = Object.entries(r.countByCat || {}).map(([k, v]) => `${k}: ${v}`).join(", ");
    page.querySelector("#pi-result").innerHTML = `
      <div class="result-block">
        <div class="result-label">Đã che ${r.totalCount} vị trí</div>
        <div class="text-muted">${cats}</div>
        <div class="result-detail" style="white-space:pre-wrap;">${r.text}</div>
      </div>`;
  };
});

// ──────────── OCR ────────────
route("ocr", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">OCR PDF/ảnh (Tesseract)</div>
      <div class="form-grid">
        <div class="form-row"><label>Đường dẫn file</label>
          <input id="o-path" type="text" placeholder="/abs/path/to/file.pdf hoặc .png/.jpg">
        </div>
        <div class="form-row"><label>Ngôn ngữ</label>
          <select id="o-lang"><option value="vie">Vietnamese</option><option value="vie+eng">VIE + ENG</option><option value="eng">English</option></select>
        </div>
        <div class="form-row"><label>Định dạng xuất</label>
          <select id="o-fmt"><option value="txt">TXT</option><option value="md">Markdown</option></select>
        </div>
        <div class="form-row"><label>DPI (cho PDF)</label><input id="o-dpi" type="number" value="300"></div>
      </div>
      <button class="btn" id="o-run">Chạy OCR</button>
      <div id="o-result" class="mt-4"></div>
    </div>
    <div class="callout">
      <b>Yêu cầu hệ thống:</b> <code>tesseract</code> + <code>vietnamese.traineddata</code> (Linux: <code>apt install tesseract-ocr tesseract-ocr-vie</code>; Windows: tải tesseract.exe).<br>
      Cho PDF cần thêm <code>poppler-utils</code> (<code>pdftoppm</code>) hoặc <code>imagemagick</code>.
    </div>
  `;
  page.querySelector("#o-run").onclick = async () => {
    const in_ = {
      lang: page.querySelector("#o-lang").value,
      outputFmt: page.querySelector("#o-fmt").value,
      dpi: +page.querySelector("#o-dpi").value,
    };
    page.querySelector("#o-result").innerHTML = "<div class='text-muted'>Đang OCR...</div>";
    const r = await wcall("go.ocr.Service.Recognize", page.querySelector("#o-path").value, in_, "ui");
    if (!r || r.pageCount === 0) {
      page.querySelector("#o-result").innerHTML = `<div class="callout callout-error">OCR thất bại. ${(r?.warnings||[]).join("; ")||"Kiểm tra file/path/deps."}</div>`;
      return;
    }
    page.querySelector("#o-result").innerHTML = `
      <div class="result-block">
        <div class="result-label">Trang: ${r.pageCount} · Conf: ${r.confidence.toFixed(1)}% · Thời gian: ${r.duration}</div>
        <div class="text-muted">Xuất ra: ${r.outputPath}</div>
        <div class="result-detail" style="white-space:pre-wrap; max-height:300px; overflow-y:auto;">${r.text}</div>
        ${(r.warnings||[]).map(w => `<div class="callout callout-warning">${w}</div>`).join("")}
      </div>`;
  };
});

// ──────────── File rename ────────────
route("file", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Đổi tên file hàng loạt</div>
      <div class="form-grid">
        <div class="form-row"><label>Thư mục</label><input id="f-dir" type="text" placeholder="/path/to/folder"></div>
        <div class="form-row"><label>Chế độ</label>
          <select id="f-mode">
            <option value="number">Đánh số</option>
            <option value="replace">Thay thế text</option>
            <option value="diacritics_strip">Bỏ dấu tiếng Việt</option>
            <option value="prefix">Thêm tiền tố</option>
            <option value="suffix">Thêm hậu tố</option>
            <option value="lowercase">Chuyển thường</option>
          </select>
        </div>
        <div class="form-row"><label>Pattern (cho "number")</label><input id="f-pattern" type="text" placeholder="{n}_{base}"></div>
        <div class="form-row"><label>Find (cho "replace")</label><input id="f-find" type="text"></div>
        <div class="form-row"><label>Replace (cho "replace")</label><input id="f-replace" type="text"></div>
        <div class="form-row"><label>Prefix/Suffix</label><input id="f-prefix" type="text"></div>
        <div class="form-row"><label>Start index (cho "number")</label><input id="f-start" type="number" value="1"></div>
        <div class="form-row"><label>Pad width</label><input id="f-pad" type="number" value="3"></div>
        <div class="form-row"><label>Lọc theo extension</label><input id="f-ext" type="text" placeholder=".pdf,.docx"></div>
        <div class="form-row"><label>Sắp xếp theo</label>
          <select id="f-sort"><option value="name">Tên</option><option value="mtime">mtime</option><option value="size">size</option></select>
        </div>
      </div>
      <label><input type="checkbox" id="f-dry" checked> Dry run (chỉ xem trước)</label>
      <button class="btn mt-4" id="f-apply">Áp dụng</button>
      <div id="f-result" class="mt-4"></div>
    </div>
  `;
  page.querySelector("#f-apply").onclick = async () => {
    const spec = {
      dir: page.querySelector("#f-dir").value,
      mode: page.querySelector("#f-mode").value,
      pattern: page.querySelector("#f-pattern").value,
      replaceFind: page.querySelector("#f-find").value,
      replaceWith: page.querySelector("#f-replace").value,
      prefix: page.querySelector("#f-prefix").value,
      suffix: page.querySelector("#f-prefix").value,
      startIndex: +page.querySelector("#f-start").value,
      padWidth: +page.querySelector("#f-pad").value,
      includeExt: page.querySelector("#f-ext").value,
      sortBy: page.querySelector("#f-sort").value,
      dryRun: page.querySelector("#f-dry").checked,
    };
    const r = await wcall("go.file.Service.Apply", spec, "ui");
    const rows = (r.preview || []).slice(0, 50).map(p => `
      <tr><td>${p.oldPath}</td><td>→</td><td>${p.newPath}</td><td>${p.status}</td></tr>
    `).join("");
    page.querySelector("#f-result").innerHTML = `
      <div class="result-block">
        <div class="result-label">Tổng: ${r.total} · Đã đổi: ${r.renamed} · Bỏ qua: ${r.skipped}</div>
        <table><thead><tr><th>Cũ</th><th></th><th>Mới</th><th>Trạng thái</th></tr></thead><tbody>${rows}</tbody></table>
      </div>`;
  };
});

// ──────────── Cases ────────────
route("cases", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Hồ sơ vụ án — CRUD đầy đủ</div>
      <div class="flex mb-4">
        <button class="btn" id="c-new">+ Vụ án mới</button>
        <select id="c-filter-type" class="ml-2">
          <option value="">Tất cả loại</option>
          <option value="HS">Hình sự</option>
          <option value="DS">Dân sự</option>
          <option value="HC">Hành chính</option>
          <option value="THADS">THA dân sự</option>
          <option value="THAHS">THA hình sự</option>
          <option value="KN-TC">Khiếu nại tố cáo</option>
          <option value="KKCI">Khởi kiện công ích</option>
        </select>
        <select id="c-filter-status" class="ml-2">
          <option value="">Tất cả trạng thái</option>
          <option value="mo">Mở</option>
          <option value="dong">Đóng</option>
          <option value="tam_ngung">Tạm ngưng</option>
        </select>
        <button class="btn btn-ghost ml-2" id="c-refresh">🔄 Tải lại</button>
      </div>
      <div id="c-list"></div>
    </div>
    <div id="c-form"></div>
  `;

  async function render() {
    const type = page.querySelector("#c-filter-type").value;
    const status = page.querySelector("#c-filter-status").value;
    const cases = await wcall("go.casefile.Service.List", type, status, 200);
    const list = page.querySelector("#c-list");
    if (!cases || cases.length === 0) {
      list.innerHTML = `<p class="text-muted">Chưa có vụ án nào. Bấm "+ Vụ án mới" để tạo.</p>`;
      return;
    }
    list.innerHTML = `<table><thead><tr>
      <th>Số vụ án</th><th>Loại</th><th>Giai đoạn</th><th>KSV</th><th>Ngày thụ lý</th><th>Deadline</th><th>Trạng thái</th><th></th>
    </tr></thead><tbody>` +
      cases.map(c => {
        const dlClass = c.deadline && new Date(c.deadline) < new Date(Date.now() + 7*24*60*60*1000) ? 'style="color: #c66646; font-weight: bold;"' : '';
        return `<tr>
          <td><b>${c.caseNumber}</b></td>
          <td><span class="badge" style="background:#e9ebec; color:#3e525c;">${c.caseType}</span></td>
          <td>${c.stage}</td>
          <td>${c.assignedKsv || "—"}</td>
          <td>${c.receivedDate || "—"}</td>
          <td ${dlClass}>${c.deadline || "—"}</td>
          <td>${c.status === 'mo' ? '🟢 Mở' : c.status === 'dong' ? '⚫ Đóng' : '🟡 Tạm ngưng'}</td>
          <td>
            <button class="btn btn-sm" data-id="${c.id}" data-act="edit">✍ Sửa</button>
            <button class="btn btn-sm" data-id="${c.id}" data-act="status">Đổi trạng thái</button>
            ${c.status === 'mo' ? `<button class="btn btn-sm btn-danger" data-id="${c.id}" data-act="close">Đóng</button>` : ''}
          </td>
        </tr>`;
      }).join("") +
      `</tbody></table>`;
    list.querySelectorAll("button[data-id]").forEach(b => {
      b.onclick = async () => {
        const id = b.dataset.id;
        const act = b.dataset.act;
        if (act === "edit") showEditForm(id);
        else if (act === "status") {
          const newStatus = prompt("Trạng thái mới (mo/dong/tam_ngung):", "mo");
          if (newStatus) {
            await wcall("go.casefile.Service.SetStatus", id, newStatus, "ui");
            render();
          }
        } else if (act === "close") {
          if (confirm("Đóng vụ án này? (soft delete — vẫn còn trong DB)")) {
            await wcall("go.casefile.Service.Delete", id, "ui");
            render();
          }
        }
      };
    });
  }

  function showEditForm(id) {
    wcall("go.casefile.Service.Get", id).then(c => {
      if (!c) return;
      const form = page.querySelector("#c-form");
      form.innerHTML = caseFormHTML(c, "Sửa vụ án", async (data) => {
        data.id = id;
        await wcall("go.casefile.Service.Update", data, "ui");
        form.innerHTML = "";
        render();
      });
    });
  }

  page.querySelector("#c-new").onclick = () => {
    const form = page.querySelector("#c-form");
    form.innerHTML = caseFormHTML({
      caseNumber: "", caseType: "DS", stage: "thu_ly", assignedKsv: "",
      receivedDate: new Date().toISOString().slice(0, 10), deadline: "",
      parties: [], notes: "", status: "mo"
    }, "Tạo vụ án mới", async (data) => {
      const created = await wcall("go.casefile.Service.Create", data, "ui");
      if (created) {
        form.innerHTML = "";
        render();
      } else {
        alert("Lỗi khi tạo vụ án");
      }
    });
  };
  page.querySelector("#c-refresh").onclick = render;
  page.querySelector("#c-filter-type").onchange = render;
  page.querySelector("#c-filter-status").onchange = render;
  render();
});

function caseFormHTML(c, title, onSubmit) {
  const div = document.createElement("div");
  div.innerHTML = `
    <div class="card">
      <div class="card-title">${title}</div>
      <div class="form-grid">
        <div class="form-row"><label>Số vụ án</label><input id="cf-num" type="text" value="${c.caseNumber || ''}" placeholder="QĐ 12/2026/TLPT-DS"></div>
        <div class="form-row"><label>Loại</label>
          <select id="cf-type">
            ${['HS','DS','HC','THADS','THAHS','KN-TC','KKCI'].map(t =>
              `<option value="${t}" ${c.caseType === t ? 'selected' : ''}>${t}</option>`).join('')}
          </select>
        </div>
        <div class="form-row"><label>Giai đoạn</label>
          <select id="cf-stage">
            ${['thu_ly','dieu_tra','truy_to','xet_xu','thi_hanh'].map(s =>
              `<option value="${s}" ${c.stage === s ? 'selected' : ''}>${s}</option>`).join('')}
          </select>
        </div>
        <div class="form-row"><label>KSV phụ trách</label><input id="cf-ksv" type="text" value="${c.assignedKsv || ''}"></div>
        <div class="form-row"><label>Ngày thụ lý</label><input id="cf-rec" type="date" value="${c.receivedDate || ''}"></div>
        <div class="form-row"><label>Deadline</label><input id="cf-dl" type="date" value="${c.deadline || ''}"></div>
      </div>
      <div class="form-row"><label>Đương sự (phẩy phân tách)</label>
        <input id="cf-parties" type="text" value="${(c.parties || []).join(', ')}" placeholder="Nguyễn A, Công ty X">
      </div>
      <div class="form-row"><label>Ghi chú</label>
        <textarea id="cf-notes" rows="3">${c.notes || ''}</textarea>
      </div>
      <div class="flex mt-4">
        <button class="btn" id="cf-save">💾 Lưu</button>
        <button class="btn btn-ghost" id="cf-cancel">Hủy</button>
      </div>
    </div>
  `;
  div.querySelector("#cf-cancel").onclick = () => { div.innerHTML = ""; };
  div.querySelector("#cf-save").onclick = () => {
    const data = {
      caseNumber: div.querySelector("#cf-num").value,
      caseType: div.querySelector("#cf-type").value,
      stage: div.querySelector("#cf-stage").value,
      assignedKsv: div.querySelector("#cf-ksv").value,
      receivedDate: div.querySelector("#cf-rec").value,
      deadline: div.querySelector("#cf-dl").value,
      parties: div.querySelector("#cf-parties").value.split(",").map(s => s.trim()).filter(s => s),
      notes: div.querySelector("#cf-notes").value,
      status: c.status || "mo",
    };
    onSubmit(data);
  };
  return div;
}

// ──────────── Evidence ────────────
route("evidence", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Gắn nhãn chứng cứ — Form thêm trực quan</div>
      <div class="form-grid">
        <div class="form-row"><label>Vụ án (case ID)</label><input id="e-case" type="text" placeholder="case_xxx"></div>
        <div class="form-row"><label>Nhãn (vd: CT-01)</label><input id="e-label" type="text" placeholder="CT-01"></div>
        <div class="form-row"><label>Màu nhãn</label><input id="e-color" type="color" value="#ff0000"></div>
        <div class="form-row"><label>Loại file</label>
          <select id="e-type">
            <option value="pdf">PDF</option>
            <option value="image">Ảnh</option>
            <option value="video">Video</option>
            <option value="audio">Audio</option>
            <option value="docx">Word</option>
            <option value="other">Khác</option>
          </select>
        </div>
        <div class="form-row" style="grid-column: span 2;"><label>Đường dẫn file</label>
          <input id="e-path" type="text" placeholder="/abs/path/to/file.pdf">
        </div>
        <div class="form-row"><label>Marker trang/thời gian</label><input id="e-marker" type="text" placeholder="p12 hoặc video@00:14:30"></div>
        <div class="form-row"><label>Tags (phẩy phân tách)</label><input id="e-tags" type="text" placeholder="quan_trong, khung_1"></div>
        <div class="form-row" style="grid-column: span 2;"><label>Mô tả</label>
          <textarea id="e-caption" rows="2" placeholder="Bản khai của nhân chứng..."></textarea>
        </div>
      </div>
      <button class="btn" id="e-add">+ Thêm chứng cứ</button>
    </div>
    <div class="card">
      <div class="card-title">Danh sách chứng cứ theo vụ án</div>
      <div class="flex mb-4">
        <input id="e-list-case" type="text" placeholder="case ID cần xem" style="width: 300px;">
        <button class="btn btn-ghost" id="e-list-btn">📋 Liệt kê</button>
      </div>
      <div id="e-list-out"></div>
    </div>
  `;
  page.querySelector("#e-add").onclick = async () => {
    const ev = {
      caseId: page.querySelector("#e-case").value,
      label: page.querySelector("#e-label").value,
      labelColor: page.querySelector("#e-color").value,
      filePath: page.querySelector("#e-path").value,
      fileType: page.querySelector("#e-type").value,
      pageMarker: page.querySelector("#e-marker").value,
      caption: page.querySelector("#e-caption").value,
      tags: page.querySelector("#e-tags").value.split(",").map(s => s.trim()).filter(s => s),
    };
    const r = await wcall("go.evidence.Service.Add", ev, "ui");
    if (r) {
      alert(`✓ Đã thêm chứng cứ: ${r.label} (ID: ${r.id})`);
      page.querySelector("#e-list-case").value = r.caseId;
      page.querySelector("#e-list-btn").click();
    }
  };
  page.querySelector("#e-list-btn").onclick = async () => {
    const caseId = page.querySelector("#e-list-case").value;
    if (!caseId) { alert("Nhập case ID"); return; }
    const items = await wcall("go.evidence.Service.ListByCase", caseId);
    const tree = await wcall("go.evidence.Service.TreeByCase", caseId);
    const out = page.querySelector("#e-list-out");
    if (!items || items.length === 0) {
      out.innerHTML = "<p class='text-muted'>Không có chứng cứ.</p>";
      return;
    }
    let html = `<p class="text-muted mb-2">${items.length} chứng cứ, gom thành ${tree.length} nhóm:</p>`;
    for (const node of tree) {
      html += `<h4 style="margin-top: 16px;">📁 ${node.category} (${node.items.length})</h4>`;
      html += `<table><thead><tr><th>Nhãn</th><th>File</th><th>Marker</th><th>Mô tả</th><th></th></tr></thead><tbody>`;
      for (const it of node.items) {
        html += `<tr>
          <td><span style="background:${it.labelColor}; color:white; padding:2px 8px; border-radius:3px;">${it.label}</span></td>
          <td><code>${it.filePath}</code></td>
          <td>${it.pageMarker || '—'}</td>
          <td>${it.caption || '—'}</td>
          <td><button class="btn btn-sm btn-danger" data-id="${it.id}">Xóa</button></td>
        </tr>`;
      }
      html += `</tbody></table>`;
    }
    out.innerHTML = html;
    out.querySelectorAll("button[data-id]").forEach(b => {
      b.onclick = async () => {
        if (confirm("Xóa chứng cứ này? (file gốc KHÔNG bị xóa)")) {
          await wcall("go.evidence.Service.Delete", b.dataset.id, "ui");
          page.querySelector("#e-list-btn").click();
        }
      };
    });
  };
});

// ──────────── Docs templates ────────────
route("docs", async (page) => {
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Thư viện biểu mẫu tố tụng — Soạn thảo ra file Word</div>
      <p class="text-muted mb-4">Bấm <b>Soạn</b> → điền form → xuất file <code>.docx</code> Word thật (mở bằng MS Word/WPS).</p>
      <div class="form-grid">
        <div class="form-row">
          <label>Lọc theo lĩnh vực</label>
          <select id="d-lv-filter">
            <option value="">Tất cả</option>
            <option value="hinh_su">Hình sự</option>
            <option value="dan_su">Dân sự</option>
            <option value="thi_hanh_an">Thi hành án</option>
            <option value="khieu_nai">Khiếu nại tố cáo</option>
            <option value="khac">Khác</option>
          </select>
        </div>
      </div>
      <div id="d-list">Đang tải...</div>
    </div>
    <div id="d-form"></div>
  `;

  async function loadList() {
    const lv = page.querySelector("#d-lv-filter").value;
    const list = await wcall("go.docs.Service.ListTemplates", lv, "");
    const container = page.querySelector("#d-list");
    if (!list || list.length === 0) {
      container.innerHTML = "<p class='text-muted'>Không có biểu mẫu.</p>";
      return;
    }
    container.innerHTML = `<table><thead><tr>
      <th>Mã mẫu</th><th>Tên mẫu</th><th>Lĩnh vực</th><th>Giai đoạn</th><th></th>
    </tr></thead><tbody>` +
      list.map(t => `<tr>
        <td><b>${t.maMau}</b></td>
        <td>${t.tenMau}</td>
        <td>${t.linhVuc}</td>
        <td>${t.giaiDoan}</td>
        <td><button class="btn btn-sm" data-id="${t.id}" data-ghichu="${t.ghiChu || ''}">✍ Soạn</button></td>
      </tr>`).join("") +
      `</tbody></table>`;
    container.querySelectorAll("button[data-id]").forEach(b => {
      b.onclick = async () => {
        const id = +b.dataset.id;
        showTemplateForm(id, b.dataset.ghichu);
      };
    });
  }
  loadList();
  page.querySelector("#d-lv-filter").onchange = loadList;

  function showTemplateForm(templateId, ghiChu) {
    const form = page.querySelector("#d-form");
    // Parse ghiChu để lấy list of fields
    // Format: "Mẫu ... — điền field1, field2, field3"
    const fieldsMatch = ghiChu.match(/điền\s+(.+)$/i);
    let fields = [];
    if (fieldsMatch) {
      fields = fieldsMatch[1].split(",").map(s => s.trim()).filter(s => s);
    }
    if (fields.length === 0) {
      // Fallback: default common fields
      fields = ["case_number", "today", "ksv"];
    }

    form.innerHTML = `
      <div class="card">
        <div class="card-title">Điền thông tin biểu mẫu</div>
        <p class="text-muted mb-4">Các trường cần điền (có thể để trống, sẽ giữ nguyên placeholder):</p>
        <div class="form-grid" id="d-fields">
          ${fields.map(f => `
            <div class="form-row">
              <label>${f}</label>
              <input data-field="${f}" type="text" value="${f === 'today' ? new Date().toLocaleDateString('vi-VN') : ''}">
            </div>
          `).join("")}
        </div>
        <div class="flex mt-4">
          <button class="btn" id="d-gen-docx">📄 Xuất file Word (.docx)</button>
          <button class="btn btn-ghost" id="d-gen-html">🌐 Xuất HTML</button>
          <button class="btn btn-ghost" id="d-cancel">Hủy</button>
        </div>
        <div id="d-output" class="mt-4"></div>
      </div>
    `;
    page.querySelector("#d-cancel").onclick = () => { form.innerHTML = ""; };
    page.querySelector("#d-gen-docx").onclick = async () => {
      const fieldsObj = {};
      page.querySelectorAll("#d-fields input[data-field]").forEach(inp => {
        fieldsObj[inp.dataset.field] = inp.value;
      });
      page.querySelector("#d-output").innerHTML = "<div class='text-muted'>Đang sinh file .docx...</div>";
      const r = await wcall("go.docs.Service.Fill", {
        templateId, fields: fieldsObj, outputFmt: "docx"
      }, "ui");
      if (!r) {
        page.querySelector("#d-output").innerHTML = "<div class='callout callout-error'>Lỗi khi sinh file.</div>";
        return;
      }
      page.querySelector("#d-output").innerHTML = `
        <div class="callout callout-success">
          <b>✓ Đã xuất file Word thành công!</b><br>
          Đường dẫn: <code>${r.outputPath}</code><br><br>
          <b>Xem trước nội dung:</b>
          <div style="background: white; padding: 12px; border-radius: 4px; margin-top: 8px; max-height: 300px; overflow-y: auto;">
            ${r.html}
          </div>
        </div>`;
    };
    page.querySelector("#d-gen-html").onclick = async () => {
      const fieldsObj = {};
      page.querySelectorAll("#d-fields input[data-field]").forEach(inp => {
        fieldsObj[inp.dataset.field] = inp.value;
      });
      const r = await wcall("go.docs.Service.Fill", {
        templateId, fields: fieldsObj, outputFmt: "html"
      }, "ui");
      if (r) {
        page.querySelector("#d-output").innerHTML = `
          <div class="callout callout-success">
            <b>✓ Đã xuất HTML:</b> <code>${r.outputPath}</code>
            <div style="background: white; padding: 12px; border-radius: 4px; margin-top: 8px; max-height: 300px; overflow-y: auto;">
              ${r.html}
            </div>
          </div>`;
      }
    };
  }
});

// ──────────── History ────────────
route("history", async (page) => {
  page.innerHTML = `<div class="card"><div class="card-title">Lịch sử tính toán</div><div id="hi-list">Đang tải...</div></div>`;
  const items = await wcall("go.calc.Service.History", 100);
  const list = page.querySelector("#hi-list");
  if (!items || items.length === 0) {
    list.innerHTML = "<p class='text-muted'>Chưa có bản ghi nào.</p>";
    return;
  }
  list.innerHTML = `<table><thead><tr><th>Thời gian</th><th>Module</th><th>Inputs</th><th>Outputs</th></tr></thead><tbody>` +
    items.map(h => `<tr><td>${h.ts}</td><td>${h.module}</td><td><code>${h.inputs.substring(0,80)}...</code></td><td><code>${h.outputs.substring(0,80)}...</code></td></tr>`).join("") +
    `</tbody></table>`;
});

// ──────────── Audit ────────────
route("audit", async (page) => {
  page.innerHTML = `<div class="card"><div class="card-title">Audit log</div><p class="text-muted">Audit log ghi lại mọi thao tác tính toán, tra cứu, OCR, che PII, xuất file — dùng cho trách nhiệm giải trình.</p><div class="callout callout-warning">Cần implement query API trong audit module để hiển thị log. Hiện tại chỉ ghi vào store.</div></div>`;
});

// ──────────── About ────────────
route("about", async (page) => {
  const info = await wcall("go.main.App.GetAppInfo");
  const cp = info?.copyright || {};
  page.innerHTML = `
    <div class="card">
      <div class="card-title">Giới thiệu</div>
      <p><b>${info?.name || "Công cụ ngành Kiểm sát"}</b></p>
      <p class="text-muted">Phiên bản ${info?.version || "?"} · Build ${info?.buildDate || "?"} · Platform ${info?.platform || "?"}</p>
      <p class="mt-4">Phần mềm hỗ trợ công tác Kiểm sát chạy 100% ngoại tuyến (offline) dưới dạng single binary Windows/Linux,
        không cần cài đặt. Xây dựng bằng <b>Go + Wails v2</b>.</p>
      <p class="mt-4 text-muted">Data dir: <code>${info?.dataDir || "?"}</code></p>
      <p class="text-muted">Modules: ${(info?.modules || []).join(", ")}</p>
    </div>

    <div class="card">
      <div class="card-title">Bản quyền</div>
      <div class="flex" style="align-items: flex-start; gap: 16px;">
        <img src="${cp.avatarPath || '/avatarC.jpg'}" alt="Coder" style="width:120px; border-radius:6px; flex-shrink:0;">
        <div>
          <p><b>${cp.fullName || "Coder Developed by Nguyễn Hữu Cường"}</b></p>
          <p>Chức vụ: <b>${cp.title || "CV CY KTCC"}</b></p>
          <p>Điện thoại: <a href="tel:${(cp.phone||'').replace(/\s/g,'')}">${cp.phone || "0905 6666 24"}</a></p>
          <p>Email: <a href="mailto:${cp.email}">${cp.email || "doremon.dn01@gmail.com"}</a></p>
          <p class="text-muted mt-4">© 2026 Nguyễn Hữu Cường. Mọi quyền được bảo lưu.</p>
        </div>
      </div>
      <button class="btn mt-4" onclick="document.getElementById('btn-copyright').click()">
        Xem ảnh đầy đủ
      </button>
    </div>

    <div class="card">
      <div class="card-title">Module đã implement</div>
      <ul>
        <li><b>Calc</b>: án phí HS/DS, tiền lãi, tuổi/thời hạn, thời hiệu TNHS, bồi thường NTNĐ</li>
        <li><b>Legal</b>: tra cứu văn bản pháp luật + biểu mẫu (offline store)</li>
        <li><b>OCR</b>: PDF/ảnh → text (Tesseract)</li>
        <li><b>PII</b>: che 10 loại thông tin cá nhân</li>
        <li><b>Spell</b>: rà soát chính tả Vietnamese + telex</li>
        <li><b>File</b>: đổi tên hàng loạt (number/replace/strip/prefix/suffix)</li>
        <li><b>Evidence</b>: gắn nhãn chứng cứ + cây nhãn phiên tòa</li>
        <li><b>Case</b>: quản lý hồ sơ vụ án + deadline</li>
        <li><b>Docs</b>: soạn thảo biểu mẫu với placeholder</li>
      </ul>
    </div>
    <div class="callout">
      <b>Văn bản áp dụng:</b> BLHS 2015 (sửa 2017), BLTTHS 2015 (sửa 128/2025), BLDS 2015,
      BLTTDS 2015 (sửa 85/2025), NR 326/2016/UBTVQH14, Luật 106/2025 (THA DS), Luật 127/2025 (THA HS),
      NĐ 151-152/2026, QĐ 39/QĐ-VKSTC 2018, QĐ 410/QĐ-VKSTC 2021, QĐ 50/QĐ-VKSTC 2026, NR 205/QH15.
    </div>
  `;
});
