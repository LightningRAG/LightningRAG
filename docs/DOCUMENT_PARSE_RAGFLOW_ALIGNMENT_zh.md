# 文档解析能力与 Ragflow（references/ragflow）对齐说明

**English:** [DOCUMENT_PARSE_RAGFLOW_ALIGNMENT.md](./DOCUMENT_PARSE_RAGFLOW_ALIGNMENT.md)

本文记录 **LightningRAG** 知识库上传/索引阶段的「文档 → 纯文本」解析能力与本地参考仓库 `references/ragflow` 中 **deepdoc/parser**、**rag/app**（naive / presentation / email 等）的对应关系，便于后续继续补齐或刻意保持差异。

实现入口：`server/service/rag/knowledge_base.go` 的 `parseDocumentContent`、`inferFileType`；`iwork_parser.go` 抽取 Apple iWork 包内预览 PDF；解析子包：`server/service/rag/docparse`（**PDF** 见 `pdf.go` / `pdf_ledongthuc.go`，**pypdf** 实现在 `docparse/pypdfplain/`；以及 JSON/JSONL、YAML、TOML、RTF、ODT/ODS/ODP、EPUB、IPynb、PPTX、EML 含附件回调与 MIME 辅助、`FileTypeFromMIME`、`RefineFileTypeByContent` **gzip / bzip2** / ZIP / PDF 魔数纠偏、MSG、XLS、Excel 扩展名纠偏、HTML 转文本）。`rag/pdf_parser.go` 仅转调 `docparse`。

---

## 类型总览

