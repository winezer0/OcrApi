package main

import (
	"flag"
	"fmt"
	"os"

	"llm_ocr_api_go/config"
	"llm_ocr_api_go/router"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.StringVar(configPath, "c", "config.yaml", "配置文件路径 (简写)")
	genConfig := flag.Bool("gen", false, "生成默认配置文件")
	flag.Parse()

	if *genConfig {
		if err := config.GenConfig(*configPath); err != nil {
			fmt.Printf("[-] %v\n", err)
			os.Exit(1)
		}
		return
	}

	config.InitConfig(*configPath)
	cfg := config.LoadConfig()
	r := router.SetupRouter()

	addr := config.GetServerAddr()
	fmt.Printf("[*] Starting LLM OCR API Server on %s (model: %s, config: %s)\n", addr, cfg.DashScope.Model, *configPath)
	if err := r.Run(addr); err != nil {
		fmt.Printf("[-] Server failed to start: %v\n", err)
	}
}
