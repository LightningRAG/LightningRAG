package rag

import "testing"

func TestKgMergeSummarizeCacheKeyStable(t *testing.T) {
	k1 := kgMergeSummarizeCacheKey("实体", "Foo", []string{"a", "b"})
	k2 := kgMergeSummarizeCacheKey("实体", "Foo", []string{"a", "b"})
	if k1 != k2 {
		t.Fatalf("same input should same key")
	}
	k3 := kgMergeSummarizeCacheKey("实体", "Foo", []string{"b", "a"})
	if k1 == k3 {
		t.Fatal("order should matter")
	}
}
