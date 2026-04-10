package system

import (
	"fmt"
	"regexp"
)

// autoCodeSQLIdent 代码生成场景下的安全子集：典型标识符，用于无法参数化、必须拼进 SQL 的片段。
var autoCodeSQLIdent = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateAutoCodeIdent(name, what string) error {
	if name == "" {
		return fmt.Errorf("%s 不能为空", what)
	}
	if !autoCodeSQLIdent.MatchString(name) {
		return fmt.Errorf("非法%s: %q", what, name)
	}
	return nil
}
