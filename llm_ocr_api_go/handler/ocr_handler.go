package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"llm_ocr_api_go/config"
	"llm_ocr_api_go/utils"

	"github.com/gin-gonic/gin"
)

type OcrRequest struct {
	ImageBase64 string `json:"imageBase64"`
}

type OcrResponse struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Result string `json:"result"`
	Raw    string `json:"raw"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Ping 心跳检测接口，允许任何请求方法
func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

// RuoyiBase64 若依验证码识别接口，接收 base64 图片并返回计算结果
func RuoyiBase64(c *gin.Context) {
	const prompt = "你是验证码识别与计算助手。图片内容固定为数字 操作符 数字 = ?格式。请识别并完成计算。只输出最终结果数字，不要输出任何解释、空格、标点或其他字符。"

	startTime := time.Now()

	var imageBase64 string

	if c.ContentType() == "application/json" {
		var req OcrRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.ImageBase64 != "" {
			imageBase64 = utils.NormalizeBase64(req.ImageBase64)
		}
	}

	if imageBase64 == "" {
		body, err := io.ReadAll(c.Request.Body)
		if err == nil {
			imageBase64 = utils.NormalizeBase64(string(body))
		}
	}

	if imageBase64 == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "image base64 is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.GetDashScopeTimeout())
	defer cancel()

	rawText, err := utils.CallDashScopeOCR(ctx, prompt, imageBase64)
	if err != nil {
		fmt.Printf("[-] OCR Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	result := utils.ExtractResultNumber(rawText)

	durationMs := time.Since(startTime).Milliseconds()
	fmt.Printf("[*] Request done: %d ms | result: %s\n", durationMs, result)

	c.JSON(http.StatusOK, OcrResponse{
		Code:   200,
		Msg:    "success",
		Result: result,
		Raw:    rawText,
	})
}
