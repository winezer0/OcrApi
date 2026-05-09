#!/usr/bin/env python
# encoding: utf-8
import base64

import pytesseract
from flask import Flask, request

# 如果PATH中没有tesseract可执行文件，请指定tesseract路径  当前 pytesseract:5.0.0
from utils import store_img

pytesseract.pytesseract.tesseract_cmd = r'C:\ISEC.Installed\Tesseract-OCR\tesseract.exe'

app = Flask(__name__)
app.config['DEBUG'] = True


@app.route('/', methods=["POST"])
@app.route('/base64ocr', methods=["POST"])
def getCode():
    img_b64 = request.get_data()
    print(img_b64)
    img_content = base64.b64decode(img_b64.strip())
    print(f"成功读取img数据Length: {len(img_content)}")
    try:
        # 将字节数据写入文件或内存
        img_path = store_img(img_content, memory=True, img_folder='image', img_name=None)
        # print(img_path)
        text = pytess_ocr(img_path)
        return text
    except Exception as error:
        print(f"验证码识别发生错误:{error}")
        return "NONE"


def pytess_ocr(img_path):
    # 指定语言识别图像字符串,eng为英语
    text = pytesseract.image_to_string(img_path, lang='eng', config="--psm 7").strip()
    return text


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
