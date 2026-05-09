#!/usr/bin/env python
# encoding: utf-8

import base64
import os

import requests


def image_to_base64(image_path):
    """读取图片并转换为base64文本。"""
    with open(image_path, "rb") as f:
        return base64.b64encode(f.read()).decode("utf-8")


def request_ocr_api(image_b64, api_url):
    """调用本地Flask服务接口并返回响应JSON。"""
    payload = {"imageBase64": image_b64}
    response = requests.post(api_url, json=payload, timeout=30)
    response.raise_for_status()
    return response.json()


def test_api_images(image_dir, api_url):
    """遍历目录图片并调用OCR服务进行测试。"""
    image_names = [name for name in os.listdir(image_dir) if name.lower().endswith((".png", ".jpg", ".jpeg"))]
    image_names.sort()
    for image_name in image_names:
        image_path = os.path.join(image_dir, image_name)
        image_b64 = image_to_base64(image_path)
        result = request_ocr_api(image_b64, api_url)
        print(f"{image_name} => {result}")


if __name__ == "__main__":
    target_dir = r"c:\Users\WINDOWS\Desktop\Deving\OcrApi\yzm_calc"
    service_url = "http://127.0.0.1:5001/ruoyi/base64"
    test_api_images(target_dir, service_url)
