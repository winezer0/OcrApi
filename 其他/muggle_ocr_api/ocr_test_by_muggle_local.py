import muggle_ocr

def recognize_image_by_muggle_ocr(images_path):
    sdk = muggle_ocr.SDK(model_type=muggle_ocr.ModelType.Captcha)
    with open(images_path, 'rb') as f:
        img_bytes = f.read()
    text = sdk.predict(image_bytes=img_bytes)
    return text


if __name__ == '__main__':
    images_path = r"yzm.png"
    text = recognize_image_by_muggle_ocr(images_path)
    print(text)
