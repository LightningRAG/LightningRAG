package utils

import (
	"testing"
)

func TestGetJSONKeys(t *testing.T) {
	jsonStr := `
	{
		"Name": "test",
		"TableName": "test",
		"TemplateID": "test",
		"TemplateInfo": "test",
		"Limit": 0
}`
	keys, err := GetJSONKeys(jsonStr)
	if err != nil {
		t.Fatalf("GetJSONKeys: %v", err)
	}
	if len(keys) != 5 {
		t.Fatalf("GetJSONKeys: want 5 keys, got %d: %v", len(keys), keys)
	}
	want := []string{"Name", "TableName", "TemplateID", "TemplateInfo", "Limit"}
	for i, w := range want {
		if keys[i] != w {
			t.Errorf("keys[%d]: want %q, got %q", i, w, keys[i])
		}
	}
}