| 类别（对照 Ragflow `rag/flow/parser/parser.py` setups） | 扩展名 / 说明 | Ragflow 主要实现 | LightningRAG 实现 | 对齐状态 |
| --- | --- | --- | --- | --- |
| PDF | `.pdf` | `deepdoc/parser/pdf_parser.py`（Plain/DeepDOC/VLM 等） | **默认优先 pypdf**：`docparse/pypdfplain/pypdf_plain_extract.py` 调 **`references/pypdf`**（子命令 **`full`** / **`meta`** / **`pages`** / **`docinfo`** / **`links`** / **`pagelabels`** / **`xmp`** / **`attachmentnames`**，与 Ragflow PlainParser / `extract_links_from_pdf` 等对齐）；**失败或正文为空**再 **`ledongthuc/pdf`**；ledongthuc 路径下若有 **书签** 则在正文后追加 `--- PDF 大纲 / 书签 ---`。环境变量：**`LIGHTNINGRAG_PDF_ENGINE`** — 未设为上述默认；**`pypdf`** 仅 pypdf；**`ledongthuc`** / **`thuc`** 仅 ledongthuc；**`auto`** 为先 ledongthuc 再 pypdf（旧行为）。另含 **`LIGHTNINGRAG_PDF_APPEND_*`**、**`LIGHTNINGRAG_PYPDF_SRC`**、**`LIGHTNINGRAG_REPO_ROOT`**、**`LIGHTNINGRAG_PYTHON`**、**`LIGHTNINGRAG_PDF_PASSWORD`**、**`LIGHTNINGRAG_PYPDF_EXTRACTION_MODE`**、页范围、**`LIGHTNINGRAG_PYPDF_STRICT`**、**`LIGHTNINGRAG_PYPDF_ROOT_RECOVERY_LIMIT`** 等；**页范围不作用于 ledongthuc**。失败则 OCR（知识库可关） | **部分对齐**：无 DeepDOC 版面；扫描件与 Ragflow VLM 类似依赖模型 |
| Word | `.docx` | `deepdoc/parser/docx_parser.py` + `rag/app/naive.py` | `word/document.xml` + `header`/`footer` + **`footnotes`/`endnotes`/`comments`** 抽段落文本 | **部分对齐**：仍无表格单元格专项结构、无 Docling |
| Word（宏/模板 OOXML） | `.docm` `.dotx` `.dotm` | 与 docx 同 OPC | 扩展名映射为 `docx`，走同一 `ParseDOCXContent`；MIME 含 `macroenabled` / `wordprocessingml.template` | **扩展** |
| Word（旧） | `.doc` | `rag/app/one.py` 等依赖 **Tika** | 明确报错，提示转 `.docx` | **未对齐**：不引入 Tika |
| WPS（OOXML） | `.wps` `.et` `.dps` | `api/utils/file_utils.py` 将 `.wps` 与文档类归并 | 扩展名分别记为 `wps`/`et`/`dps`；**ZIP 内为 OOXML** 时 `RefineFileTypeByContent` 纠偏为 **docx / xlsx / pptx** 再走同一解析器；传统 OLE 等 **明确报错** 并提示另存为 Office 格式；MIME **`application/vnd.wps-office.wps`** 等 | **部分对齐**：与 Ragflow「按扩展进文档类」一致，不引入 WPS 私有二进制解析 |
| Windows 帮助 | `.hlp` | 与 doc 同列于 `file_utils` | 明确报错（非文本化） | **未对齐** |
| 表格 | `.xlsx` | `deepdoc/parser/excel_parser.py` | `excelize` 遍历工作表与行 | **部分对齐**：输出为制表符分隔文本，非 Ragflow 表格 HTML/JSON 管线 |
| 表格（宏/模板 OOXML） | `.xlsm` `.xltx` `.xltm` | 同 spreadsheet | 映射为 `xlsx` 解析；MIME `macroenabled` / `spreadsheetml.template` | **扩展** |
| 表格（二进制） | `.xlsb` | 部分环境依赖专用库 | 明确报错提示转 `.xlsx`；MIME `sheet.binary.macroenabled` → `xlsb` | **未对齐**：不引入 xlsb 专用解析 |
| 表格（旧） | `.xls` | 同 spreadsheet | `extrame/xls` 读 OLE，UTF-8 / GBK 回退 | **基本对齐**：与 xlsx 相同「展平为文本」策略；扩展名与 zip/OLE 魔数不一致时由 `NormalizeExcelFileType` 纠偏 |
| CSV | `.csv` | 随 spreadsheet / naive | `encoding/csv` 读入行；首行 **逗号/分号** 多者作为分隔符（欧陆 `;`）；MIME `text/comma-separated-values` | **基本对齐**：纯文本化 |
| TSV | `.tsv` | 常作表格交换 | `encoding/csv` + `Comma='\t'` | **扩展** |
| TOML | `.toml` | 配置/数据常见 | `go-toml` → 缩进 JSON 再切片 | **扩展** |
| RTF | `.rtf` | Word 等导出 | `docparse` 简易剥离控制字 + 跳过 `{\*…}` | **部分对齐**：无完整 RTF 引擎 |
| ODF 字处理 | `.odt` | LibreOffice / ODF | `content.xml` 抽 `text:p` / `text:h` | **基本对齐**：轻量 XML 遍历 |
| ODF 表格 | `.ods` | 同左 | `content.xml` 抽 `table:table-row` / `table-cell` | **基本对齐** |
| ODF 演示 | `.odp` | 同左 | 与 ODT 相同 `text:p` / `text:h` 遍历 | **基本对齐** |
| ODF 绘图 | `.odg` | LibreOffice Draw | `content.xml` 同 ODP 类文本遍历；ZIP `mimetype` / manifest 含 `opendocument.graphics`；MIME `vnd.oasis.opendocument.graphics` | **扩展** |
| EPUB | `.epub` | 电子书常见 | ZIP 内 `.xhtml`/`.html` 经 `HTMLToPlainText` 拼接 | **扩展** |
| Jupyter | `.ipynb` | 数据科学 | 抽取 `markdown`/`code`/`raw` 单元 `source` | **扩展** |
| XML | `.xml` / `.rss` / `.atom` / `.xsl` / `.xslt` / `.xlf` / `.xliff` / `.plist`（文本 plist） / **`.kml` `.gpx` `.wsdl` `.xsd` `.drawio` `.dio`** / **`.rng` `.sch` `.fo`**（Relax NG XML、Schematron、XSL-FO） | 配置/数据 / 交换格式 | 按 UTF-8 文本读取（不做 XSD 校验）；MIME `rss+xml` / `atom+xml` / XLIFF / **`gpx+xml` / `vnd.google-earth.kml+xml` / `wsdl+xml` / `xslt+xml`** 等 → `xml`；**不**将 Relax NG **compact**（RNC）等非 XML 语法误映射为 `xml` | **扩展** |
| 类型纠偏 | 错扩展名 / 无后缀 | — | `RefineFileTypeByContent`：**gzip / bzip2 魔数**、`%PDF`、OOXML/ODF/EPUB ZIP 布局；**`.pdf` 实为 ZIP 办公文档** 时覆盖；**`.wps`/`.et`/`.dps` 实为 OOXML** 时覆盖为 docx/xlsx/pptx | **扩展** |
| Gzip 单文件 | `.gz` / `.tgz` / **`.svgz`** | 部分管线经解压 | 解压一层；**`.svgz` → 内层 `.svg`** 再走 SVG 文本解析；**拒绝 tar/tar.gz / 嵌套 gzip**；魔数纠偏含 `.svgz` 文件名 | **扩展** |
| Bzip2 单文件 | `.bz2` / `.tbz2` | 部分管线经解压 | `compress/bzip2` 解压一层；`stripBzip2Filename`；**拒绝 tar.bz2 / .tbz2**；**拒绝嵌套 bzip2**；MIME `application/x-bzip2` 等 | **扩展** |
| Apple iWork | `.pages` `.numbers` `.key` | `file_service` 将 `.pages` 与演示类一并列出 | ZIP 包内 **`QuickLook/Preview.pdf`**（或 `*/preview.pdf` 回退）→ 按 **PDF** 再解析（含 OCR 回退）；MIME `vnd.apple.pages` / `numbers` / `keynote` | **部分对齐**：不解析 iwa 正文，依赖系统预览 PDF |
| 字幕 / 日历 | `.vtt` / `.srt` / **`.sbv`** / `.ics` | 常作文本类 | 按 UTF-8 文本解析；MIME：`text/vtt`、`application/x-subrip`、`text/calendar` 等映射为 `txt` | **扩展** |
| 附件 MIME | EML 附件 | — | `ShouldPreferMIMEOverExt`：`txt`/`xml` 与 MIME 冲突时采 MIME；再 `RefineFileTypeByContent`；补充 `application/msword`、`vnd.ms-excel` 等旧 Office MIME | **部分对齐** |
| HTML | `.html` / `.htm` / **`.smi`**（SMIL，按 HTML 管线尽力文本化） | `deepdoc/parser/html_parser.py` | `ParseHTMLContent` 正则去 script/style；EML/EPUB 用 `HTMLToPlainText` **不遍历** script/style/noscript/iframe/template 等子树；MIME **`application/smil`**、**`application/smil+xml`**、**`text/x-sami`** → `html` | **部分对齐**：无深度 DOM 版面；内嵌 HTML 清洗思路对齐 naive |
| 文本 / Markdown | `.txt` / `.md` / `.markdown` / `.mkd` | `txt_parser` / `markdown_parser` | `ParseTextContent`（**UTF-8 + UTF-16 LE/BE BOM**）；MIME `text/x-markdown` → `md` | **基本对齐**：未做 Ragflow Markdown 表格拆分等专项 |
| 源码（naive 文本类） | Ragflow：`py/js/java/c/cpp/h/php/go/ts/sh/cs/kt/sql`；LightningRAG 另含 `mjs/cjs/tsx/jsx/swift/rs/lua/dart/vue/svelte/scala/gradle/kts/graphql/gql/proto/ps1/psm1/vim`、**`.bat` `.cmd`**、**`.glsl` `.vert` `.frag` `.geom` `.comp` `.hlsl` `.metal`**（着色器）等 | `naive.chunk` 中 `TxtParser` | 显式 `txt` + `ParseTextContent` | **对齐 naive 并扩展**；`.ts` 与视频 TS 流同名时仍按文本 |
| 源码（MIME） | — | 附件类型 | `text/x-python`、`text/x-java-source`、`text/x-go`、`text/x-rust`、`text/x-ruby`、`text/x-shellscript`、`application/graphql`、`text/css`、`text/javascript`/`application/javascript`、`application/typescript`、`text/x-sql`/`application/sql`、**`application/x-httpd-php`** / **`application/x-php`** / **`application/x-httpd-php-source`**、**`application/x-sh`** / **`text/x-sh`**、**`application/x-bat`** 等 → `txt`；**不**将泛用可执行 MIME（如 `application/x-msdos-program`）映射为文本 | **扩展** |
| Web 归档 / XHTML | `.mht` `.mhtml` `.xhtml` `.xht` | 常作 HTML 处理 | 映射为 `html`，`ParseHTMLContent` | **扩展** |
| JSON Lines（别名） | `.ldjson`（Ragflow naive） | `JsonParser` | 与 `.jsonl` / `.ndjson` 同为 `jsonl` | **对齐** |
| 字幕 | `.ass` / `.ssa` | 常见文本轨道 | 按 `txt` 解析 | **扩展** |
| MDX | `.mdx` | setups 中 `text&markdown` 含 `mdx` | 按 UTF-8 文本读取；PageIndex 与 `.md` 同逻辑建树 | **基本对齐** |
| JSON / JSONL | `.json` / `.jsonl` / `.ndjson` | `deepdoc/parser/json_parser.py`（分块 + 结构保持） | `docparse.ParseJSONText`：单 JSON 缩进序列化；JSONL 按行解析再拼接；**`.har`**、**`.geojson` `.topojson`** → `json`；MIME `hal+json` / `problem+json` / **`geo+json` / `schema+json` / `manifest+json`** / **`model/gltf+json`** / **`model/vnd.gltf+json`** 等 → `json` | **部分对齐**：不在解析阶段做 Ragflow 式子树分块，交给上游 `ChunkDocument` |
| JSON Lines（MIME） | — | 附件 Content-Type | `application/x-ndjson`、`application/jsonlines`、`application/jsonl` → `jsonl` | **扩展** |
| 配置 / 日志 / 文稿 | `.ini` `.cfg` `.conf` `.properties` `.env` `.log` `.rst` `.tex` `.adoc` `.asciidoc` `.org` | 常作文本 naive | `inferFileType` → `txt`，`ParseTextContent` | **扩展** |
| 工程忽略 / 注释 JSON / 模板 | `.gitignore` `.gitattributes` `.dockerignore` `.jsonc` `.j2` `.jinja` `.jinja2` | — | → `txt`（`jsonc` 不做 JSONC 语法清洗，整段当文本） | **扩展** |
| 前端模板 / 其他文本 | `.pug` `.jade` `.liquid` `.twig` `.erb` `.eex` `.coffee` `.cue` `.url` `.mjml` | — | → `txt` | **扩展** |
| 构建 / 证书 / 歌词 / 服务端模板 | `.nix` `.bzl` `.bazel` `.bazelrc` `.gn` `.gni` `.cmake.in`；`.pem` `.crt` `.cer`；`.lrc`；`.ejs` `.hbs` `.mustache`；MIME `text/x-handlebars-template` 等 | — | → `txt` | **扩展** |
| Web 样式 / Cursor | `.css` `.less` `.scss` `.sass` `.styl` `.mdc` | — | → `txt` | **扩展** |
| Shell | `.fish` `.zsh` | — | → `txt` | **扩展** |
| 学术 / 语言扩展 | `.bib` LaTeX `.sty` `.cls`；Fortran `.f90`–`.f08` `.for` `.f`；Perl `.pm` `.pl`；Julia `.jl`；R `.r`；Nim `.nim`；Elixir `.ex` `.exs`；Erlang `.erl` `.hrl`；Clojure `.clj` `.cljs` `.edn`；Zig `.zig`；Verilog/VHDL `.v` `.vh` `.sv` `.vhd` `.vhdl`；`.cmake` `.mk` `.patch` `.diff`；C++ `.hpp` `.hh` `.hxx` | — | 均 → `txt` | **扩展** |
| Terraform / HCL | `.tf` `.tfvars` `.hcl` | — | → `txt` | **扩展** |
| BibTeX / Org / TeX（MIME） | — | — | `application/x-bibtex` `text/x-bibtex` `text/x-org` `application/x-tex` `text/x-tex` → `txt` | **扩展** |
| 演示文稿 | `.pptx` | `deepdoc/parser/ppt_parser.py`（python-pptx） | `ppt/slides/slide*.xml` + `ppt/notesSlides/notesSlide*.xml` 抽 `a:t`（演讲者备注附于对应页） | **部分对齐**：无形状几何排序/表格结构化 |
| 演示（宏/模板 OOXML） | `.pptm` `.potx` `.potm` | 同 slides | 映射为 `pptx`，走 `docparse.ParsePPTXText`；MIME `presentation.macroenabled` / `presentationml.template` | **扩展** |
| 演示文稿（旧） | `.ppt` | presentation + python-pptx / Tika fallback | 明确报错，提示转 `.pptx` | **未对齐**：不引入 Tika |
| 邮件 | `.eml` | `rag/app/email.py`（头 + plain/html + 附件再走 naive） | `ParseEMLTextWithAttachments`：`emlAttachmentParseChain` 在单附件 ≤4MB、每封最多 8 个附件内调用 `parseDocumentContent`；嵌套 `.eml` 深度 ≤2 | **部分对齐**：与 Ragflow 一样「能解析则入库」；超大/过多附件仍仅列名 |
| 邮件（Outlook） | `.msg` | setups 含 `msg` | `docparse.ParseMSGText`：`willthrom/outlook-msg-parser` + 临时文件 | **部分对齐**：与 eml 相同附件策略；库对复杂 msg 可能不完整 |
| 结构化文本 | `.yaml` / `.yml` | 常作配置/数据与 naive 文本类同 | `yaml.v3` → 缩进 JSON 文本再切片 | **扩展**：便于与 JSON 共用下游分块 |
| 矢量 | `.svg` | 非 Ragflow image 默认后缀 | 按 UTF-8 文本读取（XML 源码入库）；MIME **`image/svg+xml`** | **扩展**：便于检索嵌入文字 |
| 音频 | 多后缀（含 `aiff`/`au`/`midi`/`ape` 等）及 **`.opus` `.spx` `.alac` `.wv`**（WavPack） | `rag/app/audio.py`；`file_utils` 含 `alac`/`opus`/`vorbis` 等 | 统一走 Speech2Text；`audio/*` MIME 亦归为 `audio` | **策略对齐** |
| 视频 | `mp4`/`avi`/`mkv`/`mov`/`webm`/`wmv` 及 **`m4v`/`mpeg`/`mpg`/`3gp`/`flv`** / **`.ogv` `.f4v` `.asf`** / **`.rm` `.rmvb` `.mpe` `.mpa`** 等 | setups 含 video + 模型；`file_utils` VISUAL 含部分容器 | 明确报错提示（不入库文本） | **未实现**：与 Ragflow「需配置视频模型」一致；扩展名与常见容器对齐 |
| 图片 | jpg/png/gif/webp/**bmp/tiff** / **`.jfif` `.jpe`**→`jpg` / **`.ico` `.icon`** / **`.apng`** / **`.avif` `.heic` `.heif`**（`heif`→`heic`）等 | OCR / VLM；`file_utils` VISUAL 含 `icon`/`ico`/`apng` 等 | 原图字节送 OCR/CV；MIME **`image/vnd.microsoft.icon`**、**`image/x-icon`**、**`image/apng`**；**缩略图**仍依赖 `image.Decode`（ICO/APNG/AVIF/HEIC 等无解码器时可能为空） | **策略对齐** |

---

## 代码索引

| 能力 | Ragflow 参考路径 | LightningRAG |
| --- | --- | --- |
| 解析器导出 | `references/ragflow/deepdoc/parser/__init__.py` | `server/service/rag/docparse/*.go` + 既有 `*_parser.go` |
| 上传侧解析入口 | `references/ragflow/api/db/services/file_service.py`（`parse` / `get_parser`） | `parseDocumentContent`、`inferFileType` |
| PPTX | `deepdoc/parser/ppt_parser.py` | `docparse/pptx.go` |
| JSON | `deepdoc/parser/json_parser.py` | `docparse/json_parser.go` |
| EML | `rag/app/email.py` | `docparse/eml_parser.go` + `eml_collector.go` |
| ODF | （开放文档） | `docparse/odf.go` |
| RTF | （富文本） | `docparse/rtf.go` |
| TOML | （扩展） | `docparse/toml.go` |
| MSG | setups / 邮件类 | `docparse/msg_parser.go` |
| XLS | spreadsheet | `docparse/xls.go` |
| YAML | （扩展） | `docparse/yaml.go` |
| EPUB / ZIP 类型 | （扩展） | `docparse/epub.go`、`zipkind.go` |
| Apple iWork 预览 PDF | `file_service` / iWork | `iwork_parser.go` |
| IPynb | （扩展） | `docparse/ipynb.go` |
| MIME 映射 | 附件等 | `docparse/mime_infer.go` |
| HTML→文本 | （共用） | `docparse/html_text.go` |
| PDF Plain / pypdf / 链接 | `deepdoc/parser/pdf_parser.py` PlainParser；`rag/utils/file_utils.py` `extract_links_from_pdf` | `docparse/pdf.go`、`docparse/pdf_ledongthuc.go`、`docparse/pypdfplain/`；`rag/pdf_parser.go` 转调 |

---

## 后续可加强方向（可选）

1. **PDF**：引入版面/表格检测（对齐 DeepDOC）或统一走外部解析服务。  
2. **DOC / PPT**：接入可选 Tika 或托管转换服务。  
3. **EML**：对附件字节按扩展名复用 `parseDocumentContent`（需注意安全与体积）。  
4. **JSON**：实现接近 `RAGFlowJsonParser._json_split` 的结构感知分块，再交给向量索引。

---

*文档版本与仓库实现同步维护；修改解析行为时请更新本表。*
