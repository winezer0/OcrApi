from flask import Flask, request
import base64
import muggle_ocr

app = Flask(__name__)
app.config['DEBUG'] = True


@app.route('/', methods=["POST"])
@app.route('/base64ocr', methods=["POST"])
def getCode():
    img_b64 = request.get_data()
    img_bytes = base64.b64decode(img_b64.strip())
    sdk = muggle_ocr.SDK(model_type=muggle_ocr.ModelType.Captcha)
    text = sdk.predict(image_bytes=img_bytes)
    print(text)
    return text


if __name__ == '__main__':
    app.run()
