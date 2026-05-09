# OcrApi

轻量级 OCR API 服务聚合平台，封装多种 OCR 引擎提供统一 HTTP 接口，面向安全/自动化场景提供验证码识别能力。

## 免责声明
继续阅读文章或使用工具视为您已同意NOVASEC免责声明
[NOVASEC免责声明](https://mp.weixin.qq.com/s/iRWRVxkYu7Fx5unxA34I7g)

## 项目结构

```
OcrApi/
├── dddd_ocr_api/          # DDDDOCR 验证码识别 API（Python/Flask）
├── llm_ocr_api_py/        # LLM 大模型 OCR API（Python/FastAPI）
├── llm_ocr_api_go/        # LLM 大模型 OCR API（Go/Gin）
├── llm_ocr_api_go_v2/     # LLM + DDDDOCR 混合 OCR API（Go/Gin）
├── ruoyi_login/           # 若依登录爆破工具（Python）
├── ruoyi_login_go/        # 若依登录爆破工具（Go）
├── other_ocr/             # 其他 OCR 引擎 API 集合
│   ├── easy_ocr_api/      #   EasyOCR 多语言识别
│   ├── paddle_ocr_api/    #   PaddleOCR 百度飞桨识别
│   ├── pytesseract_ocr/   #   Tesseract OCR 识别
│   └── rapidocr_api/      #   RapidOCR 轻量识别
├── BurpPluginsAPI/        # Burp Suite 插件适配服务端
│   ├── captcha-killer/    #   captcha-killer 插件服务端
│   └── xiapao_server/     #   瞎跑插件服务端（多引擎）
├── deprecated/            # 已废弃模块
│   └── muggle_ocr_api/    #   Muggle OCR（已失效）
├── yzm_calc/              # 计算型验证码样本图片
└── yzm_num/               # 纯数字验证码样本图片
```

## 各模块说明

### dddd_ocr_api
基于 [ddddocr](https://github.com/sml2h3/ddddocr) 的验证码识别 HTTP API 服务（Python/Flask）。  
支持 base64 图片识别，内置两套 OCR 模型可切换，提供 `/base64ocr` 和 `/ping` 接口。

### llm_ocr_api_py
基于阿里云 DashScope（通义千问）大模型的验证码识别 API 服务（Python/FastAPI）。  
通过多模态大模型识别计算型验证码（如 `3+5=?`），自动完成算术运算并返回结果。  
API Key 通过环境变量 `DASHSCOPE_API_KEY` 配置。

### llm_ocr_api_go
LLM 大模型 OCR 的 Go 语言实现（Gin 框架），功能与 Python 版一致。  
提供 `/ruoyi/base64` 接口，专为若依验证码场景优化。  
通过 `config.yaml` 配置 DashScope API Key、模型名称等参数。

### llm_ocr_api_go_v2
在 `llm_ocr_api_go` 基础上增加了本地 DDDDOCR 识别能力（Go/Gin）。  
同时提供 LLM 识别（`/ruoyi/base64`）和 DDDDOCR 识别（`/ddddocr/base64`）两种接口，  
可根据验证码类型选择不同引擎。

### ruoyi_login
若依（RuoYi）系统登录爆破工具（Python），多线程并发，自动获取验证码并调用 OCR 服务识别，  
支持断点续跑、结果缓存、CSV 导出。

### ruoyi_login_go
若依登录爆破工具的 Go 语言重写版本，性能更优。  
支持并发 worker、失败/成功缓存、进度条、日志分级、CSV 结果导出。

### other_ocr
其他 OCR 引擎的 API 封装集合，均为 Python/Flask 实现：
- **easy_ocr_api** — 基于 EasyOCR，支持多语言文字识别
- **paddle_ocr_api** — 基于百度 PaddleOCR，支持中文及角度校正
- **pytesseract_ocr** — 基于 Tesseract OCR，经典开源 OCR 引擎
- **rapidocr_api** — 基于 RapidOCR（ONNX Runtime），轻量高效

### BurpPluginsAPI
适配 Burp Suite 验证码爆破插件的服务端：
- **captcha-killer** — 适配 [captcha-killer-modified](https://github.com/f0ng/captcha-killer-modified) 插件，支持多种识别模式（纯数字/字母/混合），带 Basic Auth 认证
- **xiapao_server** — 适配瞎跑插件，支持 ddddocr/easyocr/paddleocr/pytesseract/rapidocr 多引擎切换

### deprecated
已废弃的模块：
- **muggle_ocr_api** — Muggle OCR 识别服务，因库已停止维护而失效

## 联系方式
如需获取更多信息、技术支持或定制服务，请通过以下方式联系我们： NOVASEC微信公众号或通过社交信息联系开发者【酒零】
![NOVASEC0](https://raw.githubusercontent.com/winezer0/mypics/refs/heads/main/NOVASEC0.jpg)
