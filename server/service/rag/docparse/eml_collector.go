package docparse

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"
)

// MailAttachment MIME 邮件中的附件载荷（含内联 multipart 嵌套展平结果）。
type MailAttachment struct {
	FileName    string
	ContentType string
	Data        []byte
}

const (
	// MaxEMLAttachmentExtract 单封邮件最多尝试文本化的附件个数（对齐 Ragflow 附件再 chunk 的节制版）。
	MaxEMLAttachmentExtract = 8
	// MaxEMLAttachmentBytes 单个附件参与解析的最大字节数。
	MaxEMLAttachmentBytes = 4 << 20
)

func collectMIME(body []byte, mediaType string, params map[string]string) (plain, html string, attachments []MailAttachment, err error) {
	mt := strings.ToLower(strings.TrimSpace(mediaType))
	if strings.HasPrefix(mt, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return "", "", nil, fmt.Errorf("multipart 缺少 boundary")
		}
		return collectMultipart(body, boundary)
	}
	return "", "", nil, nil
}

func collectMultipart(body []byte, boundary string) (plain, html string, attachments []MailAttachment, err error) {
	mr := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return plain, html, attachments, err
		}
		subType, subParams, _ := mime.ParseMediaType(p.Header.Get("Content-Type"))
		if subType == "" {
			subType = "application/octet-stream"
		}
		raw, rerr := io.ReadAll(p)
		if rerr != nil {
			continue
		}
		raw, _ = decodeTransferBody(raw, p.Header)

		disposition, dispParams, _ := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
		fname := dispParams["filename"]
		if fname == "" {
			fname = subParams["name"]
		}

		if strings.HasPrefix(strings.ToLower(subType), "multipart/") {
			p2, h2, a2, e2 := collectMIME(raw, subType, subParams)
			if e2 != nil {
				continue
			}
			if p2 != "" && plain == "" {
				plain = p2
			}
			if h2 != "" && html == "" {
				html = h2
			}
			attachments = append(attachments, a2...)
			continue
		}

		isAttach := strings.EqualFold(disposition, "attachment")
		if !isAttach && fname != "" && !strings.HasPrefix(strings.ToLower(subType), "text/") {
			isAttach = true
		}
		if isAttach {
			if fname != "" || len(raw) > 0 {
				attachments = append(attachments, MailAttachment{
					FileName:    fname,
					ContentType: subType,
					Data:        raw,
				})
			}
			continue
		}

		switch {
		case strings.HasPrefix(strings.ToLower(subType), "text/plain"):
			if plain == "" {
				plain = string(raw)
			}
		case strings.HasPrefix(strings.ToLower(subType), "text/html"):
			if html == "" {
				html = string(raw)
			}
		}
	}
	return plain, html, attachments, nil
}
