import os
import time
from flask import Flask, request
import base64
import easyocr

app = Flask(__name__)
app.config['DEBUG'] = True


@app.route('/', methods=["POST"])
@app.route('/base64ocr', methods=["POST"])
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
    reader = easyocr.Reader(['ch_sim', 'en'])
    result = reader.readtext(img_path, detail=0)
    text = ''.join(result)
    print(text)
    return text


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
