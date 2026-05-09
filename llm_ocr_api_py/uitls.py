#!/usr/bin/env python
# encoding: utf-8

import base64
import os
import re
from openai import AsyncOpenAI

PROMPT_TEXT = (
    "你是验证码识别与计算助手。"
    "图片内容固定为数字 操作符 数字 = ?格式。"
    "请识别并完成计算。"
    "只输出最终结果数字，不要输出任何解释、空格、标点或其他字符。"
)
DASHSCOPE_BASE_URL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
DASHSCOPE_MODEL = os.getenv("DASHSCOPE_MODEL", "qwen3.5-flash")
DEFAULT_API_KEY = "sk-xxxxxxxxxxxxxxxxxxxxxxx"

# 全局客户端实例，避免重复创建
_dashscope_client = None

def get_dashscope_api_key():
    """获取 DashScope API 密钥"""
    return os.getenv("DASHSCOPE_API_KEY", DEFAULT_API_KEY)

def get_dashscope_client():
    """获取或创建全局 DashScope 客户端实例"""
    global _dashscope_client
    if _dashscope_client is None:
        api_key = get_dashscope_api_key()
        _dashscope_client = AsyncOpenAI(api_key=api_key, base_url=DASHSCOPE_BASE_URL)
    return _dashscope_client

def image_to_base64(image_path):
    """将图片文件转换为 base64 编码"""
    with open(image_path, "rb") as f:
        return base64.b64encode(f.read()).decode("utf-8")

async def call_dashscope_ocr_async(image_b64, api_key):
    """异步调用大模型进行识别"""
    client = get_dashscope_client()
    image_url = f"data:image/png;base64,{image_b64}"
    
    response = await client.chat.completions.create(
        model=DASHSCOPE_MODEL,
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": PROMPT_TEXT},
                    {"type": "image_url", "image_url": {"url": image_url}},
                ],
            }
        ],
        temperature=0,
    )
    content = response.choices[0].message.content
    return str(content).strip()

def extract_result_number(text):
    """从文本中提取计算结果数字"""
    content = text or ""
    exp_match = re.search(r"(-?\d+)\s*([+\-xX*/÷])\s*(-?\d+)", content)
    if exp_match:
        left = int(exp_match.group(1))
        operator = exp_match.group(2)
        right = int(exp_match.group(3))
        if operator == "+": return str(left + right)
        if operator == "-": return str(left - right)
        if operator in ("x", "X", "*"): return str(left * right)
        if operator in ("÷", "/") and right != 0: return str(left // right)
    numbers = re.findall(r"-?\d+", content)
    return numbers[-1] if numbers else ""
