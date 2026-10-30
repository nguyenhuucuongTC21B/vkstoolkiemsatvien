// Package docxgen — minimal pure-Go .docx file generator.
// Không dùng deps ngoài — tạo file .docx hợp lệ bằng archive/zip + encoding/xml.
//
// .docx là file ZIP chứa các XML file theo cấu trúc Open XML.
// Files tối thiểu cần có:
//   - [Content_Types].xml
//   - _rels/.rels
//   - word/document.xml        (nội dung chính)
//   - word/_rels/document.xml.rels
//   - word/styles.xml          (tuỳ chọn, cho format đẹp)
//
// API: docxgen.New() → AddHeading/AddParagraph/AddTable/... → Save(path)
package docxgen

import (
	"archive/zip"
	"fmt"
	"os"
	"strings"
)

// Document — một tài liệu docx đang build.
type Document struct {
	paragraphs []string
}

// New tạo document rỗng.
func New() *Document {
	return &Document{}
}

// AddHeading thêm heading level 1-3.
func (d *Document) AddHeading(text string, level int) {
	if level < 1 {
		level = 1
	}
	if level > 3 {
		level = 3
	}
	style := fmt.Sprintf("Heading%d", level)
	d.paragraphs = append(d.paragraphs, fmt.Sprintf(
		`<w:p><w:pPr><w:pStyle w:val="%s"/></w:pPr><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`,
		style, escapeXML(text),
	))
}

// AddParagraph thêm paragraph thường, có hỗ trợ bold toàn bộ.
// Bắt đầu bằng "**" + kết thúc bằng "**" để đánh dấu bold.
func (d *Document) AddParagraph(text string) {
	bold := false
	if strings.HasPrefix(text, "**") && strings.HasSuffix(text, "**") {
		bold = true
		text = text[2 : len(text)-2]
	}
	runProps := ""
	if bold {
		runProps = "<w:rPr><w:b/></w:rPr>"
	}
	d.paragraphs = append(d.paragraphs, fmt.Sprintf(
		`<w:p><w:r>%s<w:t xml:space="preserve">%s</w:t></w:r></w:p>`,
		runProps, escapeXML(text),
	))
}

// AddCentered thêm paragraph căn giữa.
func (d *Document) AddCentered(text string, bold bool) {
	runProps := ""
	if bold {
		runProps = "<w:rPr><w:b/></w:rPr>"
	}
	d.paragraphs = append(d.paragraphs, fmt.Sprintf(
		`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r>%s<w:t xml:space="preserve">%s</w:t></w:r></w:p>`,
		runProps, escapeXML(text),
	))
}

// AddListItem thêm bullet item.
func (d *Document) AddListItem(text string) {
	d.paragraphs = append(d.paragraphs, fmt.Sprintf(
		`<w:p><w:pPr><w:pStyle w:val="ListParagraph"/></w:pPr><w:r><w:t xml:space="preserve">• %s</w:t></w:r></w:p>`,
		escapeXML(text),
	))
}

// AddEmptyLine thêm dòng trống.
func (d *Document) AddEmptyLine() {
	d.paragraphs = append(d.paragraphs, `<w:p/>`)
}

