#!/usr/bin/env python3
# 位置：server/service/rag/docparse/pypdfplain/pypdf_plain_extract.py（由 Go pypdfplain 包 go:embed 写入临时目录执行）。
# 对齐 references/ragflow/deepdoc/parser/pdf_parser.py PlainParser：
#   PdfReader、page.extract_text()、outline DFS（/Title）。
# 子命令（argv[2]）：
#   full    — 全文 + 书签块（与 Ragflow 一致）；可选 argv[3]=from_page argv[4]=to_page（半开切片，默认全页）
#   meta    — 仅 stdout JSON：{"page_count": N}
#   pages   — 仅 stdout JSON：{"pages": [...]}，与 full 相同页范围参数
#   docinfo     — 仅 stdout JSON：pdf_header、metadata 等（对齐 pypdf._doc_common）
#   links       — 仅 stdout JSON：{"links":[...]}，对齐 ragflow/rag/utils/file_utils.extract_links_from_pdf（/Annots → /A → /URI）
#   pagelabels  — 仅 stdout JSON：{"page_labels":[...]}，对齐 pypdf PdfDocCommon.page_labels
#   xmp         — 仅 stdout JSON：{"xmp": {...}|null}，对齐 pypdf PdfReader.xmp_metadata 常用字段
#   attachmentnames — 仅 stdout JSON：{"attachment_names":[...]}，对齐 pypdf 嵌入附件名（不输出正文）
# 环境变量（由 Go 传入）：
#   LIGHTNINGRAG_PDF_PASSWORD     — 解密密码（可选）
#   LIGHTNINGRAG_PYPDF_EXTRACTION_MODE — plain（默认）| layout
#   LIGHTNINGRAG_PDF_PREPEND_DOCINFO   — 1/true 时在 full 模式正文前附加「PDF 文档信息」文本块
#   LIGHTNINGRAG_PDF_APPEND_LINKS      — 1/true 时在 full 模式正文与书签之间附加 URI 列表块（便于检索链接）
#   LIGHTNINGRAG_PDF_APPEND_XMP        — 1/true 时在 full 模式附加 XMP 摘要文本块（在链接块之后、书签之前）
#   LIGHTNINGRAG_PDF_APPEND_ATTACHMENT_NAMES — 1/true 时附加嵌入附件文件名列表（不读取附件字节）
#   LIGHTNINGRAG_PYPDF_STRICT          — 1/true 时 PdfReader(strict=True)
#   LIGHTNINGRAG_PYPDF_ROOT_RECOVERY_LIMIT — PdfReader root_object_recovery_limit（整数）
# 依赖：argv[1] 为 pypdf 包根目录（references/pypdf）。
from __future__ import annotations

import json
import os
import sys
from datetime import datetime
from io import BytesIO


def _extraction_kwargs():
    mode = (os.environ.get("LIGHTNINGRAG_PYPDF_EXTRACTION_MODE") or "plain").strip().lower()
    if mode == "layout":
        return {"extraction_mode": "layout"}
    return {"extraction_mode": "plain"}


def _normalize_outline_root(raw):
    if raw is None:
        return []
    if isinstance(raw, dict):
        return [raw]
    if isinstance(raw, list):
        return raw
    return [raw]


def _dfs_outline(arr, depth: int, out: list[tuple[str, int]]) -> None:
    """对齐 Ragflow PlainParser 的 dfs(outlines, 0)；兼容 dict 类 Destination/OutlineItem。"""
    if arr is None:
        return
    if isinstance(arr, dict):
        t = arr.get("/Title")
        if t is not None:
            out.append((str(t), depth))
        return
    if not isinstance(arr, (list, tuple)):
        _dfs_outline(arr, depth + 1, out)
        return
    for a in arr:
        if isinstance(a, dict):
            t = a.get("/Title")
            if t is not None:
                out.append((str(t), depth))
            continue
        _dfs_outline(a, depth + 1, out)


