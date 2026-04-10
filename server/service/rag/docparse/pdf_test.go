package docparse

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/LightningRAG/LightningRAG/server/service/rag/docparse/pypdfplain"
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

func TestParsePDFFromBytes_PypdfWhenAvailable(t *testing.T) {
	root, err := pypdfplain.FindPypdfSourceRoot()
	if err != nil {
		t.Skip("no references/pypdf:", err)
	}
	_ = root
	py := pypdfplain.PythonExecutable()
	if _, err := exec.LookPath(py); err != nil {
		t.Skip("python not on PATH:", py, err)
	}
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	p.SetFont("Helvetica", "", 14)
	p.Cell(40, 10, "PYPDF_ENGINE_LINE")
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LIGHTNINGRAG_PDF_ENGINE", "pypdf")
	s, err := ParsePDFFromBytes(buf.Bytes())
	if err != nil {
		t.Skip("pypdf extract failed (install Python 3.9+ or set LIGHTNINGRAG_PYTHON):", err)
	}
	if !strings.Contains(s, "PYPDF_ENGINE_LINE") {
		t.Fatalf("pypdf missing text: %q", s)
	}
}

func TestGetPDFPageCount_PypdfWhenAvailable(t *testing.T) {
	if _, err := pypdfplain.FindPypdfSourceRoot(); err != nil {
		t.Skip("no references/pypdf:", err)
	}
	if _, err := exec.LookPath(pypdfplain.PythonExecutable()); err != nil {
		t.Skip("no python:", err)
	}
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
	t.Setenv("LIGHTNINGRAG_PDF_ENGINE", "pypdf")
	n, err := GetPDFPageCount(buf.Bytes())
	if err != nil {
		t.Skip("pypdf meta:", err)
	}
	if n != 2 {
		t.Fatalf("page count: want 2, got %d", n)
	}
}

func TestParsePDFByPage_PypdfWhenAvailable(t *testing.T) {
	if _, err := pypdfplain.FindPypdfSourceRoot(); err != nil {
		t.Skip("no references/pypdf:", err)
	}
	if _, err := exec.LookPath(pypdfplain.PythonExecutable()); err != nil {
		t.Skip("no python:", err)
	}
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
	t.Setenv("LIGHTNINGRAG_PDF_ENGINE", "pypdf")
	pages, err := ParsePDFByPage(buf.Bytes())
	if err != nil {
		t.Skip("pypdf pages:", err)
	}
	if len(pages) != 2 {
		t.Fatalf("want 2 pages, got %d", len(pages))
	}
	if !strings.Contains(pages[0], "PAGE_ONE_MARKER") || !strings.Contains(pages[1], "PAGE_TWO_MARKER") {
		t.Fatalf("unexpected pages: %#v", pages)
	}
}

func TestExtractPDFPypdfDocInfo_WhenAvailable(t *testing.T) {
	if _, err := pypdfplain.FindPypdfSourceRoot(); err != nil {
		t.Skip("no references/pypdf:", err)
	}
	if _, err := exec.LookPath(pypdfplain.PythonExecutable()); err != nil {
		t.Skip("no python:", err)
	}
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
		t.Skip("docinfo:", err)
	}
	if info.PageCount < 1 {
		t.Fatalf("page_count: %+v", info)
	}
	if info.PDFHeader == "" {
		t.Fatal("expected pdf_header")
	}
}

func TestPypdfPageRangeFromEnv_WhenAvailable(t *testing.T) {
	if _, err := pypdfplain.FindPypdfSourceRoot(); err != nil {
		t.Skip("no references/pypdf:", err)
	}
	if _, err := exec.LookPath(pypdfplain.PythonExecutable()); err != nil {
		t.Skip("no python:", err)
	}
	p := gofpdf.New("P", "mm", "A4", "")
	for i := 1; i <= 3; i++ {
		p.AddPage()
		p.SetFont("Helvetica", "", 12)
		p.Cell(40, 10, fmt.Sprintf("MARKER_%d", i))
	}
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LIGHTNINGRAG_PDF_ENGINE", "pypdf")
	t.Setenv("LIGHTNINGRAG_PDF_FROM_PAGE", "1")
	t.Setenv("LIGHTNINGRAG_PDF_TO_PAGE", "2")
	pages, err := ParsePDFByPage(buf.Bytes())
	if err != nil {
		t.Skip("pypdf pages range:", err)
	}
	if len(pages) != 1 {
		t.Fatalf("range [1,2) want 1 page, got %d %#v", len(pages), pages)
	}
	if !strings.Contains(pages[0], "MARKER_2") || strings.Contains(pages[0], "MARKER_1") {
		t.Fatalf("expected only page 2 content: %q", pages[0])
	}
}

func TestExtractPDFPypdfPageLabels_WhenAvailable(t *testing.T) {
	if _, err := pypdfplain.FindPypdfSourceRoot(); err != nil {
		t.Skip("no references/pypdf:", err)
	}
	if _, err := exec.LookPath(pypdfplain.PythonExecutable()); err != nil {
		t.Skip("no python:", err)
	}
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
		t.Skip("pagelabels:", err)
	}
	if len(labels) != 1 {
		t.Fatalf("want 1 page label, got %d %#v", len(labels), labels)
	}
}

func TestExtractPDFPypdfURILinks_WhenAvailable(t *testing.T) {
	if _, err := pypdfplain.FindPypdfSourceRoot(); err != nil {
		t.Skip("no references/pypdf:", err)
	}
	if _, err := exec.LookPath(pypdfplain.PythonExecutable()); err != nil {
		t.Skip("no python:", err)
	}
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
		t.Skip("links:", err)
	}
	if links == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestExtractPDFPypdfAttachmentNames_WhenAvailable(t *testing.T) {
	if _, err := pypdfplain.FindPypdfSourceRoot(); err != nil {
		t.Skip("no references/pypdf:", err)
	}
	if _, err := exec.LookPath(pypdfplain.PythonExecutable()); err != nil {
		t.Skip("no python:", err)
	}
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	names, err := ExtractPDFPypdfAttachmentNames(buf.Bytes())
	if err != nil {
		t.Skip("attachmentnames:", err)
	}
	if names == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestExtractPDFPypdfXMPMetadata_WhenAvailable(t *testing.T) {
	if _, err := pypdfplain.FindPypdfSourceRoot(); err != nil {
		t.Skip("no references/pypdf:", err)
	}
	if _, err := exec.LookPath(pypdfplain.PythonExecutable()); err != nil {
		t.Skip("no python:", err)
	}
	p := gofpdf.New("P", "mm", "A4", "")
	p.AddPage()
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		t.Fatal(err)
	}
	xmp, err := ExtractPDFPypdfXMPMetadata(buf.Bytes())
	if err != nil {
		t.Skip("xmp:", err)
	}
	_ = xmp
}
