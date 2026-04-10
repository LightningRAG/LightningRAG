package rag

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"strings"
	"testing"
)

func TestParseTextContent_UTF16LE(t *testing.T) {
	// BOM FF FE + UTF-16LE "Hi"
	b := []byte{0xFF, 0xFE, 'H', 0, 'i', 0}
	out, err := ParseTextContent(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if out != "Hi" {
		t.Fatalf("got %q", out)
	}
}

func TestParseTextContent_UTF16BE(t *testing.T) {
	b := []byte{0xFE, 0xFF, 0, 'H', 0, 'i'}
	out, err := ParseTextContent(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if out != "Hi" {
		t.Fatalf("got %q", out)
	}
}

func TestStripGzipFilename(t *testing.T) {
	if stripGzipFilename("a.TXT.GZ") != "a.TXT" {
		t.Fatalf("got %q", stripGzipFilename("a.TXT.GZ"))
	}
	if stripGzipFilename("x.gzip") != "x" {
		t.Fatal("gzip suffix")
	}
	if stripGzipFilename("backup.tgz") != "backup.tar" {
		t.Fatalf("tgz -> tar: got %q", stripGzipFilename("backup.tgz"))
	}
	if stripGzipFilename("icon.svgz") != "icon.svg" {
		t.Fatalf("svgz -> svg: got %q", stripGzipFilename("icon.svgz"))
	}
}

func TestInferFileType_Extensions(t *testing.T) {
	if inferFileType("data.ldjson") != "jsonl" {
		t.Fatalf("ldjson: got %q", inferFileType("data.ldjson"))
	}
	if inferFileType("page.mhtml") != "html" {
		t.Fatalf("mhtml: got %q", inferFileType("page.mhtml"))
	}
	if inferFileType("subs.ass") != "txt" {
		t.Fatalf("ass: got %q", inferFileType("subs.ass"))
	}
	if inferFileType("main.go") != "txt" {
		t.Fatalf("go source: got %q", inferFileType("main.go"))
	}
	if inferFileType("x.tgz") != "gz" {
		t.Fatalf("tgz: got %q", inferFileType("x.tgz"))
	}
	if inferFileType("macro.docm") != "docx" {
		t.Fatalf("docm: got %q", inferFileType("macro.docm"))
	}
	if inferFileType("book.xlsm") != "xlsx" {
		t.Fatalf("xlsm: got %q", inferFileType("book.xlsm"))
	}
	if inferFileType("slides.pptm") != "pptx" {
		t.Fatalf("pptm: got %q", inferFileType("slides.pptm"))
	}
	if inferFileType("binary.xlsb") != "xlsb" {
		t.Fatalf("xlsb: got %q", inferFileType("binary.xlsb"))
	}
	if inferFileType("app.properties") != "txt" {
		t.Fatalf("properties: got %q", inferFileType("app.properties"))
	}
	if inferFileType("refs.bib") != "txt" {
		t.Fatalf("bib: got %q", inferFileType("refs.bib"))
	}
	if inferFileType("notes.org") != "txt" {
		t.Fatalf("org: got %q", inferFileType("notes.org"))
	}
	if inferFileType("top.v") != "txt" {
		t.Fatalf("verilog: got %q", inferFileType("top.v"))
	}
	if inferFileType("main.tf") != "txt" {
		t.Fatalf("terraform: got %q", inferFileType("main.tf"))
	}
	if inferFileType("slide.pages") != "pages" {
		t.Fatalf("pages: got %q", inferFileType("slide.pages"))
	}
	if inferFileType("sheet.numbers") != "numbers" {
		t.Fatalf("numbers: got %q", inferFileType("sheet.numbers"))
	}
	if inferFileType("talk.key") != "key" {
		t.Fatalf("keynote: got %q", inferFileType("talk.key"))
	}
	if inferFileType("app.css") != "txt" {
		t.Fatalf("css: got %q", inferFileType("app.css"))
	}
	if inferFileType("x.bz2") != "bz2" {
		t.Fatalf("bz2: got %q", inferFileType("x.bz2"))
	}
	if inferFileType("trace.har") != "json" {
		t.Fatalf("har: got %q", inferFileType("trace.har"))
	}
	if inferFileType("feed.rss") != "xml" {
		t.Fatalf("rss: got %q", inferFileType("feed.rss"))
	}
	if inferFileType("icon.svgz") != "gz" {
		t.Fatalf("svgz: got %q", inferFileType("icon.svgz"))
	}
	if inferFileType("map.geojson") != "json" {
		t.Fatalf("geojson: got %q", inferFileType("map.geojson"))
	}
	if inferFileType("track.gpx") != "xml" {
		t.Fatalf("gpx: got %q", inferFileType("track.gpx"))
	}
	if inferFileType("flake.nix") != "txt" {
		t.Fatalf("nix: got %q", inferFileType("flake.nix"))
	}
	if inferFileType("video.m4v") != "video" {
		t.Fatalf("m4v: got %q", inferFileType("video.m4v"))
	}
	if inferFileType("diagram.odg") != "odg" {
		t.Fatalf("odg: got %q", inferFileType("diagram.odg"))
	}
	if inferFileType("chart.drawio") != "xml" {
		t.Fatalf("drawio: got %q", inferFileType("chart.drawio"))
	}
	if inferFileType("photo.avif") != "avif" {
		t.Fatalf("avif: got %q", inferFileType("photo.avif"))
	}
	if inferFileType("photo.jfif") != "jpg" {
		t.Fatalf("jfif: got %q", inferFileType("photo.jfif"))
	}
	if inferFileType("favicon.ico") != "ico" {
		t.Fatalf("ico: got %q", inferFileType("favicon.ico"))
	}
	if inferFileType("anim.apng") != "apng" {
		t.Fatalf("apng: got %q", inferFileType("anim.apng"))
	}
	if inferFileType("memo.wps") != "wps" {
		t.Fatalf("wps: got %q", inferFileType("memo.wps"))
	}
	if inferFileType("sheet.et") != "et" {
		t.Fatalf("et: got %q", inferFileType("sheet.et"))
	}
	if inferFileType("clip.rmvb") != "video" {
		t.Fatalf("rmvb: got %q", inferFileType("clip.rmvb"))
	}
	if inferFileType("track.wv") != "audio" {
		t.Fatalf("wavpack wv: got %q", inferFileType("track.wv"))
	}
}

func TestParseDocumentContent_GzipInnerSVGZ(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><text>Hi</text></svg>`))
	_ = gw.Close()
	s, err := parseDocumentContent(context.Background(), buf.Bytes(), "gz", "icon.svgz", 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "Hi") {
		t.Fatalf("got %q", s)
	}
}

func TestStripBzip2Filename(t *testing.T) {
	if stripBzip2Filename("a.TXT.BZ2") != "a.TXT" {
		t.Fatalf("got %q", stripBzip2Filename("a.TXT.BZ2"))
	}
	if stripBzip2Filename("backup.tbz2") != "backup.tar" {
		t.Fatalf("tbz2: got %q", stripBzip2Filename("backup.tbz2"))
	}
}

func TestParseDocumentContent_Bzip2InnerTxt(t *testing.T) {
	// 与 docparse 测例相同：bzip2(b"bzip inner")
	raw, err := hex.DecodeString("425a683931415926535945f2158300000091804000122150102000310c0821a68da9d2330a0f17724538509045f21583")
	if err != nil {
		t.Fatal(err)
	}
	s, err := parseDocumentContent(context.Background(), raw, "bz2", "note.txt.bz2", 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(s) != "bzip inner" {
		t.Fatalf("got %q", s)
	}
}

func TestParseDocumentContent_GzipInnerTxt(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write([]byte("inner text"))
	_ = gw.Close()
	s, err := parseDocumentContent(context.Background(), buf.Bytes(), "gz", "note.txt.gz", 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(s) != "inner text" {
		t.Fatalf("got %q", s)
	}
}
