package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yangbin1322/go-ddddocr/ddddocr"
)

func main() {
	modelDir := "models"

	onnxPath := filepath.Join(modelDir, "onnxruntime.dll")
	ddddocr.SetOnnxRuntimePath(onnxPath)

	opts := ddddocr.DefaultOptions()
	opts.Beta = true
	opts.ModelDir = modelDir

	ocr, err := ddddocr.New(opts)
	if err != nil {
		fmt.Printf("初始化 ddddocr 失败: %v\n", err)
		os.Exit(1)
	}
	defer ocr.Close()

	imageDir := `..\yzm_num`

	entries, err := os.ReadDir(imageDir)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		os.Exit(1)
	}

	var imageNames []string
	for _, entry := range entries {
		name := entry.Name()
		lower := strings.ToLower(name)
		if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
			imageNames = append(imageNames, name)
		}
	}

	sort.Strings(imageNames)

	for _, imageName := range imageNames {
		imagePath := filepath.Join(imageDir, imageName)
		data, err := os.ReadFile(imagePath)
		if err != nil {
			fmt.Printf("%s => 读取失败: %v\n", imageName, err)
			continue
		}

		result, err := ocr.Classification(data)
		if err != nil {
			fmt.Printf("%s => 识别失败: %v\n", imageName, err)
			continue
		}

		fmt.Printf("%s => %s\n", imageName, result)
	}
}
