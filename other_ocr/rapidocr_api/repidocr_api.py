#!/usr/bin/env python
# encoding: utf-8
import base64

from flask import Flask, request
from rapidocr_onnxruntime import RapidOCR

from utils import store_img

app = Flask(__name__)
app.config['DEBUG'] = True
engine = RapidOCR()


@app.route('/', methods=["POST"])
@app.route('/base64ocr', methods=["POST"])
def getCode():
    img_b64 = request.get_data()
    img_content = base64.b64decode(img_b64.strip())
    # 将字节数据写入文件或内存
    img_path = store_img(img_content, memory=True, img_folder='image', img_name=None)
    # print(img_path)
    # 重新识别图片
    text = repid_ocr(img_path)
    return text


def repid_ocr(img_path):
    result, elapse = engine(img_path)
    result = [x[1] for x in result]  # ['8', '9', '2', '5']
    return "".join(result)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
