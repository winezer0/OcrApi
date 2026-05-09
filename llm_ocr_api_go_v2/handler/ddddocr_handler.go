package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"llm_ocr_api_go/utils"

	"github.com/gin-gonic/gin"
)

var rangeDescriptions = map[int]string{
	0: "纯整数 (0-9)",
	1: "纯小写英文 (a-z)",
	2: "纯大写英文 (A-Z)",
	3: "小写英文+大写英文 (a-z A-Z)",
	4: "小写英文+整数 (a-z 0-9)",
	5: "大写英文+整数 (A-Z 0-9)",
	6: "小写英文+大写英文+整数 (a-z A-Z 0-9)",
	7: "默认字符库 (a-z A-Z 0-9)",
}

// DdddOcrBase64 ddddocr 基础 OCR 识别接口
func DdddOcrBase64(c *gin.Context) {
	startTime := time.Now()

	imgData, err := io.ReadAll(c.Request.Body)
	if err != nil || len(imgData) == 0 {
		c.String(http.StatusOK, "")
		return
	}

	fmt.Printf("[ddddocr] Base64 Img Data Length: %d\n", len(imgData))

	imgBin := utils.DecodeBase64Image(imgData)

	result, err := utils.DdddOcrClassify(imgBin)
	if err != nil {
		fmt.Printf("[ddddocr] OCR error: %v\n", err)
		c.String(http.StatusOK, "")
		return
	}

	durationMs := time.Since(startTime).Milliseconds()
	fmt.Printf("[ddddocr] OCR result: %s | Processed in %d ms\n", result, durationMs)

	c.String(http.StatusOK, result)
}

// DdddOcrBase64WithParam ddddocr 带字符范围限定的 OCR 识别接口
func DdddOcrBase64WithParam(c *gin.Context) {
	startTime := time.Now()

	param := c.Param("param")

	imgData, err := io.ReadAll(c.Request.Body)
	if err != nil || len(imgData) == 0 {
		c.String(http.StatusOK, "")
		return
	}

	fmt.Printf("[ddddocr] Base64 Img Data Length: %d\n", len(imgData))

	imgBin := utils.DecodeBase64Image(imgData)

	var rangeVal interface{}
	if len(param) == 1 && param[0] >= '0' && param[0] <= '9' {
		intVal, _ := strconv.Atoi(param)
		if _, ok := rangeDescriptions[intVal]; !ok {
			intVal = 7
		}
		rangeVal = intVal
		fmt.Printf("[ddddocr] 限定结果类型为内置范围: %d -> %s\n", intVal, rangeDescriptions[intVal])
	} else {
		rangeVal = param
		fmt.Printf("[ddddocr] 限定结果类型为指定范围: %s\n", param)
	}

	result, err := utils.DdddOcrClassifyWithRange(imgBin, rangeVal)
	if err != nil {
		fmt.Printf("[ddddocr] OCR error: %v\n", err)
		c.String(http.StatusOK, "")
		return
	}

	durationMs := time.Since(startTime).Milliseconds()
	fmt.Printf("[ddddocr] OCR result: %s | Processed in %d ms\n", result, durationMs)

	c.String(http.StatusOK, result)
}
