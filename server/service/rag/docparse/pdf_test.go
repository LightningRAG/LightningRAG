package docparse

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"github.com/ledongthuc/pdf"
)

func TestFormatPDFOutlineLines(t *testing.T) {
	root := pdf.Outline{
		Child: []pdf.Outline{
			{Title: "Chapter 1", Child: []pdf.Outline{{Title: "1.1 Intro"}}},
			{Title: "Chapter 2"},
		},
	}
	got := formatPDFOutlineLines(root)
	if len(got) != 3 {
		t.Fatalf("want 3 lines, got %d: %q", len(got), got)
	}
	if got[0] != "Chapter 1" || got[1] != "  1.1 Intro" || got[2] != "Chapter 2" {
		t.Fatalf("unexpected: %#v", got)
	}
	if formatPDFOutlineLines(pdf.Outline{}) != nil {
		t.Fatal("empty outline should be nil")
	}
}

func TestParsePDFFromBytes_LedongthucGofpdf(t *testing.T) {
	t.Setenv("LIGHTNINGRAG_PDF_ENGINE", "ledongthuc")
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	p.SetFont("Helvetica", "", 14)
	p.Cell(40, 10, "LRAG_PDF_TEST_LINE")
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	s, err := ParsePDFFromBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "LRAG_PDF_TEST_LINE") {
		t.Fatalf("missing text: %q", s)
	}
}

func gofpdfMinimalPDF(t *testing.T) []byte {
	t.Helper()
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	p.SetFont("Helvetica", "", 14)
	p.Cell(40, 10, "PDF_GO_ENGINE_LINE")
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestParsePDFFromBytes_PdfGoEngine(t *testing.T) {
	t.Setenv("LIGHTNINGRAG_PDF_ENGINE", "pdfgo")
	data := gofpdfMinimalPDF(t)
	s, err := ParsePDFFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "PDF_GO_ENGINE_LINE") {
		t.Fatalf("pdf-go missing text: %q", s)
	}
}

func TestParsePDFFromBytes_PypdfEnvAliasPdfGo(t *testing.T) {
	t.Setenv("LIGHTNINGRAG_PDF_ENGINE", "pypdf")
	data := gofpdfMinimalPDF(t)
	s, err := ParsePDFFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "PDF_GO_ENGINE_LINE") {
		t.Fatalf("pypdf alias missing text: %q", s)
	}
}

func TestGetPDFPageCount_PdfGo(t *testing.T) {
	t.Setenv("LIGHTNINGRAG_PDF_ENGINE", "pdfgo")
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	p.SetFont("Helvetica", "", 12)
	p.Cell(40, 10, "A")
	p.AddPage()
	p.Cell(40, 10, "B")
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	n, err := GetPDFPageCount(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("page count: want 2, got %d", n)
	}
}

func TestParsePDFByPage_PdfGo(t *testing.T) {
	t.Setenv("LIGHTNINGRAG_PDF_ENGINE", "pdfgo")
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	p.SetFont("Helvetica", "", 12)
	p.Cell(40, 10, "PAGE_ONE_MARKER")
	p.AddPage()
	p.Cell(40, 10, "PAGE_TWO_MARKER")
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	pages, err := ParsePDFByPage(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 {
		t.Fatalf("want 2 pages, got %d", len(pages))
	}
	if !strings.Contains(pages[0], "PAGE_ONE_MARKER") || !strings.Contains(pages[1], "PAGE_TWO_MARKER") {
		t.Fatalf("unexpected pages: %#v", pages)
	}
}

func TestExtractPDFPypdfDocInfo_PdfGo(t *testing.T) {
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	p.SetFont("Helvetica", "", 12)
	p.Cell(40, 10, "X")
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	info, err := ExtractPDFPypdfDocInfo(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if info.PageCount < 1 {
		t.Fatalf("page_count: %+v", info)
	}
	if info.PDFHeader == "" {
		t.Fatal("expected pdf_header")
	}
}

func TestExtractPDFPypdfPageLabels_PdfGo(t *testing.T) {
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	p.SetFont("Helvetica", "", 12)
	p.Cell(40, 10, "x")
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	labels, err := ExtractPDFPypdfPageLabels(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 1 {
		t.Fatalf("want 1 page label, got %d %#v", len(labels), labels)
	}
}

func TestExtractPDFPypdfURILinks_PdfGo(t *testing.T) {
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	p.SetFont("Helvetica", "", 12)
	p.Cell(40, 10, "x")
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	links, err := ExtractPDFPypdfURILinks(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if links == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestExtractPDFPypdfAttachmentNames_PdfGo(t *testing.T) {
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	names, err := ExtractPDFPypdfAttachmentNames(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if names == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestExtractPDFPypdfXMPMetadata_PdfGo(t *testing.T) {
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	xmp, err := ExtractPDFPypdfXMPMetadata(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	_ = xmp
}
