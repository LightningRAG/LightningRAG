package rag

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestExtractIWorkPreviewPDF_QuickLook(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("QuickLook/Preview.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("%PDF-1.4 test")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := extractIWorkPreviewPDF(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "%PDF-1.4 test" {
		t.Fatalf("got %q", out)
	}
}

func TestExtractIWorkPreviewPDF_SuffixFallback(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("foo/bar/preview.pdf")
	_, _ = w.Write([]byte("inner"))
	_ = zw.Close()
	out, err := extractIWorkPreviewPDF(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "inner" {
		t.Fatalf("got %q", out)
	}
}

func TestExtractIWorkPreviewPDF_Missing(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	_, _ = zw.Create("other.txt")
	_ = zw.Close()
	_, err := extractIWorkPreviewPDF(buf.Bytes())
	if err == nil || !strings.Contains(err.Error(), "Preview.pdf") {
		t.Fatalf("want missing preview error, got %v", err)
	}
}

func TestExtractIWorkPreviewPDF_NotZip(t *testing.T) {
	_, err := extractIWorkPreviewPDF([]byte("not a zip"))
	if err == nil || !strings.Contains(err.Error(), "ZIP") {
		t.Fatalf("got %v", err)
	}
}
