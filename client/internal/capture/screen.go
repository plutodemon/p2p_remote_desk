package capture

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"os"
	"strconv"

	"github.com/kbinani/screenshot"
	"github.com/nfnt/resize"
)

// ScreenCapture 屏幕捕获结构体
type ScreenCapture struct {
	DisplayIndex int    // 显示器索引
	Quality      string // 图像质量
}

// NewScreenCapture 创建新的屏幕捕获实例
func NewScreenCapture() *ScreenCapture {
	return &ScreenCapture{
		DisplayIndex: 0,
		Quality:      "100",
	}
}

// CaptureScreen 捕获屏幕
func (sc *ScreenCapture) CaptureScreen() (image.Image, error) {
	// 获取屏幕图像
	bounds := screenshot.GetDisplayBounds(sc.DisplayIndex)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, fmt.Errorf("捕获屏幕失败: %v", err)
	}

	// 调整图像质量
	quality, err := strconv.Atoi(sc.Quality)
	if err != nil {
		quality = 100
	}

	// 如果质量小于100，进行压缩处理
	if quality < 100 {
		// 计算新的尺寸
		newWidth := uint(float64(bounds.Dx()) * float64(quality) / 100)
		newHeight := uint(float64(bounds.Dy()) * float64(quality) / 100)

		// 调整图像大小
		resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

		// 创建新的RGBA图像
		finalImg := image.NewRGBA(resizedImg.Bounds())
		draw.Draw(finalImg, finalImg.Bounds(), resizedImg, resizedImg.Bounds().Min, draw.Src)
		return finalImg, nil
	}

	return img, nil
}

// GetDisplayCount 获取显示器数量
func (sc *ScreenCapture) GetDisplayCount() int {
	return screenshot.NumActiveDisplays()
}

// SaveToFile 保存图像到文件
func (sc *ScreenCapture) SaveToFile(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	quality, err := strconv.Atoi(sc.Quality)
	if err != nil {
		quality = 100
	}

	opt := jpeg.Options{
		Quality: quality,
	}

	if err := jpeg.Encode(file, img, &opt); err != nil {
		return fmt.Errorf("编码图像失败: %v", err)
	}

	return nil
}

// Cleanup 清理资源
func (sc *ScreenCapture) Cleanup() {
	// 目前没有需要清理的资源
}
