import os
import time

from flask import Flask, request
from paddleocr import PaddleOCR
import base64

app = Flask(__name__)
app.config['DEBUG'] = True


@app.route('/', methods=["POST"])
def getCode():
    img_b64 = request.get_data()
    img_content = base64.b64decode(img_b64.strip())
    # 内容写入文件
    img_folder = 'image'
    if not os.path.exists(img_folder): os.makedirs(img_folder)
    times = time.strftime('%Y%m%d%H%M', time.localtime(time.time()))
    img_path = img_folder + '/' + times + '.png'
    with open(img_path, 'wb') as file_open:
        file_open.write(img_content)
    # 重新识别图片
    ocr = PaddleOCR(use_angle_cls=True, lang="ch", show_log=False)  # need to run only once to download and load model into memory
    result = ocr.ocr(img_path, cls=True)
    text = result[-1][-1][0]
    print(text)
    return text


if __name__ == '__main__':
    app.run()
