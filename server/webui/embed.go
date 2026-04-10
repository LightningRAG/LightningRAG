package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:webdist
var distFS embed.FS

// Dist 为前端构建产物根（webdist 目录内容来自 web/dist 同步）。
func Dist() embed.FS {
	return distFS
}

// HasDist 判断是否已经同步过可发布的静态资源（存在 index.html）。
func HasDist() bool {
	_, err := fs.Stat(distFS, "webdist/index.html")
	return err == nil
}
