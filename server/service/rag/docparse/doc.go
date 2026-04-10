// Package docparse 实现与 references/ragflow 中部分 deepdoc/parser、rag/app 能力对应的文本抽取入口
// （含 PDF：默认优先 pypdf 子包 pypdfplain + references/pypdf，失败再 ledongthuc/pdf；以及 JSON/JSONL、YAML、TOML、
// RTF、ODT/ODS/ODP、EPUB、IPynb、PPTX、XLS、EML/附件链、MIME 映射、ZIP/PDF 类型纠偏、MSG、Excel 魔数纠偏等），
// 供知识库 parseDocumentContent 使用；完整差异见 docs/DOCUMENT_PARSE_RAGFLOW_ALIGNMENT.md（中文全文见 DOCUMENT_PARSE_RAGFLOW_ALIGNMENT_zh.md）。
package docparse
