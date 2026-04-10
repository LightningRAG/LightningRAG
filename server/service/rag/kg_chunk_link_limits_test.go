package rag

import (
	"reflect"
	"testing"

	"github.com/LightningRAG/LightningRAG/server/global"
)

func TestKgChunkIDsToDropForLimit(t *testing.T) {
	ids := []uint{10, 20, 30, 40}
	// KEEP: retain lowest ids → drop 30,40
	if d := kgChunkIDsToDropForLimit(ids, 2, false); !reflect.DeepEqual(d, []uint{30, 40}) {
		t.Fatalf("keep: %#v", d)
	}
	// FIFO: retain highest ids → drop 10,20
	if d := kgChunkIDsToDropForLimit(ids, 2, true); !reflect.DeepEqual(d, []uint{10, 20}) {
		t.Fatalf("fifo: %#v", d)
	}
	if kgChunkIDsToDropForLimit(ids, 10, true) != nil {
		t.Fatal("no drop when under limit")
	}
}

func TestKgChunkLinkUseFIFO(t *testing.T) {
	old := global.LRAG_CONFIG.Rag.KgChunkLinkLimitMethod
	t.Cleanup(func() { global.LRAG_CONFIG.Rag.KgChunkLinkLimitMethod = old })
	global.LRAG_CONFIG.Rag.KgChunkLinkLimitMethod = ""
	if !kgChunkLinkUseFIFO() {
		t.Fatal("empty -> fifo")
	}
	global.LRAG_CONFIG.Rag.KgChunkLinkLimitMethod = "keep"
	if kgChunkLinkUseFIFO() {
		t.Fatal("keep")
	}
}
