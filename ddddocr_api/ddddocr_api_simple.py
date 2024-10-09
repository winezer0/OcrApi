from flask import Flask, request
import json
import ddddocr
import base64

app = Flask(__name__)
app.config['DEBUG'] = True
ocr = ddddocr.DdddOcr()


@app.route('/', methods=["POST"])
@app.route('/base64ocr', methods=["POST"])
def getCode():
    img_b64 = request.get_data()
    print(f"img_b64:{img_b64}")
    img = base64.b64decode(img_b64.strip())
    text = ocr.classification(img)
    print(text)
    return text


if __name__ == '__main__':
    app.run()
