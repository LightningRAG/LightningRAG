package retriever

import (
	"context"
	"fmt"
	"sort"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
	"github.com/LightningRAG/LightningRAG/server/rag/lightragconst"
	"github.com/LightningRAG/LightningRAG/server/rag/pageindex"
	"github.com/LightningRAG/LightningRAG/server/rag/schema"
)

// DocTree 文档及其 PageIndex 树
type DocTree struct {
	DocID   uint
	DocName string
	Tree    []pageindex.TreeNode
	// NodeToRagChunkID 与索引进度一致：StructureToList 顺序对应 rag_chunks.chunk_index，用于 Ragflow 式 TOC 补全（node_id -> rag_chunk 主键）
	NodeToRagChunkID map[string]uint
}

// PageIndexRetriever 基于 PageIndex 树状推理的检索器
type PageIndexRetriever struct {
	docTrees []DocTree
	llm      interfaces.LLM
	numDocs  int
}

// NewPageIndexRetriever 创建 PageIndex 检索器
func NewPageIndexRetriever(docTrees []DocTree, llm interfaces.LLM, numDocs int) *PageIndexRetriever {
	if numDocs <= 0 {
		numDocs = lightragconst.DefaultChunkTopK
	}
	return &PageIndexRetriever{
		docTrees: docTrees,
		llm:      llm,
		numDocs:  numDocs,
	}
}

// GetRelevantDocuments 通过 PageIndex 检索文档片段。
// 主路径对齐 references/ragflow：将树压平为 TOC（level/title），用 LLM 逐项打分（toc_relevance），再映射回节点正文；
// 若打分失败或结果为空，回退为原 TreeSearch（node_list）推理路径。
func (r *PageIndexRetriever) GetRelevantDocuments(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	n := numDocs
	if n <= 0 {
		n = r.numDocs
	}
	if r.llm == nil {
		return nil, fmt.Errorf("PageIndex 检索需要 LLM 进行树推理")
	}
	if len(r.docTrees) == 0 {
		return nil, nil
	}

	perDoc := (n + len(r.docTrees) - 1) / len(r.docTrees)
	if perDoc < 1 {
		perDoc = 1
	}

	var all []schema.Document
	for _, dt := range r.docTrees {
		if len(dt.Tree) == 0 {
			continue
		}
		parentByChild := pageindex.BuildParentNormIDByChild(dt.Tree)
		enrichChunkLineage := func(nodeIDNorm string, meta map[string]any) {
			if meta == nil || dt.NodeToRagChunkID == nil {
				return
			}
			if rid, ok := dt.NodeToRagChunkID[nodeIDNorm]; ok && rid != 0 {
				meta["rag_chunk_id"] = rid
			}
			if pnorm, ok := parentByChild[nodeIDNorm]; ok && pnorm != "" {
				if prid, ok2 := dt.NodeToRagChunkID[pnorm]; ok2 && prid != 0 {
					meta["parent_rag_chunk_id"] = prid
				}
			}
		}
		entries := pageindex.CapTOCEntries(pageindex.FlattenTreeToTOC(dt.Tree))
		var fromDoc []schema.Document
		if len(entries) > 0 {
			scored, err := pageindex.TocRelevanceSearch(ctx, r.llm, query, entries)
			if err == nil && len(scored) > 0 {
				nodeMap := pageindex.CreateNodeMapping(dt.Tree)
				for _, sn := range scored {
					nn := nodeMap[sn.NodeID]
					if nn == nil {
						continue
					}
					text := pageindex.GetNodeText(nn)
					if text == "" {
						continue
					}
					normID := pageindex.NormalizeNodeID(sn.NodeID)
					md := map[string]any{
						"document_id":    dt.DocID,
						"doc_name":       dt.DocName,
						"node_id":        normID,
						"title":          nn.Title,
						"pageindex_mode": "toc_relevance",
					}
					enrichChunkLineage(normID, md)
					fromDoc = append(fromDoc, schema.Document{
						PageContent: text,
						Metadata:    md,
						Score:       float32(sn.Score),
					})
				}
			}
		}
		if len(fromDoc) == 0 {
			result, err := pageindex.TreeSearch(ctx, r.llm, query, dt.Tree, perDoc)
			if err != nil {
				continue
			}
			nodeMap := pageindex.CreateNodeMapping(dt.Tree)
			for _, nodeID := range result.NodeList {
				nodeID = pageindex.NormalizeNodeID(nodeID)
				if nn := nodeMap[nodeID]; nn != nil {
					text := pageindex.GetNodeText(nn)
					if text != "" {
						md := map[string]any{
							"document_id":    dt.DocID,
							"doc_name":       dt.DocName,
							"node_id":        nodeID,
							"title":          nn.Title,
							"pageindex_mode": "tree_search",
						}
						enrichChunkLineage(nodeID, md)
						fromDoc = append(fromDoc, schema.Document{
							PageContent: text,
							Metadata:    md,
							Score:       1.0,
						})
					}
				}
			}
		}
		all = append(all, fromDoc...)
	}

	all = MergeRetrievalByChildren(ctx, all)

	sort.SliceStable(all, func(i, j int) bool {
		si, sj := all[i].Score, all[j].Score
		if si != sj {
			return si > sj
		}
		li, lj := len(all[i].PageContent), len(all[j].PageContent)
		if li != lj {
			return li > lj
		}
		return all[i].PageContent < all[j].PageContent
	})
	if len(all) > n {
		all = all[:n]
	}
	return all, nil
}

// RetrieverType 返回检索类型
func (r *PageIndexRetriever) RetrieverType() interfaces.RetrieverType {
	return interfaces.RetrieverTypePageIndex
}