def _parse_page_range(argc: int, argv: list[str], n_pages: int) -> tuple[int, int]:
    """返回 [start, end) 与 Python slice 一致。"""
    start, end = 0, n_pages
    if argc >= 4:
        try:
            start = max(0, int(argv[3]))
        except ValueError:
            start = 0
    if argc >= 5:
        try:
            end = min(n_pages, int(argv[4]))
        except ValueError:
            end = n_pages
    if start > n_pages:
        start = n_pages
    if end < start:
        end = start
    return start, end


def _reader_init_kwargs() -> dict:
    kw: dict = {}
    pwd = os.environ.get("LIGHTNINGRAG_PDF_PASSWORD")
    if pwd is not None and pwd != "":
        kw["password"] = pwd
    lim = (os.environ.get("LIGHTNINGRAG_PYPDF_ROOT_RECOVERY_LIMIT") or "").strip()
    if lim:
        try:
            kw["root_object_recovery_limit"] = int(lim)
        except ValueError:
            pass
    strict = (os.environ.get("LIGHTNINGRAG_PYPDF_STRICT") or "").strip().lower()
    kw["strict"] = strict in ("1", "true", "yes", "on")
    return kw


def _open_reader(pdf_reader_cls, data: bytes):
    return pdf_reader_cls(BytesIO(data), **_reader_init_kwargs())


def _dt_iso(d) -> str | None:
    if d is None:
        return None
    if hasattr(d, "isoformat"):
        try:
            return d.isoformat()
        except Exception:
            return str(d)
    return str(d)


def _docinfo_payload(reader) -> dict:
    """对齐 pypdf PdfReader.metadata / pdf_header。"""
    out: dict = {
        "page_count": len(reader.pages),
        "pdf_header": getattr(reader, "pdf_header", "") or "",
    }
    try:
        md = reader.metadata
    except Exception:
        return out
    if md is None:
        return out
    try:
        out["title"] = md.title
        out["author"] = md.author
        out["subject"] = md.subject
        out["creator"] = md.creator
        out["producer"] = md.producer
        out["keywords"] = md.keywords
        out["creation_date"] = _dt_iso(md.creation_date)
        out["modification_date"] = _dt_iso(md.modification_date)
        cr = md.creation_date_raw
        mr = md.modification_date_raw
        out["creation_date_raw"] = None if cr is None else str(cr)
        out["modification_date_raw"] = None if mr is None else str(mr)
    except Exception:
        pass
    return out


def _extract_uri_links(reader) -> list[str]:
    """对齐 references/ragflow/rag/utils/file_utils.py extract_links_from_pdf。"""
    from pypdf.generic import IndirectObject

    links: set[str] = set()
    for page in reader.pages:
        annots = page.get("/Annots")
        if not annots or isinstance(annots, IndirectObject):
            continue
        try:
            for annot_ref in annots:
                try:
                    obj = annot_ref.get_object()
                except Exception:
                    continue
                a = obj.get("/A")
                if not a:
                    continue
                if isinstance(a, IndirectObject):
                    try:
                        a = a.get_object()
                    except Exception:
                        continue
                uri = a.get("/URI") if hasattr(a, "get") else None
                if uri is not None:
                    links.add(str(uri))
        except Exception:
            continue
    return sorted(links)


def _jsonify_xmp_value(v):
    if v is None:
        return None
    if isinstance(v, datetime):
        return v.isoformat()
    if isinstance(v, list):
        return [_jsonify_xmp_value(i) for i in v]
    if isinstance(v, dict):
        return {str(k): _jsonify_xmp_value(val) for k, val in v.items()}
    return v


