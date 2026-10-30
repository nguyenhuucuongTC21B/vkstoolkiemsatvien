package db

// legalDocsSeed — bản snapshot văn bản pháp luật được nhúng vào binary.
// Trong production sẽ thay bằng JSON file đầy đủ (hàng nghìn văn bản)
// nhúng qua go:embed. Ở MVP giữ 10 văn bản tiêu biểu để demo.
var legalDocsSeed = []map[string]interface{}{
        {
                "id":              1,
                "loai":            "Bộ luật",
                "so_hieu":         "100/2015/QH13",
                "co_quan_ban_hanh": "Quốc hội",
                "ngay_ban_hanh":   "2015-11-27",
                "trich_yeu":       "Bộ luật Hình sự",
                "tom_tat":         "BLHS 2015 (sửa 2017) — điều 27 thời hiệu TNHS; điều 12 tuổi chịu TNHS; điều 12 quy định tội phạm",
                "noi_dung":        "Toàn văn BLHS 2015 (sửa đổi, bổ sung 2017). Phần chung quy định về tội phạm, hình phạt, thời hiệu...",
                "linh_vuc":        "hinh_su",
                "con_hieu_luc":    1,
        },
        {
                "id":              2,
                "loai":            "Bộ luật",
                "so_hieu":         "92/2015/QH13",
                "co_quan_ban_hanh": "Quốc hội",
                "ngay_ban_hanh":   "2015-11-25",
                "trich_yeu":       "Bộ luật Tố tụng hình sự",
                "tom_tat":         "BLTTHS 2015 (sửa 128/2025) — tạm giữ 12-72h (điều 128), tạm giam 4 tháng (điều 119), thời hạn tố tụng",
                "noi_dung":        "Toàn văn BLTTHS 2015 với sửa đổi từ Luật 128/2025/QH15...",
                "linh_vuc":        "hinh_su",
                "con_hieu_luc":    1,
        },
        {
                "id":              3,
                "loai":            "Bộ luật",
                "so_hieu":         "91/2015/QH13",
                "co_quan_ban_hanh": "Quốc hội",
                "ngay_ban_hanh":   "2015-11-24",
                "trich_yeu":       "Bộ luật Dân sự",
                "tom_tat":         "BLDS 2015 — điều 306 lãi chậm trả 150%; điều 35 quy đổi tháng=30 ngày; điều 468 lãi suất; điều 589-609 bồi thường NTNĐ",
                "noi_dung":        "Toàn văn BLDS 2015...",
                "linh_vuc":        "dan_su",
                "con_hieu_luc":    1,
        },
        {
                "id":              4,
                "loai":            "Bộ luật",
                "so_hieu":         "92/2015/QH13",
                "co_quan_ban_hanh": "Quốc hội",
                "ngay_ban_hanh":   "2015-11-25",
                "trich_yeu":       "Bộ luật Tố tụng dân sự",
                "tom_tat":         "BLTTDS 2015 (sửa bởi Luật 85/2025) — thời hạn tố tụng dân sự, thẩm quyền, trình tự",
                "noi_dung":        "Toàn văn BLTTDS 2015...",
                "linh_vuc":        "dan_su",
                "con_hieu_luc":    1,
        },
        {
                "id":              5,
                "loai":            "Nghị quyết",
                "so_hieu":         "326/2016/UBTVQH14",
                "co_quan_ban_hanh": "Ủy ban Thường vụ Quốc hội",
                "ngay_ban_hanh":   "2016-08-09",
                "trich_yeu":       "Quy định mức thu, miễn, giảm, thu, nộp, quản lý án phí, lệ phí Tòa án",
                "tom_tat":         "Án phí HS: 100.000đ/bị cáo sơ thẩm (≤5tr/tổ chức); phúc thẩm 50%; DS: bậc thang 5/4/3/2% theo giá trị (≤100tr/100-400tr/400tr-1 tỷ/>1 tỷ); vô giá 300.000đ",
                "noi_dung":        "Nghị quyết 326/2016/UBTVQH14 — toàn văn quy định án phí, lệ phí Tòa án.",
                "linh_vuc":        "dan_su",
                "con_hieu_luc":    1,
        },
        {
                "id":              6,
                "loai":            "Luật",
                "so_hieu":         "106/2025/QH15",
                "co_quan_ban_hanh": "Quốc hội",
                "ngay_ban_hanh":   "2025-06-13",
                "trich_yeu":       "Luật Thi hành án dân sự",
                "tom_tat":         "Luật 106/2025/QH15 — kê biên, bán đấu giá, hạ giá sau 2 lần không thành, thu chi tiền THA DS",
                "noi_dung":        "Luật 106/2025 — thi hành án dân sự.",
                "linh_vuc":        "thi_hanh_an",
                "con_hieu_luc":    1,
        },
        {
                "id":              7,
                "loai":            "Luật",
                "so_hieu":         "127/2025/QH15",
                "co_quan_ban_hanh": "Quốc hội",
                "ngay_ban_hanh":   "2025-11-21",
                "trich_yeu":       "Luật Thi hành án hình sự",
                "tom_tat":         "Luật 127/2025/QH15 — tạm giữ, tạm giam, phạm nhân, chế độ ăn mặc, giảm án, tha tù",
                "noi_dung":        "Luật 127/2025 — thi hành án hình sự.",
                "linh_vuc":        "thi_hanh_an",
                "con_hieu_luc":    1,
        },
        {
                "id":              8,
                "loai":            "Nghị định",
                "so_hieu":         "151/2026/NĐ-CP",
                "co_quan_ban_hanh": "Chính phủ",
                "ngay_ban_hanh":   "2026-03-15",
                "trich_yeu":       "Quy định chi tiết thi hành Luật THA dân sự",
                "tom_tat":         "NĐ 151-152/2026 về THADS — kê biên, bán đấu giá, thu chi tiền",
                "noi_dung":        "Nghị định 151/2026/NĐ-CP quy định chi tiết một số điều của Luật THA DS.",
                "linh_vuc":        "thi_hanh_an",
                "con_hieu_luc":    1,
        },
        {
                "id":              9,
                "loai":            "Quyết định",
                "so_hieu":         "39/QĐ-VKSTC",
                "co_quan_ban_hanh": "Viện trưởng VKSND tối cao",
                "ngay_ban_hanh":   "2018-01-26",
                "trich_yeu":       "Quy chế nghiệp vụ VKS thi hành án hình sự",
                "tom_tat":         "Biểu mẫu nghiệp vụ THAHS theo QĐ 39/QĐ-VKSTC",
                "noi_dung":        "Quyết định 39/QĐ-VKSTC — quy chế nghiệp vụ VKS thi hành án hình sự.",
                "linh_vuc":        "thi_hanh_an",
                "con_hieu_luc":    1,
        },
        {
                "id":              10,
                "loai":            "Quyết định",
                "so_hieu":         "410/QĐ-VKSTC",
                "co_quan_ban_hanh": "Viện trưởng VKSND tối cao",
                "ngay_ban_hanh":   "2021-11-25",
                "trich_yeu":       "Quy định biểu mẫu tố tụng dân sự, hành chính",
                "tom_tat":         "65 mẫu DS, 65 mẫu HC, 21 mẫu PL09 theo QĐ 410/QĐ-VKSTC",
                "noi_dung":        "Quyết định 410/QĐ-VKSTC — biểu mẫu tố tụng dân sự, hành chính.",
                "linh_vuc":        "dan_su",
                "con_hieu_luc":    1,
        },
        {
                "id":              11,
                "loai":            "Quyết định",
                "so_hieu":         "50/QĐ-VKSTC",
                "co_quan_ban_hanh": "Viện trưởng VKSND tối cao",
                "ngay_ban_hanh":   "2026-02-09",
                "trich_yeu":       "Quy chế nghiệp vụ giải quyết khiếu nại, tố cáo",
                "tom_tat":         "QĐ 50/QĐ-VKSTC 2026 — quy trình thụ lý, xác minh, giải quyết đơn KN-TC",
                "noi_dung":        "Quyết định 50/QĐ-VKSTC — quy chế nghiệp vụ giải quyết khiếu nại tố cáo.",
                "linh_vuc":        "khieu_nai",
                "con_hieu_luc":    1,
        },
        {
                "id":              12,
                "loai":            "Án lệ",
                "so_hieu":         "01/2024/AL",
                "co_quan_ban_hanh": "TANDTC",
                "ngay_ban_hanh":   "2024-04-15",
                "trich_yeu":       "Án lệ số 01 về bồi thường thiệt hại do nguồn nguy hiểm cao độ",
                "tom_tat":         "V/v bồi thường thiệt hại ngoài hợp đồng do nguồn nguy hiểm cao độ gây ra",
                "noi_dung":        "Án lệ 01/2024/AL — bồi thường thiệt hại do nguồn nguy hiểm cao độ.",
                "linh_vuc":        "dan_su",
                "con_hieu_luc":    1,
        },
        {
                "id":              13,
                "loai":            "Nghị quyết",
                "so_hieu":         "205/2025/QH15",
                "co_quan_ban_hanh": "Quốc hội",
                "ngay_ban_hanh":   "2025-11-15",
                "trich_yeu":       "Khởi kiện công ích bảo vệ quyền lợi công cộng",
                "tom_tat":         "NR 205/QH15 + TTLT 09/2025 — tiếp nhận, thu thập chứng cứ, kiến nghị, khởi kiện vụ án DS bảo vệ công ích",
                "noi_dung":        "Nghị quyết 205/2025/QH15 về khởi kiện công ích.",
                "linh_vuc":        "khac",
                "con_hieu_luc":    1,
        },
}

