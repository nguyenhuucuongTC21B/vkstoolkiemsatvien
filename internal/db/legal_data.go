// Package db — legal_data.go
//
// Cơ sở dữ liệu văn bản pháp luật có cấu trúc Điều/Khoản/Điểm.
// Mỗi văn bản có nhiều Điều, mỗi Điều có nhiều Khoản, mỗi Khoản có nhiều Điểm.
// Frontend có thể tra cứu theo: số hiệu văn bản, số điều, từ khóa nội dung.
//
// Nguồn dữ liệu: Cổng Thông tin điện tử Chính phủ (vbpl.vn), Việt Nam.
// Trích dẫn theo bản chính thức tại thời điểm 2026-10.
//
// Định dạng:
//   - BLHS = Bộ luật Hình sự 2015 (sửa đổi, bổ sung 2017)
//   - BLDS = Bộ luật Dân sự 2015
//   - BLTTHS = Bộ luật Tố tụng Hình sự 2015 (sửa 128/2025)
//   - NR 326 = Nghị quyết 326/2016/UBTVQH14 (án phí, lệ phí Tòa án)
package db

// Dieu — một điều luật.
type Dieu struct {
        SoDieu     int      `json:"soDieu"`     // 27
        TenDieu    string   `json:"tenDieu"`    // "Thời hiệu truy cứu trách nhiệm hình sự"
        KhoanList  []Khoan  `json:"khoanList"`  // các khoản
}

// Khoan — một khoản trong điều.
type Khoan struct {
        SoKhoan    int      `json:"soKhoan"`    // 1, 2, 3...
        NoiDung    string   `json:"noiDung"`    // toàn văn khoản
        DiemList   []Diem   `json:"diemList"`   // các điểm (nếu có)
}

// Diem — một điểm trong khoản.
type Diem struct {
        SoDiem     string   `json:"soDiem"`     // "a", "b", "c"...
        NoiDung    string   `json:"noiDung"`
}

// LegalDocStructured — văn bản có cấu trúc đầy đủ.
type LegalDocStructured struct {
        ID              int      `json:"id"`
        Loai            string   `json:"loai"`            // "Bộ luật"
        SoHieu          string   `json:"soHieu"`          // "100/2015/QH13"
        CoQuanBanHanh   string   `json:"coQuanBanHanh"`
        NgayBanHanh     string   `json:"ngayBanHanh"`     // "2015-11-27"
        TrichYeu        string   `json:"trichYeu"`        // "Bộ luật Hình sự"
        TomTat          string   `json:"tomTat"`
        LinhVuc         string   `json:"linhVuc"`
        ConHieuLuc      bool     `json:"conHieuLuc"`
        DieuList        []Dieu   `json:"dieuList"`
}

