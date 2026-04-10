package docparse

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

func TestParseJSONText_Object(t *testing.T) {
	s := `{"a":1,"b":"x"}`
	out, err := ParseJSONText([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"a"`) || !strings.Contains(out, `"b"`) {
		t.Fatalf("unexpected: %q", out)
	}
}

func TestParseJSONText_JSONL(t *testing.T) {
	s := "{\"x\": 1}\n{\"y\": 2}\n"
	out, err := ParseJSONText([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 records, got %q", out)
	}
}

func TestExtractPptxSlideText(t *testing.T) {
	xml := `<xml xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:t>Hello</a:t><a:t> World</a:t></xml>`
	got := extractPptxSlideText(xml)
	if got != "Hello\nWorld" {
		t.Fatalf("got %q", got)
	}
}

func TestParseTOMLText(t *testing.T) {
	out, err := ParseTOMLText([]byte(`a = 1
[nested]
b = "x"
`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "nested") {
		t.Fatalf("got %q", out)
	}
}

func TestParseRTFPlainText(t *testing.T) {
	raw := []byte(`{\rtf1\ansi Hello\par World}`)
	out, err := ParseRTFPlainText(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Hello") || !strings.Contains(out, "World") {
		t.Fatalf("got %q", out)
	}
}

func TestParseYAMLText(t *testing.T) {
	out, err := ParseYAMLText([]byte("a: 1\nb: hello\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("got %q", out)
	}
}

func TestHTMLToPlainText_StripsScriptAndStyle(t *testing.T) {
	html := `<!doctype html><html><head><style>body{display:none}</style></head><body><p>Hi</p><script>evil()</script><noscript>NS_FALLBACK_X</noscript></body></html>`
	out := HTMLToPlainText(html)
	if strings.Contains(out, "evil") || strings.Contains(out, "display") || strings.Contains(out, "NS_FALLBACK_X") {
		t.Fatalf("script/style/noscript leaked: %q", out)
	}
	if !strings.Contains(out, "Hi") {
		t.Fatalf("missing body text: %q", out)
	}
}

func TestShouldPreferMIMEOverExt(t *testing.T) {
	if !ShouldPreferMIMEOverExt("txt", "pdf") {
		t.Fatal("txt should yield to pdf MIME")
	}
	if ShouldPreferMIMEOverExt("pdf", "docx") {
		t.Fatal("pdf should not yield to docx MIME")
	}
	if !ShouldPreferMIMEOverExt("xml", "json") {
		t.Fatal("xml should yield to json MIME")
	}
}

func TestFileTypeFromMIME_JSONLAndOOXML(t *testing.T) {
	if FileTypeFromMIME("application/x-ndjson") != "jsonl" {
		t.Fatal("ndjson MIME")
	}
	if FileTypeFromMIME("application/vnd.ms-word.document.macroenabled.12") != "docx" {
		t.Fatal("docm MIME")
	}
	if FileTypeFromMIME("application/vnd.ms-excel.sheet.binary.macroenabled.12") != "xlsb" {
		t.Fatal("xlsb MIME")
	}
	if FileTypeFromMIME("application/x-bibtex") != "txt" {
		t.Fatal("bibtex MIME")
	}
	if FileTypeFromMIME("text/x-org") != "txt" {
		t.Fatal("org MIME")
	}
	if FileTypeFromMIME("application/x-bzip2") != "bz2" {
		t.Fatal("bzip2 MIME")
	}
	if FileTypeFromMIME("text/x-markdown") != "md" {
		t.Fatal("text/x-markdown MIME")
	}
	if FileTypeFromMIME("application/vnd.apple.pages") != "pages" {
		t.Fatal("pages MIME")
	}
	if FileTypeFromMIME("text/css") != "txt" {
		t.Fatal("text/css MIME")
	}
	if FileTypeFromMIME("image/svg+xml") != "svg" {
		t.Fatal("svg MIME")
	}
	if FileTypeFromMIME("text/comma-separated-values") != "csv" {
		t.Fatal("text/comma-separated-values MIME")
	}
	if FileTypeFromMIME("application/hal+json") != "json" {
		t.Fatal("hal+json MIME")
	}
	if FileTypeFromMIME("application/rss+xml") != "xml" {
		t.Fatal("rss+xml MIME")
	}
	if FileTypeFromMIME("application/geo+json") != "json" {
		t.Fatal("geo+json MIME")
	}
	if FileTypeFromMIME("application/gpx+xml") != "xml" {
		t.Fatal("gpx+xml MIME")
	}
	if FileTypeFromMIME("application/vnd.oasis.opendocument.graphics") != "odg" {
		t.Fatal("odg MIME")
	}
	if FileTypeFromMIME("image/avif") != "avif" {
		t.Fatal("avif MIME")
	}
	if FileTypeFromMIME("application/schema+json") != "json" {
		t.Fatal("schema+json MIME")
	}
	if FileTypeFromMIME("image/vnd.microsoft.icon") != "ico" {
		t.Fatal("ico MIME")
	}
	if FileTypeFromMIME("image/apng") != "apng" {
		t.Fatal("apng MIME")
	}
	if FileTypeFromMIME("application/vnd.wps-office.wps") != "wps" {
		t.Fatal("wps MIME")
	}
	if FileTypeFromMIME("application/vnd.wps-office.et") != "et" {
		t.Fatal("et MIME")
	}
	if FileTypeFromMIME("application/vnd.wps-office.dps") != "dps" {
		t.Fatal("dps MIME")
	}
}

func TestRefineFileTypeByContent_PDFMagic(t *testing.T) {
	got := RefineFileTypeByContent("wrong.txt", "txt", []byte("%PDF-1.4\n%âãÏÓ\n"))
	if got != "pdf" {
		t.Fatalf("want pdf, got %q", got)
	}
}

// bzip2CompressedBzipInner 为 printf 'bzip inner' | bzip2 的原始字节（用于无 Writer 环境下的解压测例）。
var bzip2CompressedBzipInner, _ = hex.DecodeString("425a683931415926535945f2158300000091804000122150102000310c0821a68da9d2330a0f17724538509045f21583")

func TestRefineFileTypeByContent_Bzip2Magic(t *testing.T) {
	if !IsBzip2Magic(bzip2CompressedBzipInner) {
		t.Fatal("expected bzip2 magic")
	}
	got := RefineFileTypeByContent("blob.bin", "txt", bzip2CompressedBzipInner)
	if got != "bz2" {
		t.Fatalf("want bz2, got %q", got)
	}
}

func TestRefineFileTypeByContent_GzipMagic(t *testing.T) {
	gzHead := []byte{0x1f, 0x8b, 8, 0}
	got := RefineFileTypeByContent("data.gz", "gz", gzHead)
	if got != "gz" {
		t.Fatalf("want gz, got %q", got)
	}
	// 无 .gz 后缀但标成 txt 时也应纠偏
	got2 := RefineFileTypeByContent("upload.bin", "txt", gzHead)
	if got2 != "gz" {
		t.Fatalf("want gz for txt+gzip magic, got %q", got2)
	}
	got3 := RefineFileTypeByContent("icon.svgz", "svg", gzHead)
	if got3 != "gz" {
		t.Fatalf("want gz for .svgz+gzip magic, got %q", got3)
	}
}

func TestParseIPynbText(t *testing.T) {
	raw := `{"cells":[{"cell_type":"markdown","source":"# Hi\n"},{"cell_type":"code","source":["print(","1",")"]}]}`
	out, err := ParseIPynbText([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Hi") || !strings.Contains(out, "print") {
		t.Fatalf("got %q", out)
	}
}

func TestRefineFileTypeByContent_WPSOfficeOOXMLAliases(t *testing.T) {
	var wpsBuf bytes.Buffer
	zw := zip.NewWriter(&wpsBuf)
	wf, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wf.Write([]byte(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>X</w:t></w:r></w:p></w:body></w:document>`)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if got := RefineFileTypeByContent("memo.wps", "wps", wpsBuf.Bytes()); got != "docx" {
		t.Fatalf("wps zip -> docx: got %q", got)
	}

	var etBuf bytes.Buffer
	zet := zip.NewWriter(&etBuf)
	ef, err := zet.Create("xl/workbook.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ef.Write([]byte(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"/>`)); err != nil {
		t.Fatal(err)
	}
	if err := zet.Close(); err != nil {
		t.Fatal(err)
	}
	if got := RefineFileTypeByContent("book.et", "et", etBuf.Bytes()); got != "xlsx" {
		t.Fatalf("et zip -> xlsx: got %q", got)
	}

	var dpsBuf bytes.Buffer
	zd := zip.NewWriter(&dpsBuf)
	pf, err := zd.Create("ppt/presentation.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pf.Write([]byte(`<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"/>`)); err != nil {
		t.Fatal(err)
	}
	if err := zd.Close(); err != nil {
		t.Fatal(err)
	}
	if got := RefineFileTypeByContent("slides.dps", "dps", dpsBuf.Bytes()); got != "pptx" {
		t.Fatalf("dps zip -> pptx: got %q", got)
	}
}

func TestNormalizeExcelFileType(t *testing.T) {
	zipHead := []byte{'P', 'K', 3, 4}
	if NormalizeExcelFileType("xls", zipHead) != "xlsx" {
		t.Fatal("expected xlsx from zip magic")
	}
	oleHead := []byte{0xD0, 0xCF, 0x11, 0xE0}
	if NormalizeExcelFileType("xlsx", oleHead) != "xls" {
		t.Fatal("expected xls from ole magic")
	}
}

func TestParseEMLText_Minimal(t *testing.T) {
	raw := "From: a@b.com\r\nTo: c@d.com\r\nSubject: T\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nBody line\r\n"
	out, err := ParseEMLText([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Body line") || !strings.Contains(out, "Subject") {
		t.Fatalf("got %q", out)
	}
}
