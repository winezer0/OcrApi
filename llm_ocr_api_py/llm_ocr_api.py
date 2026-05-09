#!/usr/bin/env python
# encoding: utf-8

import base64
from datetime import datetime
from typing import Optional

from fastapi import FastAPI, Request, HTTPException
from pydantic import BaseModel
import uvicorn

from uitls import call_dashscope_ocr_async, extract_result_number, get_dashscope_api_key

app = FastAPI(title="LLM OCR API (FastAPI)")

class OcrRequest(BaseModel):
    imageBase64: Optional[str] = None

def normalize_base64(text: str):
    if not text: return ""
    text = text.strip()
    if "," in text and text.lower().startswith("data:image"):
        text = text.split(",", 1)[1]
    return text.strip()

@app.api_route("/ping", methods=["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"])
async def ping():
    """心跳检测接口，允许任何请求方法"""
    return "pong"

@app.post("/ruoyi/base64")
async def ruoyi_base64(request: Request, body: Optional[OcrRequest] = None):
    start_time = datetime.now()
    
    # 提取 base64
    image_b64 = ""
    if body and body.imageBase64:
        image_b64 = normalize_base64(body.imageBase64)
    else:
        raw_data = await request.body()
        image_b64 = normalize_base64(raw_data.decode("utf-8", errors="ignore"))

    if not image_b64:
        raise HTTPException(status_code=400, detail="image base64 is required")

    api_key = get_dashscope_api_key()
    if not api_key:
        raise HTTPException(status_code=500, detail="DASHSCOPE_API_KEY is empty")

    try:
        raw_text = await call_dashscope_ocr_async(image_b64, api_key)
        result = extract_result_number(raw_text)
        
        duration_ms = (datetime.now() - start_time).total_seconds() * 1000
        print(f"[*] Async Request done: {duration_ms:.2f} ms | result: {result}")
        
        return {"code": 200, "msg": "success", "result": result, "raw": raw_text}
    except Exception as e:
        print(f"[-] OCR Error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=5001)
