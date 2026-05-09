#!/usr/bin/env python
# encoding: utf-8

import asyncio
import os

from uitls import call_dashscope_ocr_async, extract_result_number, get_dashscope_api_key, image_to_base64


async def test_local_images(image_dir):
    """遍历目录中的验证码图片并输出识别结果。"""
    api_key = get_dashscope_api_key()
    image_names = [name for name in os.listdir(image_dir) if name.lower().endswith((".png", ".jpg", ".jpeg"))]
    image_names.sort()
    for image_name in image_names:
        image_path = os.path.join(image_dir, image_name)
        image_b64 = image_to_base64(image_path)
        raw_text = await call_dashscope_ocr_async(image_b64, api_key)
        result = extract_result_number(raw_text)
        print(f"{image_name} => result={result}, raw={raw_text}")


if __name__ == "__main__":
    target_dir = r"c:\Users\WINDOWS\Desktop\Deving\OcrApi\yzm_calc"
    asyncio.run(test_local_images(target_dir))
