package docparse

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	lrpdf "github.com/lightningrag/pdf-go/pdf"
)

func pdfGoOpenReader(data []byte) (*lrpdf.PdfReader, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("PDF 数据为空")
	}
	return lrpdf.NewPdfReader(bytes.NewReader(data), false)
}

func pdfHeaderLineFromData(data []byte) string {
	idx := bytes.Index(data, []byte("%PDF-"))
	if idx < 0 {
		return ""
	}
	rest := data[idx:]
	if nl := bytes.IndexByte(rest, '\n'); nl >= 0 {
		return strings.TrimSpace(string(rest[:nl]))
	}
	return strings.TrimSpace(string(rest))
}

func extractPlainFullPdfGo(data []byte) (string, error) {
	r, err := pdfGoOpenReader(data)
	if err != nil {
		return "", fmt.Errorf("解析 PDF 失败: %w", err)
	}
	defer r.Close()
	n, err := r.NumPages()
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", fmt.Errorf("PDF 未提取到文本（可能是扫描件或图片型 PDF，需 OCR）")
	}
	// 与 pdf-go readtextadvanced 示例一致：每页 ExtractTextAdvanced，单页失败不阻断其余页
	opts := lrpdf.ExtractTextOptions{}
	var sb strings.Builder
	for i := 0; i < n; i++ {
		p, err := r.Page(i)
		if err != nil {
			continue
		}
		txt, err := p.ExtractTextAdvanced(opts)
		if err != nil {
			continue
		}
		t := strings.TrimSpace(txt)
		if t == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(t)
	}
	s := strings.TrimSpace(sb.String())
	if s == "" {
		return "", fmt.Errorf("PDF 未提取到文本（可能是扫描件或图片型 PDF，需 OCR）")
	}
	return s, nil
}

func extractPagesPdfGo(data []byte) ([]string, error) {
	r, err := pdfGoOpenReader(data)
	if err != nil {
		return nil, fmt.Errorf("解析 PDF 失败: %w", err)
	}
	defer r.Close()
	n, err := r.NumPages()
	if err != nil {
		return nil, err
	}
	opts := lrpdf.ExtractTextOptions{}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		p, err := r.Page(i)
		if err != nil {
			out[i] = ""
			continue
		}
		txt, err := p.ExtractTextAdvanced(opts)
		if err != nil {
			out[i] = ""
			continue
		}
		out[i] = txt
	}
	return out, nil
}

func extractPageCountPdfGo(data []byte) (int, error) {
	r, err := pdfGoOpenReader(data)
	if err != nil {
		return 0, err
	}
	return r.NumPages()
}

func extractDocInfoPdfGo(data []byte) (*PDFPypdfDocInfo, error) {
	r, err := pdfGoOpenReader(data)
	if err != nil {
		return nil, err
	}
	n, err := r.NumPages()
	if err != nil {
		return nil, err
	}
	info := &PDFPypdfDocInfo{
		PageCount: n,
		PDFHeader: pdfHeaderLineFromData(data),
	}
	di, err := r.DocumentInformation()
	if err != nil {
		return info, err
	}
	if di == nil {
		return info, nil
	}
	info.Title = di.TitleWithReader(r)
	info.Author = di.AuthorWithReader(r)
	info.Subject = di.SubjectWithReader(r)
	info.Creator = di.CreatorWithReader(r)
	info.Producer = di.ProducerWithReader(r)
	info.Keywords = di.KeywordsWithReader(r)
	info.CreationDate = di.CreationDateWithReader(r)
	info.ModificationDate = di.ModDateWithReader(r)
	if o, ok := di.Raw(lrpdf.InfoCreationDate); ok {
		info.CreationDateRaw = fmt.Sprintf("%v", o)
	}
	if o, ok := di.Raw(lrpdf.InfoModDate); ok {
		info.ModificationDateRaw = fmt.Sprintf("%v", o)
	}
	return info, nil
}

func extractURILinksPdfGo(data []byte) ([]string, error) {
	r, err := pdfGoOpenReader(data)
	if err != nil {
		return nil, err
	}
	n, err := r.NumPages()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	var out []string
	for i := 0; i < n; i++ {
		p, err := r.Page(i)
		if err != nil {
			return nil, err
		}
		dicts, err := p.AnnotDicts()
		if err != nil || len(dicts) == 0 {
			continue
		}
		for _, d := range dicts {
			uri, _ := lrpdf.LinkAnnotationTarget(r, d)
			if uri == "" {
				continue
			}
			if _, ok := seen[uri]; ok {
				continue
			}
			seen[uri] = struct{}{}
			out = append(out, uri)
		}
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func extractPageLabelsPdfGo(data []byte) ([]string, error) {
	r, err := pdfGoOpenReader(data)
	if err != nil {
		return nil, err
	}
	return r.PageLabels()
}

func extractAttachmentNamesPdfGo(data []byte) ([]string, error) {
	r, err := pdfGoOpenReader(data)
	if err != nil {
		return nil, err
	}
	names, err := r.AttachmentNames()
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return []string{}, nil
	}
	seen := make(map[string]struct{}, len(names))
	uniq := make([]string, 0, len(names))
	for _, n := range names {
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		uniq = append(uniq, n)
	}
	sort.Strings(uniq)
	return uniq, nil
}

func extractXMPMetadataPdfGo(data []byte) (map[string]any, error) {
	r, err := pdfGoOpenReader(data)
	if err != nil {
		return nil, err
	}
	raw, err := r.XMPMetadata()
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	x := lrpdf.ParseXmpInformation(raw)
	if x == nil {
		return nil, nil
	}
	m := map[string]any{}
	if v := x.DcTitleDefault(); v != "" {
		m["dc_title"] = v
	}
	if desc := x.DcDescription(); len(desc) > 0 {
		m["dc_description"] = desc
	}
	if v := x.DcCreator(); len(v) > 0 {
		m["dc_creator"] = v
	}
	if v := x.DcSubject(); len(v) > 0 {
		m["dc_subject"] = v
	}
	if v := x.DcFormat(); v != "" {
		m["dc_format"] = v
	}
	if v := x.XmpCreatorTool(); v != "" {
		m["xmp_creator_tool"] = v
	}
	if v := x.PdfProducer(); v != "" {
		m["pdf_producer"] = v
	}
	if v := x.PdfKeywords(); v != "" {
		m["pdf_keywords"] = v
	}
	if v := x.PdfVersion(); v != "" {
		m["pdf_version"] = v
	}
	if v := x.XmpmmDocumentID(); v != "" {
		m["xmpmm_document_id"] = v
	}
	if v := x.XmpmmInstanceID(); v != "" {
		m["xmpmm_instance_id"] = v
	}
	if t, ok := x.XmpCreateDate(); ok {
		m["xmp_create_date"] = t.UTC().String()
	}
	if t, ok := x.XmpModifyDate(); ok {
		m["xmp_modify_date"] = t.UTC().String()
	}
	if len(m) == 0 {
		return nil, nil
	}
	return m, nil
}
