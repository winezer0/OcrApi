#!/usr/bin/env python
# encoding: utf-8
import base64

import easyocr
from flask import Flask, request

from utils import store_img

app = Flask(__name__)
app.config['DEBUG'] = True


@app.route('/', methods=["POST"])
def getCode():
    img_b64 = request.get_data()
    img_content = base64.b64decode(img_b64.strip())
    # 将字节数据写入文件或内存
    img_path = store_img(img_content, memory=True, img_folder='image', img_name=None)
    # print(img_path)
    # 识别图片
    result = easy_ocr(easyocr, img_path)
    return str(result)


# 初始化识别器
reader = easyocr.Reader(['ch_sim', 'en'])


def easy_ocr(img_path):
    result = reader.readtext(img_path, detail=0)
    return result


if __name__ == '__main__':
    app.run()