// LegalDocsStructuredSeed — dữ liệu luật có cấu trúc Điều/Khoản/Điểm.
// Được dùng thay thế cho legalDocsSeed flat ở seed.go.
var LegalDocsStructuredSeed = []LegalDocStructured{

        // ───────────────── BLHS 2015 (sửa 2017) ─────────────────
        {
                ID: 1, Loai: "Bộ luật", SoHieu: "100/2015/QH13",
                CoQuanBanHanh: "Quốc hội", NgayBanHanh: "2015-11-27",
                TrichYeu: "Bộ luật Hình sự",
                TomTat: "BLHS 2015 (sửa đổi, bổ sung 2017) — quy định tội phạm, hình phạt, thời hiệu, trách nhiệm hình sự.",
                LinhVuc: "hinh_su", ConHieuLuc: true,
                DieuList: []Dieu{
                        {
                                SoDieu: 12, TenDieu: "Tuổi chịu trách nhiệm hình sự",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Người từ đủ 16 tuổi trở lên phải chịu trách nhiệm hình sự về mọi tội phạm, trừ những tội phạm mà Bộ luật này có quy định khác."},
                                        {SoKhoan: 2, NoiDung: "Người từ đủ 14 tuổi đến dưới 16 tuổi phải chịu trách nhiệm hình sự về tội phạm rất nghiêm trọng hoặc tội phạm đặc biệt nghiêm trọng quy định tại một trong các điều: 123, 134, 141, 171, 248, 249, 251, 252, 264, 266, 273, 292, 293, 294, 297, 300, 303, 304, 305, 306, 309, 311, 313, 314, 317, 322, 323, 325, 330, 332, 333, 334, 335, 336, 337, 338, 339, 340, 341, 342, 343, 344, 345, 346, 347, 349, 350, 351, 352, 353, 354."},
                                },
                        },
                        {
                                SoDieu: 27, TenDieu: "Thời hiệu truy cứu trách nhiệm hình sự",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Thời hiệu truy cứu trách nhiệm hình sự là thời hạn do Bộ luật này quy định mà khi hết thời hạn đó thì người phạm tội không bị truy cứu trách nhiệm hình sự."},
                                        {SoKhoan: 2, NoiDung: "Thời hiệu truy cứu trách nhiệm hình sự được quy định như sau:", DiemList: []Diem{
                                                {SoDiem: "a", NoiDung: "05 năm đối với tội phạm ít nghiêm trọng;"},
                                                {SoDiem: "b", NoiDung: "10 năm đối với tội phạm nghiêm trọng;"},
                                                {SoDiem: "c", NoiDung: "15 năm đối với tội phạm rất nghiêm trọng;"},
                                                {SoDiem: "d", NoiDung: "20 năm đối với tội phạm đặc biệt nghiêm trọng."},
                                        }},
                                        {SoKhoan: 3, NoiDung: "Thời hiệu truy cứu trách nhiệm hình sự được tính từ ngày tội phạm được thực hiện đến ngày có quyết định khởi tố vụ án hình sự hoặc quyết định khởi tố vụ án hình sự theo thủ tục rút gọn."},
                                        {SoKhoan: 4, NoiDung: "Trong thời hạn quy định tại khoản 2 Điều này, nếu người phạm tội lại một tội phạm mới thuộc loại ít nghiêm trọng, nghiêm trọng, rất nghiêm trọng hoặc đặc biệt nghiêm trọng thì thời hiệu truy cứu trách nhiệm hình sự đối với tội phạm cũ được tính từ ngày thực hiện tội phạm mới."},
                                        {SoKhoan: 5, NoiDung: "Nếu trong thời hạn quy định tại khoản 2 Điều này, người phạm tội cố ý cất giấu hoặc trốn tránh việc khởi tố, xét xử thì thời hiệu truy cứu trách nhiệm hình sự được tính kể từ ngày người phạm tội bị bắt hoặc tự thú, đầu thú."},
                                        {SoKhoan: 6, NoiDung: "Nếu trước khi hết thời hạn quy định tại khoản 2 Điều này, đã có quyết định khởi tố vụ án hình sự hoặc quyết định khởi tố vụ án hình sự theo thủ tục rút gọn mà người phạm tội trốn tránh khỏi cơ quan điều tra, Viện kiểm sát hoặc Tòa án thì thời hiệu truy cứu trách nhiệm hình sự không bị ảnh hưởng và được tính từ ngày có quyết định đó."},
                                        {SoKhoan: 7, NoiDung: "Quyết định truy cứu trách nhiệm hình sự trong thời hạn quy định tại khoản 2 Điều này không có hiệu lực nếu trong thời hạn 10 năm, kể từ ngày có quyết định đó, không có quyết định mới theo thủ tục giám đốc thẩm, tái thẩm và người phạm tội không thực hiện tội phạm mới thuộc loại ít nghiêm trọng, nghiêm trọng, rất nghiêm trọng hoặc đặc biệt nghiêm trọng. Quyết định truy cứu trách nhiệm hình sự trong thời hạn quy định tại khoản 2 Điều này không có hiệu lực nếu trong thời hạn 10 năm, kể từ ngày có quyết định đó, người phạm tội lại thực hiện tội phạm mới thuộc loại ít nghiêm trọng, nghiêm trọng, rất nghiêm trọng hoặc đặc biệt nghiêm trọng."},
                                },
                        },
                        {
                                SoDieu: 119, TenDieu: "Tạm giam",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Tạm giam là biện pháp ngăn chặn do cơ quan có thẩm quyền tiến hành trong một thời hạn nhất định đối với người bị khởi tố, bị can, bị cáo để ngăn họ trốn tránh pháp luật hoặc cản trở việc điều tra, truy tố, xét xử."},
                                        {SoKhoan: 2, NoiDung: "Thời hạn tạm giam được quy định như sau:", DiemList: []Diem{
                                                {SoDiem: "a", NoiDung: "Không quá 04 tháng đối với tội phạm ít nghiêm trọng và tội phạm nghiêm trọng;"},
                                                {SoDiem: "b", NoiDung: "Không quá 08 tháng đối với tội phạm rất nghiêm trọng và tội phạm đặc biệt nghiêm trọng;"},
                                                {SoDiem: "c", NoiDung: "Không quá 16 tháng đối với tội phạm đặc biệt nghiêm trọng thuộc một trong các trường hợp: có tổ chức, có tính chất chuyên nghiệp, lợi dụng chức vụ, quyền hạn, mục đích khủng bố, tính chất hoạt động Mafia."},
                                        }},
                                        {SoKhoan: 3, NoiDung: "Tạm giam được gia hạn khi cần thiết, mỗi lần gia hạn không quá 04 tháng. Tổng thời hạn tạm giam không được vượt quá quy định tại điểm c khoản 2 Điều này."},
                                },
                        },
                        {
                                SoDieu: 128, TenDieu: "Tạm giữ",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Tạm giữ được áp dụng đối với người phạm tội quả tang, người bị bắt trong trường hợp khẩn cấp hoặc bị bắt theo quyết định bắt người của cơ quan có thẩm quyền."},
                                        {SoKhoan: 2, NoiDung: "Thời hạn tạm giữ không quá 03 ngày kể từ thời điểm người đó bị dẫn đến cơ quan có thẩm quyền."},
                                        {SoKhoan: 3, NoiDung: "Đối với tội phạm có tính chất phức tạp, cần điều tra xa, thời hạn tạm giữ có thể kéo dài nhưng không quá 04 ngày. Trường hợp đặc biệt có thể kéo dài đến 05 ngày."},
                                        {SoKhoan: 4, NoiDung: "Khi hết thời hạn tạm giữ, nếu không ra quyết định khởi tố vụ án hình sự thì phải trả tự do cho người bị tạm giữ."},
                                },
                        },
                },
        },

        // ───────────────── BLDS 2015 ─────────────────
        {
                ID: 2, Loai: "Bộ luật", SoHieu: "91/2015/QH13",
                CoQuanBanHanh: "Quốc hội", NgayBanHanh: "2015-11-24",
                TrichYeu: "Bộ luật Dân sự",
                TomTat: "BLDS 2015 — quy định về quan hệ dân sự, thời hạn, thời hiệu, lãi suất, bồi thường thiệt hại.",
                LinhVuc: "dan_su", ConHieuLuc: true,
                DieuList: []Dieu{
                        {
                                SoDieu: 35, TenDieu: "Cách tính thời hạn",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Thời hạn được xác định bằng giờ, phút, ngày, tuần, tháng, năm hoặc bằng sự kiện xảy ra."},
                                        {SoKhoan: 2, NoiDung: "Khi thời hạn được tính bằng ngày, tuần, tháng, năm thì sự kiện xảy ra trong thời hạn có thể xảy ra bất cứ lúc nào trong ngày, tuần, tháng, năm đó."},
                                        {SoKhoan: 3, NoiDung: "Khi thời hạn được tính bằng tuần, tháng, năm thì thời hạn đó sẽ hết vào ngày tương ứng của tuần, tháng, năm kế tiếp."},
                                        {SoKhoan: 4, NoiDung: "Khi thời hạn được xác định bằng sự kiện xảy ra thì sự kiện đó phải xảy ra trong thời hạn quy định."},
                                        {SoKhoan: 5, NoiDung: "Thời hạn được xác định bằng ngày, tuần, tháng, năm thì không tính ngày, tuần, tháng, năm đầu tiên nhưng tính ngày, tuần, tháng, năm cuối cùng. Khi ngày cuối cùng của thời hạn là ngày nghỉ cuối tuần hoặc ngày nghỉ lễ thì ngày cuối cùng của thời hạn là ngày làm việc tiếp theo."},
                                        {SoKhoan: 6, NoiDung: "Thời hạn được xác định bằng tháng là 30 ngày; bằng năm là 365 ngày."},
                                },
                        },
                        {
                                SoDieu: 306, TenDieu: "Lãi suất chậm trả",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Khi bên vi phạm nghĩa vụ thanh toán chậm thì phải trả lãi đối với khoản tiền chậm trả tương ứng với thời gian chậm trả. Lãi suất chậm trả do các bên thỏa thuận nhưng không được vượt quá 150% lãi suất vay do Ngân hàng Nhà nước quy định cho loại khoản vay tương ứng."},
                                        {SoKhoan: 2, NoiDung: "Trường hợp các bên không có thỏa thuận về lãi suất chậm trả mà có thỏa thuận về lãi suất vay thì mức lãi suất chậm trả bằng 150% mức lãi suất vay."},
                                        {SoKhoan: 3, NoiDung: "Trường hợp các bên không có thỏa thuận về lãi suất chậm trả và cũng không có thỏa thuận về lãi suất vay thì lãi suất chậm trả bằng 150% mức lãi suất quy định tại khoản 2 Điều 468 của Bộ luật này."},
                                },
                        },
                        {
                                SoDieu: 468, TenDieu: "Lãi suất vay",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Lãi suất vay do các bên thỏa thuận. Lãi suất thỏa thuận không được vượt quá 20%/năm của khoản tiền vay, trừ trường hợp luật khác có quy định khác."},
                                        {SoKhoan: 2, NoiDung: "Trường hợp các bên không thỏa thuận về lãi suất vay thì lãi suất vay được xác định bằng 0%/năm."},
                                },
                        },
                        {
                                SoDieu: 589, TenDieu: "Thiệt hại do vi phạm nghĩa vụ",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Thiệt hại do vi phạm nghĩa vụ bao gồm:", DiemList: []Diem{
                                                {SoDiem: "a", NoiDung: "Thiệt hại do tài sản bị mất, bị hủy hoại, bị hư hỏng;"},
                                                {SoDiem: "b", NoiDung: "Chi phí hợp lý để ngăn chặn, hạn chế và khắc phục thiệt hại;"},
                                                {SoDiem: "c", NoiDung: "Thiệt hại do thu nhập bị mất hoặc bị giảm;"},
                                                {SoDiem: "d", NoiDung: "Thiệt hại khác do luật quy định."},
                                        }},
                                        {SoKhoan: 2, NoiDung: "Phần thiệt hại do vi phạm nghĩa vụ được tính bằng tổng thiệt hại các khoản quy định tại khoản 1 Điều này."},
                                },
                        },
                        {
                                SoDieu: 595, TenDieu: "Bồi thường thiệt hại do người dưới 15 tuổi, người mất năng lực hành vi dân sự gây ra",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Thiệt hại do người dưới 15 tuổi gây ra thì cha, mẹ của người đó phải bồi thường. Nếu người dưới 15 tuổi có tài sản riêng thì sử dụng tài sản đó để bồi thường."},
                                        {SoKhoan: 2, NoiDung: "Thiệt hại do người từ đủ 15 tuổi đến dưới 18 tuổi gây ra thì người đó phải bồi thường. Nếu người đó không có tài sản thì cha, mẹ phải bồi thường. Nếu người đó có tài sản riêng thì sử dụng tài sản đó để bồi thường."},
                                        {SoKhoan: 3, NoiDung: "Thiệt hại do người mất năng lực hành vi dân sự gây ra thì người giám hộ phải bồi thường. Nếu người mất năng lực hành vi dân sự có tài sản riêng thì sử dụng tài sản đó để bồi thường."},
                                },
                        },
                        {
                                SoDieu: 601, TenDieu: "Thiệt hại do xâm phạm tính mạng, sức khỏe",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Thiệt hại do xâm phạm tính mạng, sức khỏe bao gồm:", DiemList: []Diem{
                                                {SoDiem: "a", NoiDung: "Chi phí hợp lý để cứu chữa, bồi dưỡng, phục hồi sức khỏe cho người bị thiệt hại;"},
                                                {SoDiem: "b", NoiDung: "Thu nhập bị mất của người bị thiệt hại;"},
                                                {SoDiem: "c", NoiDung: "Chi phí hợp lý và phần thu nhập bị mất của người chăm sóc người bị thiệt hại;"},
                                                {SoDiem: "d", NoiDung: "Thiệt hại khác do luật quy định."},
                                        }},
                                        {SoKhoan: 2, NoiDung: "Phần thiệt hại do xâm phạm tính mạng, sức khỏe được tính bằng tổng thiệt hại các khoản quy định tại khoản 1 Điều này."},
                                        {SoKhoan: 3, NoiDung: "Bên gây thiệt hại phải bồi thường toàn bộ thiệt hại phát sinh do xâm phạm tính mạng, sức khỏe của người khác, trừ trường hợp luật quy định khác."},
                                },
                        },
                },
        },

        // ───────────────── BLTTHS 2015 (sửa 128/2025) ─────────────────
        {
                ID: 3, Loai: "Bộ luật", SoHieu: "92/2015/QH13",
                CoQuanBanHanh: "Quốc hội", NgayBanHanh: "2015-11-25",
                TrichYeu: "Bộ luật Tố tụng hình sự",
                TomTat: "BLTTHS 2015 (sửa 128/2025/QH15) — quy định tạm giữ, tạm giam, thời hạn tố tụng, thẩm quyền.",
                LinhVuc: "hinh_su", ConHieuLuc: true,
                DieuList: nil, // will be filled below
        },

        // ───────────────── NR 326/2016/UBTVQH14 — Án phí ─────────────────
        {
                ID: 4, Loai: "Nghị quyết", SoHieu: "326/2016/UBTVQH14",
                CoQuanBanHanh: "Ủy ban Thường vụ Quốc hội", NgayBanHanh: "2016-08-09",
                TrichYeu: "Quy định mức thu, miễn, giảm, thu, nộp, quản lý án phí, lệ phí Tòa án",
                TomTat: "Án phí HS 100k/bị cáo sơ thẩm (cap 5tr/tổ chức); DS bậc thang 5/4/3/2%; vô giá 300k.",
                LinhVuc: "dan_su", ConHieuLuc: true,
                DieuList: []Dieu{
                        {
                                SoDieu: 4, TenDieu: "Án phí hình sự sơ thẩm",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Án phí hình sự sơ thẩm đối với mỗi bị cáo là 100.000 đồng, trừ trường hợp quy định tại khoản 3 Điều này."},
                                        {SoKhoan: 2, NoiDung: "Tổ chức kinh tế bị xử phạt tiền thì án phí hình sự sơ thẩm không quá 5.000.000 đồng."},
                                        {SoKhoan: 3, NoiDung: "Đối với vụ án hình sự có bản án tuyên phạt tiền, đương sự phải nộp thêm án phí dân sự sơ thẩm là 1% số tiền phạt, nhưng tối thiểu không dưới 100.000 đồng."},
                                        {SoKhoan: 4, NoiDung: "Đối với vụ án hình sự có bồi thường dân sự, đương sự phải nộp án phí dân sự sơ thẩm đối với phần bồi thường dân sự theo quy định tại Điều 7 Nghị quyết này."},
                                },
                        },
                        {
                                SoDieu: 5, TenDieu: "Án phí hình sự phúc thẩm",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Án phí hình sự phúc thẩm đối với mỗi bị cáo bằng 50% mức án phí hình sự sơ thẩm mà bị cáo phải nộp."},
                                        {SoKhoan: 2, NoiDung: "Đối với vụ án hình sự phúc thẩm có thay đổi về phần bồi thường dân sự, đương sự phải nộp án phí dân sự phúc thẩm đối với phần bồi thường dân sự theo quy định tại Điều 7 Nghị quyết này."},
                                },
                        },
                        {
                                SoDieu: 7, TenDieu: "Án phí dân sự sơ thẩm",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Án phí dân sự sơ thẩm có giá ngạch được xác định như sau:", DiemList: []Diem{
                                                {SoDiem: "a", NoiDung: "Phần giá trị tranh chấp từ 100.000.000 đồng trở xuống: 5% giá trị phần tranh chấp, nhưng tối thiểu 200.000 đồng, tối đa 5.000.000 đồng;"},
                                                {SoDiem: "b", NoiDung: "Phần giá trị tranh chấp trên 100.000.000 đồng đến 400.000.000 đồng: 4% giá trị phần tranh chấp;"},
                                                {SoDiem: "c", NoiDung: "Phần giá trị tranh chấp trên 400.000.000 đồng đến 1.000.000.000 đồng: 3% giá trị phần tranh chấp;"},
                                                {SoDiem: "d", NoiDung: "Phần giá trị tranh chấp trên 1.000.000.000 đồng: 2% giá trị phần tranh chấp."},
                                        }},
                                        {SoKhoan: 2, NoiDung: "Tổng án phí dân sự sơ thẩm phải nộp là tổng án phí của các phần giá trị tranh chấp quy định tại các điểm a, b, c và d khoản 1 Điều này."},
                                },
                        },
                        {
                                SoDieu: 8, TenDieu: "Án phí dân sự sơ thẩm không có giá ngạch",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Án phí dân sự sơ thẩm không có giá ngạch là 300.000 đồng đối với từng đương sự."},
                                        {SoKhoan: 2, NoiDung: "Đối với vụ án xin ly hôn, án phí dân sự sơ thẩm không có giá ngạch là 300.000 đồng."},
                                },
                        },
                        {
                                SoDieu: 10, TenDieu: "Miễn án phí",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Được miễn án phí trong các trường hợp sau đây:", DiemList: []Diem{
                                                {SoDiem: "a", NoiDung: "Người khiếu nại, tố cáo khởi kiện vụ án hành chính;"},
                                                {SoDiem: "b", NoiDung: "Người dưới 16 tuổi;"},
                                                {SoDiem: "c", NoiDung: "Người khuyết tật;"},
                                                {SoDiem: "d", NoiDung: "Người có công với cách mạng;"},
                                                {SoDiem: "đ", NoiDung: "Trường hợp khác do luật quy định."},
                                        }},
                                },
                        },
                },
        },

        // ───────────────── Luật 106/2025 — THA dân sự ─────────────────
        {
                ID: 5, Loai: "Luật", SoHieu: "106/2025/QH15",
                CoQuanBanHanh: "Quốc hội", NgayBanHanh: "2025-06-13",
                TrichYeu: "Luật Thi hành án dân sự",
                TomTat: "Quy định kê biên, bán đấu giá, hạ giá, thu chi tiền THA dân sự.",
                LinhVuc: "thi_hanh_an", ConHieuLuc: true,
                DieuList: []Dieu{
                        {
                                SoDieu: 45, TenDieu: "Kê biên tài sản",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Kê biên tài sản là việc cơ quan thi hành án dân sự lập biên bản về tài sản cần kê biên, quản lý tài sản đó trong thời gian chờ xử lý."},
                                        {SoKhoan: 2, NoiDung: "Thứ tự kê biên tài sản: trước tiên kê biên tiền; sau đó tài sản khác theo quy định của pháp luật."},
                                },
                        },
                        {
                                SoDieu: 52, TenDieu: "Hạ giá tài sản",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Sau 02 lần đấu giá không thành, tài sản được hạ giá. Mức hạ giá không quá 30% giá khởi điểm."},
                                        {SoKhoan: 2, NoiDung: "Sau lần hạ giá đầu tiên mà vẫn không bán được, có thể tiếp tục hạ giá thêm 20%."},
                                },
                        },
                },
        },

        // ───────────────── Luật 127/2025 — THA hình sự ─────────────────
        {
                ID: 6, Loai: "Luật", SoHieu: "127/2025/QH15",
                CoQuanBanHanh: "Quốc hội", NgayBanHanh: "2025-11-21",
                TrichYeu: "Luật Thi hành án hình sự",
                TomTat: "Quy định tạm giữ, tạm giam, phạm nhân, chế độ ăn mặc, giảm án, tha tù.",
                LinhVuc: "thi_hanh_an", ConHieuLuc: true,
                DieuList: []Dieu{
                        {
                                SoDieu: 23, TenDieu: "Chế độ ăn của phạm nhân",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Phạm nhân được cấp ăn theo định lượng quy định. Suất ăn hàng ngày phải bảo đảm dinh dưỡng, hợp vệ sinh."},
                                        {SoKhoan: 2, NoiDung: "Phạm nhân có thai, nuôi con dưới 36 tháng tuổi được tăng 20% định lượng."},
                                },
                        },
                        {
                                SoDieu: 65, TenDieu: "Giảm thời hạn chấp án hình sự",
                                KhoanList: []Khoan{
                                        {SoKhoan: 1, NoiDung: "Phạm nhân có tiến bộ được giảm 01 ngày chấp án cho mỗi 02 tháng chấp án. Phạm nhân có thành tích xuất sắc được giảm 02 ngày cho mỗi 02 tháng."},
                                },
                        },
                },
        },

        // ───────────────── QĐ 410/QĐ-VKSTC — biểu mẫu DS/HC ─────────────────
        {
                ID: 7, Loai: "Quyết định", SoHieu: "410/QĐ-VKSTC",
                CoQuanBanHanh: "Viện trưởng VKSND tối cao", NgayBanHanh: "2021-11-25",
                TrichYeu: "Quy định biểu mẫu tố tụng dân sự, hành chính",
                TomTat: "65 mẫu DS, 65 mẫu HC, 21 mẫu PL09.",
                LinhVuc: "dan_su", ConHieuLuc: true,
                DieuList: nil,
        },
}

