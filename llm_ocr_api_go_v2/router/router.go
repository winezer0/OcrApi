package router

import (
	"github.com/gin-gonic/gin"
	"llm_ocr_api_go/handler"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Any("/ping", handler.Ping)

	r.POST("/ruoyi/base64", handler.RuoyiBase64)

	r.POST("/ddddocr/base64", handler.DdddOcrBase64)
	r.POST("/ddddocr/base64/:param", handler.DdddOcrBase64WithParam)

	return r
}