// bieuMauSeed — biểu mẫu tố tụng mẫu, nội dung phong phú theo form chuẩn.
var bieuMauSeed = []map[string]interface{}{
        {
                "id":        1,
                "ma_mau":    "PL01-HS01",
                "ten_mau":   "Quyết định khởi tố vụ án hình sự",
                "linh_vuc":  "hinh_su",
                "giai_doan": "dieu_tra",
                "file_html": `<h1>CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM</h1><h2>Độc lập - Tự do - Hạnh phúc</h2>
<h3>QUYẾT ĐỊNH KHỞI TỐ VỤ ÁN HÌNH SỰ</h3>
<p>Số: {{case_number}}/QĐ-KT</p>
<p>Hôm nay, ngày {{today}}, tại {{don_vi}}</p>
<p><b>VỀ VIỆC:</b> Khởi tố vụ án hình sự theo tố cáo của {{nguoi_to_cao}} về hành vi {{hanh_vi}}</p>
<p><b>XÉT:</b></p>
<li>Qua kiểm tra nguồn tin về tội phạm, có căn cứ xác định đã có tội phạm xảy ra.</li>
<li>Hành vi nêu trên có dấu hiệu tội phạm quy định tại Điều {{so_dieu}} Bộ luật Hình sự.</li>
<li>Căn cứ khoản {{so_khoan}} Điều {{so_dieu_dinh}} BLTTHS 2015 (sửa 128/2025).</li>
<p><b>QUYẾT ĐỊNH:</b></p>
<li>Khởi tố vụ án hình sự về tội {{ten_toi}} quy định tại Điều {{so_dieu}} BLHS.</li>
<li>Giao {{co_quan_tham_tra}} tiến hành điều tra vụ án.</li>
<li>Thời hạn điều tra: {{thoi_han}} tháng kể từ ngày khởi tố.</li>
<p><b>Nơi nhận:</b></p>
<li>Như trên (để thực hiện);</li>
<li>Lưu hồ sơ vụ án.</li>`,
                "ghi_chu": "Mẫu theo Thông tư 03/2026/TT-VKSTC — điền case_number, don_vi, nguoi_to_cao, hanh_vi, so_dieu, so_khoan, ten_toi, co_quan_tham_tra, thoi_han",
        },
        {
                "id":        2,
                "ma_mau":    "PL01-DS01",
                "ten_mau":   "Thông báo thụ lý dân sự",
                "linh_vuc":  "dan_su",
                "giai_doan": "thu_ly",
                "file_html": `<h1>CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM</h1><h2>Độc lập - Tự do - Hạnh phúc</h2>
<h3>THÔNG BÁO THỤ LÝ VỤ ÁN DÂN SỰ</h3>
<p>Số: {{case_number}}/TB-TL</p>
<p>Hôm nay, ngày {{today}}, tại Tòa án nhân dân {{toa_an}}</p>
<p><b>VỀ VIỆC:</b> Thụ lý vụ án dân sự về việc {{noi_dung_vu_an}}</p>
<p><b>NGUYÊN ĐƠN:</b> {{nguyen_don}} — địa chỉ: {{dc_nguyen}}</p>
<p><b>BỊ ĐƠN:</b> {{bi_don}} — địa chỉ: {{dc_bi}}</p>
<p><b>XÉT:</b> Đơn khởi kiện của nguyên đơn đầy đủ thủ tục theo Điều 189 BLTTDS 2015 (sửa 85/2025).</p>
<p><b>THÔNG BÁO:</b></p>
<li>Tòa án nhân dân {{toa_an}} đã thụ lý vụ án dân sự số {{case_number}}.</li>
<li>Đương sự có 15 ngày kể từ ngày nhận thông báo để chuẩn bị tài liệu chứng cứ.</li>
<li>Thời hạn chuẩn bị xét xử: {{thoi_han}} tháng.</li>
<p><b>Nơi nhận:</b></p>
<li>Nguyên đơn, bị đơn (để biết);</li>
<li>VKSND cùng cấp (để kiểm sát);</li>
<li>Lưu hồ sơ vụ án.</li>`,
                "ghi_chu": "Mẫu theo QĐ 410/QĐ-VKSTC — điền case_number, toa_an, noi_dung_vu_an, nguyen_don, dc_nguyen, bi_don, dc_bi, thoi_han",
        },
        {
                "id":        3,
                "ma_mau":    "PL02-DS01",
                "ten_mau":   "Quyết định phân công Kiểm sát viên",
                "linh_vuc":  "dan_su",
                "giai_doan": "thu_ly",
                "file_html": `<h1>CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM</h1><h2>Độc lập - Tự do - Hạnh phúc</h2>
<h3>QUYẾT ĐỊNH PHÂN CÔNG KIỂM SÁT VIÊN</h3>
<p>Số: {{case_number}}/QĐ-PC</p>
<p>Hôm nay, ngày {{today}}</p>
<p><b>VỀ VIỆC:</b> Phân công KSV thực hành quyền công tố, kiểm sát việc thụ lý vụ án dân sự</p>
<p><b>VỤ ÁN:</b> {{case_number}} — nguyên đơn: {{nguyen_don}}, bị đơn: {{bi_don}}</p>
<p><b>XÉT:</b></p>
<li>Vụ án trên đã được Tòa án nhân dân {{toa_an}} thụ lý ngày {{ngay_thu_ly}}.</li>
<li>Căn cứ Điều 25 Luật Tổ chức VKSND; khoản 2 Điều 21 BLTTDS 2015.</li>
<p><b>QUYẾT ĐỊNH:</b></p>
<li>Phân công {{ksv}} — KSV ngạch {{ngach}} thực hành quyền công tố, kiểm sát việc giải quyết vụ án dân sự số {{case_number}}.</li>
<li>KSV được phân công có trách nhiệm kiểm sát việc thụ lý, chuẩn bị xét xử, xét xử vụ án theo quy định pháp luật.</li>
<p><b>Nơi nhận:</b></p>
<li>KSV được phân công (để thực hiện);</li>
<li>Lưu hồ sơ.</li>`,
                "ghi_chu": "Mẫu phân công KSV theo QĐ 410 — điền case_number, toa_an, ngay_thu_ly, ksv, ngach, nguyen_don, bi_don",
        },
        {
                "id":        4,
                "ma_mau":    "PL07-KN01",
                "ten_mau":   "Quyết định giải quyết khiếu nại",
                "linh_vuc":  "khieu_nai",
                "giai_doan": "giai_quyet",
                "file_html": `<h1>CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM</h1><h2>Độc lập - Tự do - Hạnh phúc</h2>
<h3>QUYẾT ĐỊNH GIẢI QUYẾT KHIẾU NẠI</h3>
<p>Số: {{case_number}}/QĐ-KN</p>
<p>Hôm nay, ngày {{today}}</p>
<p><b>NGƯỜI KHIẾU NẠI:</b> {{nguoi_kn}}, sinh năm {{nam_sinh}}, CCCD {{cccd}}</p>
<p><b>NỘI DUNG KHIẾU NẠI:</b> {{noi_dung_kn}}</p>
<p><b>QUYẾT ĐỊNH:</b></p>
<li>{{ket_qua}}</li>
<li>Áp dụng {{can_cu_phap_ly}}</li>
<p><b>Nơi nhận:</b></p>
<li>Người khiếu nại;</li>
<li>Cơ quan, tổ chức, cá nhân có liên quan;</li>
<li>Lưu hồ sơ.</li>`,
                "ghi_chu": "Mẫu theo QĐ 50/QĐ-VKSTC 2026 — điền case_number, nguoi_kn, nam_sinh, cccd, noi_dung_kn, ket_qua, can_cu_phap_ly",
        },
        {
                "id":        5,
                "ma_mau":    "PL05-THA01",
                "ten_mau":   "Quyết định thi hành án dân sự",
                "linh_vuc":  "thi_hanh_an",
                "giai_doan": "thi_hanh",
                "file_html": `<h1>CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM</h1><h2>Độc lập - Tự do - Hạnh phúc</h2>
<h3>QUYẾT ĐỊNH THI HÀNH ÁN DÂN SỰ</h3>
<p>Số: {{case_number}}/QĐ-THA</p>
<p>Hôm nay, ngày {{today}}</p>
<p><b>VỀ VIỆC:</b> Thi hành bản án dân sự số {{so_ban_an}} của Tòa án nhân dân {{toa_an}}</p>
<p><b>BÊN KIỆN:</b> {{ben_kien}}</p>
<p><b>BÊN BỊ KIỆN:</b> {{ben_bi_kien}}</p>
<p><b>NỘI DUNG BẢN ÁN:</b> {{noi_dung_ban_an}}</p>
<p><b>QUYẾT ĐỊNH:</b></p>
<li>Thi hành bản án số {{so_ban_an}} theo quy định Luật 106/2025/QH15.</li>
<li>Thời hạn thi hành tự nguyện: 15 ngày kể từ ngày nhận quyết định.</li>
<li>Án phí: {{an_phi}} đồng.</li>
<p><b>Nơi nhận:</b></p>
<li>Bên kiện, bên bị kiện;</li>
<li>Cơ quan thi hành án;</li>
<li>Lưu hồ sơ.</li>`,
                "ghi_chu": "Mẫu theo Luật 106/2025 — điền case_number, so_ban_an, toa_an, ben_kien, ben_bi_kien, noi_dung_ban_an, an_phi",
        },
        {
                "id":        6,
                "ma_mau":    "PL03-KKCI01",
                "ten_mau":   "Đơn khởi kiện công ích",
                "linh_vuc":  "khac",
                "giai_doan": "khoi_kien",
                "file_html": `<h1>CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM</h1><h2>Độc lập - Tự do - Hạnh phúc</h2>
<h3>ĐƠN KHỞI KIỆN CÔNG ÍCH</h3>
<p>Kính gửi: Tòa án nhân dân {{toa_an}}</p>
<p><b>NGƯỜI KHỞI KIỆN:</b> Viện kiểm sát nhân dân {{vien_ks}}</p>
<p><b>VỀ VIỆC:</b> Khởi kiện vụ án dân sự bảo vệ quyền lợi công cộng theo Nghị quyết 205/QH15</p>
<p><b>I. NỘI DUNG SỰ VIỆC:</b></p>
<li>{{noi_dung_su_viec}}</li>
<p><b>II. CĂN CỨ PHÁP LUẬT:</b></p>
<li>Nghị quyết số 205/2025/QH15 của Quốc hội;</li>
<li>Thông tư liên tịch số 09/2025/TTLT-TANDTC-VKSNDTC-BTP;</li>
<li>Bộ luật Dân sự 2015, Bộ luật Tố tụng dân sự 2015 (sửa 85/2025).</li>
<p><b>III. YÊU CẦU:</b></p>
<li>Xử lý hành vi vi phạm pháp luật gây thiệt hại đến quyền lợi công cộng.</li>
<li>Buộc {{doi_tuong}} thực hiện {{yeu_cau}}.</li>
<p><b>BIỆN LAI:</b> {{bien_lai}}</p>`,
                "ghi_chu": "Mẫu khởi kiện công ích theo NQ 205 + TTLT 09/2025 — điền toa_an, vien_ks, noi_dung_su_viec, doi_tuong, yeu_cau, bien_lai",
        },
        {
                "id":        7,
                "ma_mau":    "PL01-HS02",
                "ten_mau":   "Quyết định phê chuẩn lệnh tạm giữ",
                "linh_vuc":  "hinh_su",
                "giai_doan": "dieu_tra",
                "file_html": `<h1>CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM</h1><h2>Độc lập - Tự do - Hạnh phúc</h2>
<h3>QUYẾT ĐỊNH PHÊ CHUẨN LỆNH TẠM GIỮ</h3>
<p>Số: {{case_number}}/QĐ-PC-TG</p>
<p>Hôm nay, ngày {{today}}</p>
<p><b>VỀ VIỆC:</b> Phê chuẩn lệnh tạm giữ {{doi_tuong}}</p>
<p><b>XÉT:</b> Lệnh tạm giữ số {{so_lenh}} do {{co_quan_ban_lenh}} ban hành ngày {{ngay_ban_lenh}}</p>
<p>Căn cứ điểm {{so_diem}} khoản {{so_khoan}} Điều 128 BLTTHS 2015 (sửa 128/2025):</p>
<p><b>QUYẾT ĐỊNH:</b></p>
<li>Phê chuẩn lệnh tạm giữ {{doi_tuong}} trong thời hạn {{so_gio}} giờ kể từ {{gio_bat_dau}}.</li>
<li>Thời điểm hết hạn tạm giữ: {{gio_ket_thuc}}.</li>
<p><b>Nơi nhận:</b></p>
<li>Cơ quan ban hành lệnh;</li>
<li>Đối tượng bị tạm giữ;</li>
<li>Lưu hồ sơ.</li>`,
                "ghi_chu": "Mẫu phê chuẩn tạm giữ theo BLTTHS 128/2025 — điền case_number, doi_tuong, so_lenh, co_quan_ban_lenh, ngay_ban_lenh, so_diem, so_khoan, so_gio, gio_bat_dau, gio_ket_thuc",
        },
        {
                "id":        8,
                "ma_mau":    "PL01-HS03",
                "ten_mau":   "Quyết định phê chuẩn tạm giam",
                "linh_vuc":  "hinh_su",
                "giai_doan": "dieu_tra",
                "file_html": `<h1>CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM</h1><h2>Độc lập - Tự do - Hạnh phúc</h2>
<h3>QUYẾT ĐỊNH PHÊ CHUẨN TẠM GIAM</h3>
<p>Số: {{case_number}}/QĐ-PC-TGiam</p>
<p>Hôm nay, ngày {{today}}</p>
<p><b>VỀ VIỆC:</b> Phê chuẩn quyết định tạm giam {{doi_tuong}}</p>
<p><b>TỘI DANH:</b> {{toi_danh}} — Điều {{so_dieu}} BLHS</p>
<p><b>XÉT:</b></p>
<li>Quyết định tạm giam số {{so_qd}} của {{co_quan_qd}} ngày {{ngay_qd}}.</li>
<li>Căn cứ khoản {{so_khoan}} Điều 119 BLTTHS 2015.</li>
<p><b>QUYẾT ĐỊNH:</b></p>
<li>Phê chuẩn quyết định tạm giam {{doi_tuong}} trong thời hạn {{so_thang}} tháng.</li>
<li>Thời hạn tạm giam đến hết ngày {{ngay_ket_thuc}}.</li>
<p><b>Nơi nhận:</b></p>
<li>Cơ quan ra quyết định;</li>
<li>Trại tạm giam;</li>
<li>Bị can, người đại diện hợp pháp;</li>
<li>Lưu hồ sơ.</li>`,
                "ghi_chu": "Mẫu phê chuẩn tạm giam theo BLTTHS điều 119 — điền case_number, doi_tuong, toi_danh, so_dieu, so_qd, co_quan_qd, ngay_qd, so_khoan, so_thang, ngay_ket_thuc",
        },
        {
                "id":        9,
                "ma_mau":    "PL06-VP01",
                "ten_mau":   "Báo cáo thống kê hàng tháng",
                "linh_vuc":  "khac",
                "giai_doan": "van_phong",
                "file_html": `<h1>CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM</h1><h2>Độc lập - Tự do - Hạnh phúc</h2>
<h3>BÁO CÁO THỐNG KÊ HÀNG THÁNG</h3>
<p>Đơn vị: {{don_vi}}</p>
<p>Tháng: {{thang}}/{{nam}}</p>
<p><b>I. SỐ LIỆU THỐNG KÊ:</b></p>
<li>Vụ án hình sự thụ lý mới: {{hs_thu_ly}}</li>
<li>Vụ án hình sự đã giải quyết: {{hs_giai_quyet}}</li>
<li>Vụ án dân sự thụ lý mới: {{ds_thu_ly}}</li>
<li>Vụ án dân sự đã giải quyết: {{ds_giai_quyet}}</li>
<li>Đơn khiếu nại, tố cáo thụ lý: {{kn_tc}}</li>
<p><b>II. ĐÁNH GIÁ:</b></p>
<li>{{danh_gia}}</li>
<p><b>III. KHUYẾN NGHỊ:</b></p>
<li>{{khuyen_nghi}}</li>
<p><b>Nơi nhận:</b></p>
<li>VKSND cấp trên;</li>
<li>Lưu hồ sơ.</li>`,
                "ghi_chu": "Mẫu báo cáo tháng theo QĐ 358/QĐ-VKSTC — điền don_vi, thang, nam, hs_thu_ly, hs_giai_quyet, ds_thu_ly, ds_giai_quyet, kn_tc, danh_gia, khuyen_nghi",
        },
}

