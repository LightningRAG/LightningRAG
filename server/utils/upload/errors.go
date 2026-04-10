package upload

import "errors"

// ErrFileNotFound 存储中已无对应对象或本地文件（删除数据库记录时仍可继续）。
var ErrFileNotFound = errors.New("file not found")