// AddPageBreak thêm page break.
func (d *Document) AddPageBreak() {
	d.paragraphs = append(d.paragraphs,
		`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
}

// AddTable thêm bảng đơn giản (rows[0] là header).
func (d *Document) AddTable(rows [][]string) {
	if len(rows) == 0 {
		return
	}
	cols := len(rows[0])
	var b strings.Builder
	b.WriteString(`<w:tbl>`)
	b.WriteString(`<w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblBorders>`)
	for _, side := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		b.WriteString(fmt.Sprintf(`<w:%s w:val="single" w:sz="4" w:space="0" w:color="000000"/>`, side))
	}
	b.WriteString(`</w:tblBorders></w:tblPr>`)
	// Header row
	b.WriteString(`<w:tr>`)
	for c := 0; c < cols; c++ {
		cellText := ""
		if c < len(rows[0]) {
			cellText = rows[0][c]
		}
		b.WriteString(fmt.Sprintf(
			`<w:tc><w:tcPr><w:tcW w:w="0" w:type="auto"/><w:shd w:val="clear" w:color="auto" w:fill="3E525C"/></w:tcPr>`+
				`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/></w:rPr>`+
				`<w:t xml:space="preserve">%s</w:t></w:r></w:p></w:tc>`,
			escapeXML(cellText),
		))
	}
	b.WriteString(`</w:tr>`)
	// Body rows
	for r := 1; r < len(rows); r++ {
		b.WriteString(`<w:tr>`)
		for c := 0; c < cols; c++ {
			cellText := ""
			if c < len(rows[r]) {
				cellText = rows[r][c]
			}
			b.WriteString(fmt.Sprintf(
				`<w:tc><w:tcPr><w:tcW w:w="0" w:type="auto"/></w:tcPr>`+
					`<w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p></w:tc>`,
				escapeXML(cellText),
			))
		}
		b.WriteString(`</w:tr>`)
	}
	b.WriteString(`</w:tbl>`)
	d.paragraphs = append(d.paragraphs, b.String())
}

// AddKV thêm dòng "Label: value" với label in đậm.
func (d *Document) AddKV(label, value string) {
	d.paragraphs = append(d.paragraphs, fmt.Sprintf(
		`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">%s: </w:t></w:r>`+
			`<w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`,
		escapeXML(label), escapeXML(value),
	))
}

// AddSignatureBlock thêm khối ký 2 cột.
func (d *Document) AddSignatureBlock(leftTitle, leftName, rightTitle, rightName string) {
	d.AddTable([][]string{
		{leftTitle, rightTitle},
		{"", ""},
		{"", ""},
		{"(Ký, ghi rõ họ tên)", "(Ký, ghi rõ họ tên)"},
		{leftName, rightName},
	})
}

// Save ghi document ra file .docx (ZIP).
func (d *Document) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()

	files := []struct {
		Name    string
		Content string
	}{
		{
			Name: "[Content_Types].xml",
			Content: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`,
		},
		{
			Name: "_rels/.rels",
			Content: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`,
		},
		{
			Name: "word/_rels/document.xml.rels",
			Content: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`,
		},
		{
			Name: "word/styles.xml",
			Content: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Heading1">
    <w:name w:val="heading 1"/>
    <w:pPr><w:spacing w:before="240" w:after="120"/><w:outlineLvl w:val="0"/></w:pPr>
    <w:rPr><w:b/><w:sz w:val="32"/></w:rPr>
  </w:style>
  <w:style w:type="paragraph" w:styleId="Heading2">
    <w:name w:val="heading 2"/>
    <w:pPr><w:spacing w:before="200" w:after="100"/><w:outlineLvl w:val="1"/></w:pPr>
    <w:rPr><w:b/><w:sz w:val="26"/></w:rPr>
  </w:style>
  <w:style w:type="paragraph" w:styleId="Heading3">
    <w:name w:val="heading 3"/>
    <w:pPr><w:spacing w:before="160" w:after="80"/><w:outlineLvl w:val="2"/></w:pPr>
    <w:rPr><w:b/><w:sz w:val="22"/></w:rPr>
  </w:style>
  <w:style w:type="paragraph" w:styleId="ListParagraph">
    <w:name w:val="List Paragraph"/>
    <w:pPr><w:ind w:left="720"/></w:pPr>
  </w:style>
</w:styles>`,
		},
	}

	for _, file := range files {
		w, err := zw.Create(file.Name)
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(file.Content)); err != nil {
			return err
		}
	}

	// word/document.xml
	var body strings.Builder
	body.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	body.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	body.WriteString(`<w:body>`)
	for _, p := range d.paragraphs {
		body.WriteString(p)
	}
	body.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/></w:sectPr>`)
	body.WriteString(`</w:body></w:document>`)

	w, err := zw.Create("word/document.xml")
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(body.String()))
	return err
}

// escapeXML escapes special XML chars.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
