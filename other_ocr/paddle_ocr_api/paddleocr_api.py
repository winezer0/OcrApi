#!/usr/bin/env python
# encoding: utf-8
import base64

from flask import Flask, request
from paddleocr import PaddleOCR

from utils import store_img

app = Flask(__name__)
app.config['DEBUG'] = True

ocr = PaddleOCR(use_angle_cls=True, lang="ch", show_log=False)

@app.route('/', methods=["POST"])
@app.route('/base64ocr', methods=["POST"])
def getCode():
    img_b64 = request.get_data()
    img_content = base64.b64decode(img_b64.strip())
    # 将字节数据写入文件或内存
    img_path = store_img(img_content, memory=True, img_folder='image', img_name=None)
    # print(img_path)
    # 重新识别图片
    text = paddle_ocr(img_path)
    return text


def paddle_ocr(img_path):
    result = ocr.ocr(img_path, cls=False)
    print(f"识别结果:{result}")
    text = result[0][0][0]
    return text


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
