package progress

import (
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
)

// NewProcessBar 初始化进度显示条
func NewProcessBar(total int64, desc string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions64(
		total,                                     // 任务总数
		progressbar.OptionSetDescription(desc),    // 进度条描述信息
		progressbar.OptionSetWidth(15),            // 进度条宽度
		progressbar.OptionThrottle(1*time.Second), // 更新频率限制，避免刷新过快
		progressbar.OptionShowCount(),             // 显示已完成/总任务数
		progressbar.OptionShowIts(),               // 显示每秒完成任务数
		progressbar.OptionSetPredictTime(true),    // 启用自动ETA预测
		progressbar.OptionOnCompletion(func() { // 完成时的回调函数
			fmt.Println() // 完成后换行，保持输出整洁
		}),
		progressbar.OptionSpinnerType(14), // 设置进度条动画样式
		progressbar.OptionFullWidth(),     // 启用全屏宽度
		progressbar.OptionSetTheme(progressbar.Theme{ // 自定义进度条样式
			Saucer:        "=", // 已完成部分的填充字符
			SaucerHead:    ">", // 进度条头部字符
			SaucerPadding: " ", // 未完成部分的填充字符
			BarStart:      "[", // 进度条起始字符
			BarEnd:        "]", // 进度条结束字符
		}),
	)
	return bar
}

// NewSpinner 初始化不确定总数的进度条（Spinner）
func NewSpinner(desc string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions64(
		-1, // -1 表示不确定总数
		progressbar.OptionSetDescription(desc),
		progressbar.OptionSetWidth(15),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	return bar
}

// NewByteProgressBar 初始化字节进度显示条
func NewByteProgressBar(total int64, desc string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions64(
		total,                                     // 任务总数
		progressbar.OptionSetDescription(desc),    // 进度条描述信息
		progressbar.OptionSetWidth(15),            // 进度条宽度
		progressbar.OptionThrottle(1*time.Second), // 更新频率限制，避免刷新过快
		progressbar.OptionShowBytes(true),         // 显示已完成/总字节数
		progressbar.OptionShowIts(),               // 显示每秒完成字节数
		progressbar.OptionSetPredictTime(true),    // 启用自动ETA预测
		progressbar.OptionOnCompletion(func() { // 完成时的回调函数
			fmt.Println() // 完成后换行，保持输出整洁
		}),
		progressbar.OptionSpinnerType(14), // 设置进度条动画样式
		progressbar.OptionFullWidth(),     // 启用全屏宽度
		progressbar.OptionSetTheme(progressbar.Theme{ // 自定义进度条样式
			Saucer:        "=", // 已完成部分的填充字符
			SaucerHead:    ">", // 进度条头部字符
			SaucerPadding: " ", // 未完成部分的填充字符
			BarStart:      "[", // 进度条起始字符
			BarEnd:        "]", // 进度条结束字符
		}),
	)
	return bar
}