def _xmp_summary_payload(reader):
    """对齐 pypdf PdfReader.xmp_metadata 的常用 Dublin Core / PDF / XMP 字段。"""
    try:
        x = reader.xmp_metadata
    except Exception:
        return None
    if x is None:
        return None
    payload: dict = {}
    for name in (
        "dc_title",
        "dc_creator",
        "dc_description",
        "dc_subject",
        "dc_date",
        "pdf_keywords",
        "pdf_producer",
        "pdf_pdfversion",
        "xmp_create_date",
        "xmp_modify_date",
        "xmp_metadata_date",
        "xmp_creator_tool",
    ):
        try:
            v = getattr(x, name, None)
        except Exception:
            continue
        if v is None:
            continue
        payload[name] = _jsonify_xmp_value(v)
    return payload or None


def _append_xmp_block(reader) -> str | None:
    if (os.environ.get("LIGHTNINGRAG_PDF_APPEND_XMP") or "").strip().lower() not in (
        "1",
        "true",
        "yes",
        "on",
    ):
        return None
    xp = _xmp_summary_payload(reader)
    if not xp:
        return None
    lines = ["--- PDF XMP 摘要（pypdf xmp_metadata）---"]
    for k in sorted(xp.keys()):
        lines.append(f"{k}: {json.dumps(xp[k], ensure_ascii=False, default=str)}")
    lines.append("---")
    return "\n".join(lines)


def _list_attachment_names(reader) -> list[str]:
    try:
        raw = reader._list_attachments()
    except Exception:
        return []
    return sorted(set(raw))


def _append_attachment_names_block(reader) -> str | None:
    if (os.environ.get("LIGHTNINGRAG_PDF_APPEND_ATTACHMENT_NAMES") or "").strip().lower() not in (
        "1",
        "true",
        "yes",
        "on",
    ):
        return None
    names = _list_attachment_names(reader)
    if not names:
        return None
    lines = ["--- PDF 嵌入附件文件名（pypdf，无正文）---"]
    lines.extend(names)
    lines.append("---")
    return "\n".join(lines)


def _append_links_block(reader) -> str | None:
    if (os.environ.get("LIGHTNINGRAG_PDF_APPEND_LINKS") or "").strip().lower() not in (
        "1",
        "true",
        "yes",
        "on",
    ):
        return None
    urls = _extract_uri_links(reader)
    if not urls:
        return None
    lines = ["--- PDF 超链接（URI，对齐 Ragflow extract_links_from_pdf）---"]
    lines.extend(urls)
    lines.append("---")
    return "\n".join(lines)


def _prepend_docinfo_block(reader) -> str | None:
    if (os.environ.get("LIGHTNINGRAG_PDF_PREPEND_DOCINFO") or "").strip().lower() not in (
        "1",
        "true",
        "yes",
        "on",
    ):
        return None
    info = _docinfo_payload(reader)
    lines = ["--- PDF 文档信息（pypdf metadata）---"]
    for k in (
        "pdf_header",
        "page_count",
        "title",
        "author",
        "subject",
        "creator",
        "producer",
        "keywords",
        "creation_date",
        "modification_date",
    ):
        if k not in info or info[k] is None or info[k] == "":
            continue
        lines.append(f"{k}: {info[k]}")
    lines.append("---")
    return "\n".join(lines)


def _emit_docinfo(reader) -> None:
    raw = _docinfo_payload(reader)
    # 去掉 null，便于 Go encoding/json 解到 string 字段（JSON null 不能写入 Go string）
    clean = {k: v for k, v in raw.items() if v is not None}
    sys.stdout.buffer.write(json.dumps(clean, ensure_ascii=False, default=str).encode("utf-8"))