// spellRulesSeed — quy tắc chính tả/Nghiệp vụ mặc định.
type spellRuleSeed struct {
        Pattern, Replacement, RuleType, Notes string
}

var spellRulesSeed = []spellRuleSeed{
        // term_norm: normalization of legal terms
        {"thi han an", "thi hành án", "term_norm", "normalize"},
        {"thi han", "thi hành", "term_norm", "normalize"},
        {"kiem sat", "kiểm sát", "term_norm", "normalize"},
        {"vien kiem sat", "viện kiểm sát", "term_norm", "normalize"},
        {"toa an", "tòa án", "term_norm", "normalize"},
        {"bo luat", "bộ luật", "term_norm", "normalize"},
        {"tam giu", "tạm giữ", "term_norm", "normalize"},
        {"tam giam", "tạm giam", "term_norm", "normalize"},
        {"an phi", "án phí", "term_norm", "normalize"},
        {"tien lai", "tiền lãi", "term_norm", "normalize"},
        {"boi thuong", "bồi thường", "term_norm", "normalize"},
        {"thoi hieu", "thời hiệu", "term_norm", "normalize"},
        // literal: common typos
        {"nhung", "nhưng", "literal", "typo"},
        {"truong hop", "trường hợp", "literal", "typo"},
        {"cua", "của", "literal", "typo"},
        {"duoc", "được", "literal", "typo"},
}
