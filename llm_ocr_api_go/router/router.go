package router

import (
	"github.com/gin-gonic/gin"
	"llm_ocr_api_go/handler"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Any("/ping", handler.Ping)

	r.POST("/ruoyi/base64", handler.RuoyiBase64)

	return r
}