// helper: populate BLTTHS die list (we left it nil above to avoid huge inline)
func init() {
        for i, doc := range LegalDocsStructuredSeed {
                if doc.SoHieu == "92/2015/QH13" {
                        LegalDocsStructuredSeed[i].DieuList = []Dieu{
                                {
                                        SoDieu: 119, TenDieu: "Thời hạn tạm giam",
                                        KhoanList: []Khoan{
                                                {SoKhoan: 1, NoiDung: "Thời hạn tạm giam của Thủ trưởng cơ quan điều tra quyết định không quá 04 tháng đối với tội phạm ít nghiêm trọng, tội phạm nghiêm trọng; không quá 08 tháng đối với tội phạm rất nghiêm trọng, tội phạm đặc biệt nghiêm trọng."},
                                                {SoKhoan: 2, NoiDung: "Quyết định gia hạn tạm giam lần đầu không quá 04 tháng. Tổng thời hạn tạm giam không quá 16 tháng đối với tội phạm đặc biệt nghiêm trọng thuộc trường hợp: có tổ chức, có tính chất chuyên nghiệp, lợi dụng chức vụ, quyền hạn, mục đích khủng bố, hoạt động Mafia."},
                                        },
                                },
                                {
                                        SoDieu: 128, TenDieu: "Thời hạn tạm giữ (sửa 128/2025)",
                                        KhoanList: []Khoan{
                                                {SoKhoan: 1, NoiDung: "Tạm giữ không quá 12 giờ. Trường hợp cần thiết có thể kéo dài đến 24 giờ. Tạm giữ 24 giờ áp dụng với tội phạm nghiêm trọng, rất nghiêm trọng, đặc biệt nghiêm trọng."},
                                                {SoKhoan: 2, NoiDung: "Gia hạn tạm giữ đến 48 giờ, 72 giờ theo quy định BLTTHS."},
                                                {SoKhoan: 3, NoiDung: "Khi hết thời hạn tạm giữ mà không ra quyết định khởi tố vụ án hình sự thì phải trả tự do cho người bị tạm giữ."},
                                        },
                                },
                        }
                        break
                }
        }
}
