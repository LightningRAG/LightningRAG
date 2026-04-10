package docparse

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"
)

// AttachmentExtractFunc 由上层注入（通常为 parseDocumentContent），用于把附件转为可索引文本。
// 返回 ("", nil) 表示跳过该附件（仍可能输出 [attachment] 文件名行）。
type AttachmentExtractFunc func(ctx context.Context, filename, contentType string, data []byte) (string, error)

// ParseEMLText 等价于 ParseEMLTextWithAttachments(ctx, data, nil)，附件仅列文件名。
func ParseEMLText(data []byte) (string, error) {
	return ParseEMLTextWithAttachments(context.Background(), data, nil)
}

// ParseEMLTextWithAttachments 解析 EML；若 onAttach 非 nil，则在限制内尝试抽取附件正文（对齐 Ragflow naive 对附件再 chunk）。
func ParseEMLTextWithAttachments(ctx context.Context, data []byte, onAttach AttachmentExtractFunc) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("邮件数据为空")
	}
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("解析 EML 失败: %w", err)
	}
	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return "", fmt.Errorf("读取邮件正文失败: %w", err)
	}
	return buildEMLText(ctx, msg, body, onAttach)
}

func buildEMLText(ctx context.Context, msg *mail.Message, body []byte, onAttach AttachmentExtractFunc) (string, error) {
	var sb strings.Builder
	for k, vals := range msg.Header {
		kn := textproto.CanonicalMIMEHeaderKey(k)
		switch kn {
		case "From", "To", "Cc", "Bcc", "Subject", "Date", "Reply-To", "Message-Id":
			for _, v := range vals {
				if dec, derr := (&mime.WordDecoder{}).DecodeHeader(v); derr == nil && dec != "" {
					v = dec
				}
				sb.WriteString(kn)
				sb.WriteString(": ")
				sb.WriteString(v)
				sb.WriteString("\n")
			}
		}
	}

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil || mediaType == "" {
		mediaType = "text/plain"
	}

	var plain, htmlPart string
	var attachments []MailAttachment
	if strings.HasPrefix(strings.ToLower(mediaType), "multipart/") {
		plain, htmlPart, attachments, err = collectMIME(body, mediaType, params)
		if err != nil {
			return "", err
		}
	} else {
		raw, _ := decodeTransferBody(body, mailHeaderAdapter(msg.Header))
		switch {
		case strings.HasPrefix(strings.ToLower(mediaType), "text/html"):
			htmlPart = string(raw)
		default:
			plain = string(raw)
		}
	}

	if plain != "" {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(strings.TrimSpace(plain))
	}
	if htmlPart != "" {
		st := HTMLToPlainText(htmlPart)
		if st != "" {
			if sb.Len() > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString(st)
		}
	}

	if err := appendEMLAttachmentSections(ctx, &sb, attachments, onAttach); err != nil {
		return "", err
	}

	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("EML 未提取到文本")
	}
	return out, nil
}

func appendEMLAttachmentSections(ctx context.Context, sb *strings.Builder, attachments []MailAttachment, onAttach AttachmentExtractFunc) error {
	if len(attachments) == 0 {
		return nil
	}
	extracted := 0
	for _, a := range attachments {
		fname := strings.TrimSpace(a.FileName)
		label := fname
		if label == "" {
			label = "(unnamed)"
		}
		if onAttach == nil || len(a.Data) == 0 || len(a.Data) > MaxEMLAttachmentBytes || extracted >= MaxEMLAttachmentExtract {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString("[attachment] ")
			sb.WriteString(label)
			continue
		}
		text, err := onAttach(ctx, fname, a.ContentType, a.Data)
		if err != nil || strings.TrimSpace(text) == "" {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString("[attachment] ")
			sb.WriteString(label)
			continue
		}
		extracted++
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("--- attachment: ")
		sb.WriteString(label)
		sb.WriteString(" ---\n")
		sb.WriteString(strings.TrimSpace(text))
	}
	return nil
}

type mimeHeader interface {
	Get(key string) string
}

type mailHeaderAdapter mail.Header

func (a mailHeaderAdapter) Get(key string) string {
	return mail.Header(a).Get(key)
}

func decodeTransferBody(b []byte, h mimeHeader) ([]byte, error) {
	cte := strings.ToLower(strings.TrimSpace(h.Get("Content-Transfer-Encoding")))
	switch cte {
	case "base64":
		s := strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
				return -1
			}
			return r
		}, string(b))
		out, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return b, nil
		}
		return out, nil
	case "quoted-printable":
		r := quotedprintable.NewReader(bytes.NewReader(b))
		out, err := io.ReadAll(r)
		if err != nil {
			return b, nil
		}
		return out, nil
	default:
		return b, nil
	}
}
