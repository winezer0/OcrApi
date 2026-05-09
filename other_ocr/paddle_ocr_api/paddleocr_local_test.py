#!/usr/bin/env python
# encoding: utf-8
from paddleocr import PaddleOCR


def recognize_image_by_paddleocr(images_path):
    ocr = PaddleOCR(use_angle_cls=True, lang="ch",
                    show_log=False)  # need to run only once to download and load model into memory
    result = ocr.ocr(images_path, det=False, cls=False)

    for idx in range(len(result)):
        res = result[idx]
        for line in res:
            print(f"验证码及可能性元组:{line}")
    text = result[0][0][0]
    return text


if __name__ == '__main__':
    images_path = r"C:\Users\WINDOWS\GithubProject\OcrApi\插件API\xiapao_server\image\202501140038.jpg"
    text = recognize_image_by_paddleocr(images_path)
    print(text)
