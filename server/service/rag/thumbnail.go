package rag

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const (
	thumbnailMaxWidth  = 200
	thumbnailMaxHeight = 200
	thumbnailQuality   = 75
)

// GenerateThumbnail 根据文件类型生成缩略图，返回 base64 编码的 JPEG 字符串
// 支持: jpg/jpeg/png/gif/webp → 缩放原图
// 其他格式 → 返回空字符串（前端可按文件类型展示图标）
func GenerateThumbnail(data []byte, fileType string) string {
	ft := strings.ToLower(fileType)
	switch ft {
	case "jpg", "jpeg", "png", "gif", "webp", "bmp", "tif", "tiff", "ico", "apng":
		return generateImageThumbnail(data)
	case "pdf":
		return generatePDFPlaceholder()
	default:
		return ""
	}
}

// GenerateThumbnailFromReader 从 Reader 生成缩略图
func GenerateThumbnailFromReader(r io.Reader, fileType string) string {
	data, err := io.ReadAll(r)
	if err != nil {
		return ""
	}
	return GenerateThumbnail(data, fileType)
}

// generateImageThumbnail 对图片进行等比缩放并编码为 base64 JPEG
func generateImageThumbnail(data []byte) string {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return ""
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return ""
	}

	newW, newH := fitDimensions(w, h, thumbnailMaxWidth, thumbnailMaxHeight)

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	resizeNearestNeighbor(dst, img)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: thumbnailQuality}); err != nil {
		return ""
	}

	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// generatePDFPlaceholder 为 PDF 生成一个简单的占位缩略图
func generatePDFPlaceholder() string {
	const w, h = 120, 160
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	bgColor := color.RGBA{R: 220, G: 53, B: 69, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	// "PDF" text approximation: draw a white rectangle in center
	inner := image.Rect(20, 60, 100, 100)
	draw.Draw(img, inner, &image.Uniform{C: white}, image.Point{}, draw.Src)

	// Smaller red rect inside to indicate text
	textArea := image.Rect(30, 70, 90, 75)
	draw.Draw(img, textArea, &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
	textArea2 := image.Rect(30, 80, 80, 85)
	draw.Draw(img, textArea2, &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
	textArea3 := image.Rect(30, 90, 70, 95)
	draw.Draw(img, textArea3, &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: thumbnailQuality}); err != nil {
		return ""
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// fitDimensions 计算等比缩放后的尺寸
func fitDimensions(srcW, srcH, maxW, maxH int) (int, int) {
	if srcW <= maxW && srcH <= maxH {
		return srcW, srcH
	}
	ratioW := float64(maxW) / float64(srcW)
	ratioH := float64(maxH) / float64(srcH)
	ratio := ratioW
	if ratioH < ratio {
		ratio = ratioH
	}
	newW := int(float64(srcW) * ratio)
	newH := int(float64(srcH) * ratio)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}
	return newW, newH
}

// resizeNearestNeighbor 使用最近邻插值缩放图像（无需额外依赖）
func resizeNearestNeighbor(dst *image.RGBA, src image.Image) {
	srcBounds := src.Bounds()
	dstBounds := dst.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	dstW := dstBounds.Dx()
	dstH := dstBounds.Dy()

	for y := 0; y < dstH; y++ {
		srcY := srcBounds.Min.Y + y*srcH/dstH
		for x := 0; x < dstW; x++ {
			srcX := srcBounds.Min.X + x*srcW/dstW
			dst.Set(dstBounds.Min.X+x, dstBounds.Min.Y+y, src.At(srcX, srcY))
		}
	}
}

// 确保 gif 解码器已注册
func init() {
	image.RegisterFormat("gif", "GIF8", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("png", "\x89PNG", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "\xff\xd8", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("bmp", "BM", bmp.Decode, bmp.DecodeConfig)
	image.RegisterFormat("tiff-le", "II", tiff.Decode, tiff.DecodeConfig)
	image.RegisterFormat("tiff-be", "MM", tiff.Decode, tiff.DecodeConfig)
}

// GetFileTypeIcon 返回文件类型对应的图标标识（前端可用）
func GetFileTypeIcon(fileType string) string {
	switch strings.ToLower(fileType) {
	case "pdf":
		return "pdf"
	case "doc", "docx":
		return "word"
	case "xls", "xlsx", "xlsb":
		return "excel"
	case "ppt", "pptx", "pages", "key":
		return "ppt"
	case "numbers":
		return "excel"
	case "txt", "md", "mdx", "json", "jsonl", "eml", "yaml", "yml", "svg", "tsv", "toml", "rtf", "odt", "ods", "odp", "odg", "epub", "ipynb", "xml":
		return "text"
	case "audio":
		return "audio"
	case "video":
		return "video"
	case "csv":
		return "csv"
	case "html", "htm":
		return "html"
	case "jpg", "jpeg", "png", "gif", "webp", "bmp", "tif", "tiff", "avif", "heic":
		return "image"
	default:
		return "file"
	}
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	if size <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	fSize := float64(size)
	i := 0
	for fSize >= 1024 && i < len(units)-1 {
		fSize /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d B", size)
	}
	return fmt.Sprintf("%.1f %s", fSize, units[i])
}
