package retriever

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

func TestWrapKeywordEmptyRetry_nil(t *testing.T) {
	if WrapKeywordEmptyRetry(nil) != nil {
		t.Fatal()
	}
}

func TestKeywordEmptyRetryRetriever_RetrieverType(t *testing.T) {
	k := &KeywordRetriever{reportType: interfaces.RetrieverTypeLocal}
	w := WrapKeywordEmptyRetry(k).(*KeywordEmptyRetryRetriever)
	if w.RetrieverType() != interfaces.RetrieverTypeLocal {
		t.Fatalf("%v", w.RetrieverType())
	}
}