def _emit_full(reader, start: int, end: int) -> None:
    ext_kw = _extraction_kwargs()
    lines: list[str] = []
    for page in reader.pages[start:end]:
        try:
            txt = page.extract_text(**ext_kw)
        except TypeError:
            txt = page.extract_text()
        except Exception:
            txt = ""
        if txt:
            lines.extend([t for t in txt.split("\n")])

    body = "\n".join(lines).strip()
    outlines_flat: list[tuple[str, int]] = []
    try:
        _dfs_outline(_normalize_outline_root(reader.outline), 0, outlines_flat)
    except Exception:
        pass

    parts: list[str] = []
    pre = _prepend_docinfo_block(reader)
    if pre:
        parts.append(pre)
    if body:
        parts.append(body)
    link_block = _append_links_block(reader)
    if link_block:
        parts.append(link_block)
    xmp_block = _append_xmp_block(reader)
    if xmp_block:
        parts.append(xmp_block)
    att_block = _append_attachment_names_block(reader)
    if att_block:
        parts.append(att_block)
    if outlines_flat:
        parts.append("--- PDF 大纲 / 书签 ---")
        for title, depth in outlines_flat:
            parts.append("  " * depth + title)

    out = "\n\n".join(parts) if parts else ""
    sys.stdout.buffer.write(out.encode("utf-8"))


def _emit_meta(reader) -> None:
    n = len(reader.pages)
    sys.stdout.buffer.write(json.dumps({"page_count": n}, ensure_ascii=False).encode("utf-8"))


def _emit_pages(reader, start: int, end: int) -> None:
    ext_kw = _extraction_kwargs()
    pages: list[str] = []
    for page in reader.pages[start:end]:
        try:
            txt = page.extract_text(**ext_kw)
        except TypeError:
            txt = page.extract_text()
        except Exception:
            txt = ""
        pages.append(txt or "")
    sys.stdout.buffer.write(json.dumps({"pages": pages}, ensure_ascii=False).encode("utf-8"))


def _emit_links(reader) -> None:
    urls = _extract_uri_links(reader)
    sys.stdout.buffer.write(json.dumps({"links": urls}, ensure_ascii=False).encode("utf-8"))


def _emit_pagelabels(reader) -> None:
    try:
        labs = reader.page_labels
    except Exception:
        labs = []
    sys.stdout.buffer.write(json.dumps({"page_labels": labs}, ensure_ascii=False, default=str).encode("utf-8"))


def _emit_xmp(reader) -> None:
    xp = _xmp_summary_payload(reader)
    sys.stdout.buffer.write(json.dumps({"xmp": xp}, ensure_ascii=False, default=str).encode("utf-8"))


def _emit_attachmentnames(reader) -> None:
    names = _list_attachment_names(reader)
    sys.stdout.buffer.write(json.dumps({"attachment_names": names}, ensure_ascii=False).encode("utf-8"))


def main() -> int:
    if len(sys.argv) < 3:
        print(
            "usage: plain_extract.py PYPDF_ROOT {full|meta|pages|docinfo|links|pagelabels|xmp|attachmentnames} [from_page [to_page]] < pdf",
            file=sys.stderr,
        )
        return 2
    sys.path.insert(0, sys.argv[1])
    try:
        from pypdf import PdfReader as PdfReaderCls
    except ImportError as e:
        print(f"import pypdf failed: {e}", file=sys.stderr)
        return 1

    data = sys.stdin.buffer.read()
    if not data:
        print("empty stdin", file=sys.stderr)
        return 1

    sub = sys.argv[2].strip().lower()
    try:
        reader = _open_reader(PdfReaderCls, data)
    except Exception as e:
        print(f"PdfReader: {e}", file=sys.stderr)
        return 1

    n = len(reader.pages)
    start, end = _parse_page_range(len(sys.argv), sys.argv, n)

    try:
        if sub == "meta":
            _emit_meta(reader)
        elif sub == "docinfo":
            _emit_docinfo(reader)
        elif sub == "links":
            _emit_links(reader)
        elif sub == "pagelabels":
            _emit_pagelabels(reader)
        elif sub == "xmp":
            _emit_xmp(reader)
        elif sub == "attachmentnames":
            _emit_attachmentnames(reader)
        elif sub == "pages":
            _emit_pages(reader, start, end)
        elif sub == "full":
            _emit_full(reader, start, end)
        else:
            print(f"unknown mode: {sub}", file=sys.stderr)
            return 2
    except Exception as e:
        print(f"extract: {e}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
