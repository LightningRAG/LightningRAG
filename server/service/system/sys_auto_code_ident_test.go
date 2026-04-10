package system

import "testing"

func Test_validateAutoCodeIdent(t *testing.T) {
	if err := validateAutoCodeIdent("", "表名"); err == nil {
		t.Fatal("empty name should error")
	}
	if err := validateAutoCodeIdent("users", "表名"); err != nil {
		t.Fatal(err)
	}
	if err := validateAutoCodeIdent("users;drop", "表名"); err == nil {
		t.Fatal("injection-like name should error")
	}
	if err := validateAutoCodeIdent("1bad", "表名"); err == nil {
		t.Fatal("leading digit should error")
	}
}
