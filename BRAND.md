# BRAND.md — Quy định thương hiệu & bản quyền (persistent)

> **File này được ghi nhớ cho mọi phiên bản sau này của phần mềm.**
> Mọi thay đổi UI/UX phải tuân thủ các quy định trong file này.

## Logo & Icon

| Asset | File nguồn | Vị trí trong project | Cách sử dụng |
|---|---|---|---|
| Logo chính | `VKS PNG 192.png` (192×192 PNG) | `assets/logo.png` + `frontend/dist/logo.png` | Hiển thị ở sidebar (brand-mark), splash screen, header popup bản quyền |
| Icon Windows | `VKS128.ico` (128×128 ICO) | `build/windows/icon.ico` | Icon file `.exe`, icon cửa sổ trên Windows |
| Icon Linux | `logo.png` | `build/linux/icon.png` | Icon file binary, icon desktop entry |

**Logo là huy hiệu ngành Kiểm sát** (Viện Kiểm sát nhân dân) — nền đen, huy hiệu đỏ với ngôi sao vàng 5 cánh, lúa vàng, bánh răng, hai kiếm bắt chéo che bằng tròn xanh có chữ "KS", dải băng đỏ ghi "CỘNG HÒA XÃ HỘI CHỦ NGHĨA VIỆT NAM".

⚠️ **Không được thay logo/icon** bằng bất kỳ hình ảnh nào khác trong mọi phiên bản sau này trừ khi được chủ sở hữu yêu cầu rõ ràng.

## Footer + nút Bản quyền

**Bắt buộc có ở mọi màn hình** trong sidebar footer (hoặc bottom bar nếu layout thay đổi):

1. **Badge "100% Offline"** — luôn hiện trạng thái offline
2. **DB stats** — số văn bản pháp luật trong CSDL
3. **Nút "Bản quyền"** — clickable, mở popup bản quyền

## Popup Bản quyền

Khi click nút "Bản quyền", hiện modal popup chứa:

1. **Ảnh avatarC.jpg** (576×1024) hiển thị đầy đủ, có thể cuộn nếu cần
2. **Thông tin coder**:
   ```
   Coder Developed by Nguyễn Hữu Cường
   Chức vụ: CV CY KTCC

   Liên hệ: 0905 6666 24
   Email: doremon.dn01@gmail.com
   ```
3. **Nút "Đóng"** để đóng popup

**Asset**: `frontend/dist/avatarC.jpg` (đã embed vào binary qua go:embed frontend/dist)

## Cấu hình Wails

`wails.json` phải có:
```json
"info": {
  "companyName": "Viện Kiểm sát nhân dân",
  "productName": "Công cụ ngành Kiểm sát",
  "productVersion": "0.1.0",
  "copyright": "© 2026 Nguyễn Hữu Cường — doremon.dn01@gmail.com",
  "comments": "Bộ công cụ nghiệp vụ Kiểm sát offline, single binary Windows/Linux."
}
```

Windows metadata (khi build .exe) sẽ hiển thị:
- File description: "Công cụ ngành Kiểm sát"
- Product name: "Công cụ ngành Kiểm sát"
- Copyright: "© 2026 Nguyễn Hữu Cường — doremon.dn01@gmail.com"

## Tác giả / liên hệ

- **Coder**: Nguyễn Hữu Cường
- **Chức vụ**: CV CY KTCC
- **Điện thoại**: 0905 6666 24
- **Email**: doremon.dn01@gmail.com

## Lịch sử thay đổi

- **v0.1.0** (2026-10-30): Thêm logo VKS, icon Windows, avatar coder, popup bản quyền.
