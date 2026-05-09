package utils

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/yangbin1322/go-ddddocr/ddddocr"
	"llm_ocr_api_go/config"
)

var (
	ocrInstance *ddddocr.DdddOcr
	ocrOnce     sync.Once
	ocrMu       sync.Mutex
)

// InitDdddOcr 初始化 ddddocr 实例（使用 beta 模型）
func InitDdddOcr() error {
	var initErr error
	ocrOnce.Do(func() {
		modelDir := config.GetDdddOcrModelDir()

		onnxPath := filepath.Join(modelDir, "onnxruntime.dll")
		ddddocr.SetOnnxRuntimePath(onnxPath)

		opts := ddddocr.DefaultOptions()
		opts.Beta = true
		opts.ModelDir = modelDir

		ocrInstance, initErr = ddddocr.New(opts)
		if initErr != nil {
			initErr = fmt.Errorf("初始化 ddddocr 失败: %w", initErr)
		}
	})
	return initErr
}

// DdddOcrClassify 常规 OCR 识别（线程安全）
func DdddOcrClassify(imageData []byte) (string, error) {
	if err := InitDdddOcr(); err != nil {
		return "", err
	}

	ocrMu.Lock()
	defer ocrMu.Unlock()

	result, err := ocrInstance.Classification(imageData)
	if err != nil {
		return "", fmt.Errorf("ddddocr 识别失败: %w", err)
	}
	return result, nil
}

// DdddOcrClassifyWithRange 带字符范围限定的 OCR 识别（线程安全）
func DdddOcrClassifyWithRange(imageData []byte, rangeVal interface{}) (string, error) {
	if err := InitDdddOcr(); err != nil {
		return "", err
	}

	ocrMu.Lock()
	defer ocrMu.Unlock()

	ocrInstance.SetRanges(rangeVal)

	probResult, err := ocrInstance.ClassificationProbability(imageData)
	if err != nil {
		return "", fmt.Errorf("ddddocr 概率识别失败: %w", err)
	}

	ocrInstance.ClearRanges()

	var result string
	for _, prob := range probResult.Probability {
		if len(prob) == 0 {
			continue
		}
		maxIdx := 0
		maxVal := prob[0]
		for j, v := range prob {
			if v > maxVal {
				maxVal = v
				maxIdx = j
			}
		}
		if maxIdx < len(probResult.Charsets) {
			result += probResult.Charsets[maxIdx]
		}
	}

	return result, nil
}

// DecodeBase64Image 解码 base64 图像数据，失败时返回原始数据作为二进制
func DecodeBase64Image(data []byte) []byte {
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return data
	}
	return decoded
}

// GetNumCPU 获取 CPU 核心数
func GetNumCPU() int {
	return runtime.NumCPU()
}
