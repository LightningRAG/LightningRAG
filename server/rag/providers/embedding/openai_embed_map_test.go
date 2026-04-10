package embedding

import (
	"reflect"
	"testing"
)

func TestMapOpenAIEmbeddingBatch_ordersByIndex(t *testing.T) {
	data := []openaiEmbedDatum{
		{Embedding: []float32{2}, Index: 1},
		{Embedding: []float32{1}, Index: 0},
	}
	out, err := mapOpenAIEmbeddingBatch(2, data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out[0], []float32{1}) || !reflect.DeepEqual(out[1], []float32{2}) {
		t.Fatalf("got %#v", out)
	}
}

func TestMapOpenAIEmbeddingBatch_positionalFallbackWhenIndexCollapsed(t *testing.T) {
	data := []openaiEmbedDatum{
		{Embedding: []float32{1}, Index: 0},
		{Embedding: []float32{2}, Index: 0},
	}
	out, err := mapOpenAIEmbeddingBatch(2, data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out[0], []float32{1}) || !reflect.DeepEqual(out[1], []float32{2}) {
		t.Fatalf("got %#v", out)
	}
}

func TestMapOpenAIEmbeddingBatch_countMismatch(t *testing.T) {
	_, err := mapOpenAIEmbeddingBatch(2, []openaiEmbedDatum{{Embedding: []float32{1}, Index: 0}})
	if err == nil {
		t.Fatal("expected error")
	}
}
