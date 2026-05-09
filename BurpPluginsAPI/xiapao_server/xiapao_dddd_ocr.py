#!/usr/bin/env python
# -*- coding: utf-8 -*-
import base64
import re
import time
from urllib.parse import parse_qs

import ddddocr
import requests
from PIL import ImageFile
from flask import Flask, request, jsonify, render_template_string

from xp_utils import save_latest_entries, send_http_package, guess_captcha_format, decode_b64, \
    get_log_content

ImageFile.LOAD_TRUNCATED_IMAGES = True

app = Flask(__name__)

# ocr = ddddocr.DdddOcr()
ocr = ddddocr.DdddOcr(beta=True)

LOG_FILE = 'temp/log.txt'
LOG_COUNT = 50  # 保存多少个验证码及结果


@app.route('/')
def index():
    data = get_log_content(log_file=LOG_FILE)
    return render_template_string(data)


@app.route('/imgurl', methods=['POST'])
def img_url_ocr():
    req_datas = request.data.decode()
    json_req_datas = {k: v[0] for k, v in parse_qs(req_datas).items()}

    xp_url = decode_b64(json_req_datas.get("xp_url", ""))
    xp_type = json_req_datas.get("xp_type", "")
    xp_cookie = decode_b64(json_req_datas.get("xp_cookie", ""))
    xp_set_ranges = json_req_datas.get("xp_set_ranges", "")
    xp_complex_request = decode_b64(json_req_datas.get("xp_complex_request", ""))
    xp_rf = json_req_datas.get("xp_rf", "")
    xp_re = decode_b64(json_req_datas.get("xp_re", ""))
    xp_is_re_run = json_req_datas.get("xp_is_re_run", "")

    print(f"xp_url: {xp_url}")
    print(f"xp_type: {xp_type}")
    print(f"xp_cookie: {xp_cookie}")
    print(f"xp_set_ranges: {xp_set_ranges}")
    print(f"xp_complex_request: {xp_complex_request}")
    print(f"xp_rf: {xp_rf}")
    print(f"xp_re: {xp_re}")
    print(f"xp_is_re_run: {xp_is_re_run}")

    try:
        CAPTCHA = None

        # 普通GET模式传递URL进行验证码获取
        if xp_type == "1":
            headers = {
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
                # "Referer": "https://www.baidu.com",
                "Cookie": xp_cookie
            }
            response = requests.get(xp_url, headers=headers, timeout=3, verify=False)
            CAPTCHA = response.text  # 获取图片
            print("图片地址响应码：", response.status_code)

        # 自定义报文模式解析请求头+传递URL进行验证码获取
        if xp_type == "2":
            response = send_http_package(xp_url, xp_complex_request)
            CAPTCHA = response.text  # 获取图片
            print("图片地址响应码：", response.status_code)

        if CAPTCHA is None:
            print("[!] CAPTCHA数据获取失败!!!")
            return "CAPTCHA ERROR", 500

        # 开启高级RE提取模式
        if xp_is_re_run == "true" and xp_set_ranges == '8':
            try:
                if xp_rf == '0':
                    re_data = re.findall(xp_re, CAPTCHA)[0]
                    print(f"正则匹配结果：[{re_data}]")
                elif xp_rf == '1':
                    rp_head = xp_re.split("|")
                    head_key = rp_head[0]
                    re_zz = xp_re[len(head_key) + 1:]
                    re_data = re.findall(re_zz, response.headers[head_key])[0]
                    print(f"正则匹配结果：[{re_data}]")
            except Exception as error:
                print(f"正则匹配出错: xp_rf=>{xp_rf}  Error: {error}!!!")
                re_data = ""
            # 返回识别结果
            ocr_text = "0000|" + re_data
            return jsonify({"result": ocr_text})

        # 简单判断当前验证码数据格式
        img_is_bin, captcha_base64 = guess_captcha_format(CAPTCHA)

        # 保存验证码图片
        img_bytes = response.content if img_is_bin else base64.b64decode(captcha_base64)

        # 进行验证码识别
        img_time = time.time()
        ocr_text = ddddocr_ocr(img_bytes, xp_set_ranges)
        # 保存最新count个的验证码及识别结果
        save_latest_entries(img_bytes, ocr_text, img_time, xp_type, count=LOG_COUNT, log_file=LOG_FILE)
        return ocr_text, 200

    except Exception as error:
        print(f"Error: {error}")
        return error, 500


def ddddocr_ocr(img_bytes, xp_set_ranges):
    ocr.set_ranges(int(xp_set_ranges))
    ocr_result = ocr.classification(img_bytes, probability=True)
    ocr_text = "".join(ocr_result['charsets'][i.index(max(i))] for i in ocr_result['probability'])
    print(f'识别结果:{ocr_text}')
    return ocr_text


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8899)
