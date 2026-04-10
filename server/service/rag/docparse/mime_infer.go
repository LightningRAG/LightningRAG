package docparse

import (
	"mime"
	"strings"
)

// FileTypeFromMIME 将常见 Content-Type 映射为本项目 fileType；无法识别时返回 ""。
func FileTypeFromMIME(contentType string) string {
	ct := strings.TrimSpace(contentType)
	if ct == "" {
		return ""
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return ""
	}
	mt = strings.ToLower(mt)
	switch mt {
	case "application/pdf":
		return "pdf"
	case "text/html", "application/xhtml+xml", "application/smil", "application/smil+xml",
		"text/x-sami":
		return "html"
	case "text/plain", "text/x-python", "text/x-java-source", "text/x-c", "text/x-c++",
		"text/x-go", "text/x-rust", "text/x-ruby", "text/x-shellscript", "text/x-script.python",
		"text/css", "text/javascript", "application/javascript", "application/ecmascript", "text/ecmascript",
		"application/typescript", "text/x-typescript", "text/x-sql", "application/sql",
		"text/x-handlebars-template", "text/x-mustache-template",
		"application/x-httpd-php", "application/x-php", "application/x-httpd-php-source",
		"application/x-sh", "text/x-sh", "application/x-bat":
		return "txt"
	case "application/graphql":
		return "txt"
	case "application/x-bibtex", "text/x-bibtex":
		return "txt"
	case "text/x-org":
		return "txt"
	case "application/x-tex", "text/x-tex":
		return "txt"
	case "text/vtt":
		return "txt"
	case "application/x-subrip", "text/srt":
		return "txt"
	case "text/calendar":
		return "txt"
	case "text/csv", "text/comma-separated-values":
		return "csv"
	case "text/tab-separated-values", "text/tsv":
		return "tsv"
	case "application/json", "text/json", "application/ld+json",
		"application/hal+json", "application/problem+json", "application/vnd.api+json",
		"application/jsonc", "text/jsonc", "application/har+json",
		"application/geo+json", "application/vnd.geo+json",
		"application/schema+json", "application/manifest+json",
		"model/gltf+json", "model/vnd.gltf+json":
		return "json"
	case "application/x-ndjson", "application/jsonlines", "application/jsonl":
		return "jsonl"
	case "text/markdown", "text/x-markdown":
		return "md"
	case "application/xml", "text/xml", "application/rss+xml", "application/atom+xml",
		"application/xliff+xml", "application/x-xliff+xml",
		"application/gpx+xml", "application/vnd.google-earth.kml+xml",
		"application/wsdl+xml", "application/x-wsdl+xml", "application/xslt+xml":
		return "xml"
	case "application/yaml", "application/x-yaml":
		return "yaml"
	case "application/toml":
		return "toml"
	case "application/rtf", "text/rtf":
		return "rtf"
	case "application/x-ipynb+json", "application/vnd.jupyter+json", "application/jupyter+json":
		return "ipynb"
	case "application/epub+zip":
		return "epub"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-word.document.macroenabled.12",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.template":
		return "docx"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-excel.sheet.macroenabled.12",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.template":
		return "xlsx"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/vnd.ms-powerpoint.presentation.macroenabled.12",
		"application/vnd.openxmlformats-officedocument.presentationml.template":
		return "pptx"
	case "application/vnd.ms-excel.sheet.binary.macroenabled.12":
		return "xlsb"
	case "application/vnd.oasis.opendocument.text":
		return "odt"
	case "application/vnd.oasis.opendocument.spreadsheet":
		return "ods"
	case "application/vnd.oasis.opendocument.presentation":
		return "odp"
	case "application/vnd.oasis.opendocument.graphics":
		return "odg"
	case "message/rfc822":
		return "eml"
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/bmp", "image/x-ms-bmp":
		return "bmp"
	case "image/tiff", "image/x-tiff":
		return "tiff"
	case "image/avif":
		return "avif"
	case "image/heic", "image/heif":
		return "heic"
	case "image/svg+xml":
		return "svg"
	case "image/x-icon", "image/vnd.microsoft.icon":
		return "ico"
	case "image/apng":
		return "apng"
	case "application/kswps", "application/wps", "application/vnd.wps-office.wps", "application/x-wps":
		return "wps"
	case "application/kset", "application/vnd.wps-office.et", "application/x-et":
		return "et"
	case "application/ksdps", "application/vnd.wps-office.dps", "application/x-dps":
		return "dps"
	case "application/vnd.apple.pages":
		return "pages"
	case "application/vnd.apple.numbers":
		return "numbers"
	case "application/vnd.apple.keynote":
		return "key"
	case "application/msword":
		return "doc"
	case "application/vnd.ms-excel":
		return "xls"
	case "application/vnd.ms-powerpoint":
		return "ppt"
	case "application/vnd.ms-outlook":
		return "msg"
	case "application/gzip", "application/x-gzip":
		return "gz"
	case "application/x-bzip2", "application/x-bzip":
		return "bz2"
	}
	if strings.HasPrefix(mt, "audio/") {
		return "audio"
	}
	if strings.HasPrefix(mt, "video/") {
		return "video"
	}
	return ""
}

// ShouldPreferMIMEOverExt 在邮件附件等场景下，若扩展名推断过泛（如 .dat→txt）而 MIME 更具体，则采用 MIME。
func ShouldPreferMIMEOverExt(inferred, fromMIME string) bool {
	if fromMIME == "" || inferred == fromMIME {
		return false
	}
	switch inferred {
	case "txt", "xml":
		return true
	default:
		return false
	}
}
