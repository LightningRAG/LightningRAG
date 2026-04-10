# Document parsing vs. Ragflow (`references/ragflow`)

**Language:** English | **中文：** [DOCUMENT_PARSE_RAGFLOW_ALIGNMENT_zh.md](./DOCUMENT_PARSE_RAGFLOW_ALIGNMENT_zh.md)

This document describes how **LightningRAG** turns uploaded files into plain text during knowledge-base indexing, compared with **`references/ragflow`** (**deepdoc/parser**, **rag/app** naive / presentation / email, etc.), so we can deliberately align or diverge.

**Implementation entry points:** `parseDocumentContent` and `inferFileType` in `server/service/rag/knowledge_base.go`; Apple iWork preview PDF in `iwork_parser.go`; parsers under `server/service/rag/docparse` (PDF: `pdf.go` / `pdf_ledongthuc.go`, **pypdf** in `docparse/pypdfplain/`; JSON/JSONL, YAML, TOML, RTF, ODF, EPUB, IPynb, PPTX, EML with attachment callbacks, MIME helpers, `FileTypeFromMIME`, `RefineFileTypeByContent` for **gzip/bzip2** / ZIP / PDF magic, MSG, XLS, Excel extension fixes, HTML to text). `rag/pdf_parser.go` delegates to `docparse`.

---

## Type matrix (full detail)

The full per-format table (extensions, Ragflow vs. LightningRAG, alignment notes) is maintained in Chinese because it is very large and updated with the codebase:

**→ [DOCUMENT_PARSE_RAGFLOW_ALIGNMENT_zh.md § 类型总览](./DOCUMENT_PARSE_RAGFLOW_ALIGNMENT_zh.md#类型总览)**

---

## Code index

| Capability | Ragflow reference | LightningRAG |
| --- | --- | --- |
| Parser surface | `references/ragflow/deepdoc/parser/__init__.py` | `server/service/rag/docparse/*.go` and existing `*_parser.go` |
| Upload parse entry | `references/ragflow/api/db/services/file_service.py` (`parse` / `get_parser`) | `parseDocumentContent`, `inferFileType` |
| PPTX | `deepdoc/parser/ppt_parser.py` | `docparse/pptx.go` |
| JSON | `deepdoc/parser/json_parser.py` | `docparse/json_parser.go` |
| EML | `rag/app/email.py` | `docparse/eml_parser.go`, `eml_collector.go` |
| ODF | (open document) | `docparse/odf.go` |
| RTF | | `docparse/rtf.go` |
| TOML | (extension) | `docparse/toml.go` |
| MSG | setups / mail | `docparse/msg_parser.go` |
| XLS | spreadsheet | `docparse/xls.go` |
| YAML | (extension) | `docparse/yaml.go` |
| EPUB / ZIP typing | (extension) | `docparse/epub.go`, `zipkind.go` |
| Apple iWork preview PDF | `file_service` / iWork | `iwork_parser.go` |
| IPynb | (extension) | `docparse/ipynb.go` |
| MIME mapping | attachments, etc. | `docparse/mime_infer.go` |
| HTML → text | (shared) | `docparse/html_text.go` |
| PDF Plain / pypdf / links | `deepdoc/parser/pdf_parser.py` PlainParser; `rag/utils/file_utils.py` `extract_links_from_pdf` | `docparse/pdf.go`, `docparse/pdf_ledongthuc.go`, `docparse/pypdfplain/`; `rag/pdf_parser.go` delegates |

---

## Possible follow-ups

1. **PDF:** layout/table detection (DeepDOC-style) or a unified external parser service.  
2. **DOC / PPT:** optional Tika or hosted conversion.  
3. **EML:** parse attachment bytes by extension via `parseDocumentContent` (watch size and safety).  
4. **JSON:** structure-aware splitting closer to `RAGFlowJsonParser._json_split` before vector indexing.

---

*Keep this doc in sync with parser behavior; update the Chinese matrix when formats change.*
