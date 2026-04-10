package retriever

import (
	"testing"

	"github.com/LightningRAG/LightningRAG/server/rag/interfaces"
)

func TestWrapVectorEmptyRetry_nil(t *testing.T) {
	if WrapVectorEmptyRetry(nil) != nil {
		t.Fatal()
	}
}

func TestVectorEmptyRetryRetriever_RetrieverType(t *testing.T) {
	// 仅校验类型委托；无真实 store 时不调用 GetRelevantDocuments
	v := &VectorRetriever{reportType: interfaces.RetrieverTypeGlobal}
	w := WrapVectorEmptyRetry(v).(*VectorEmptyRetryRetriever)
	if w.RetrieverType() != interfaces.RetrieverTypeGlobal {
		t.Fatalf("%v", w.RetrieverType())
	}
}
