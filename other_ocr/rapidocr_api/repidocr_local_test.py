#!/usr/bin/env python
# encoding: utf-8

from rapidocr_onnxruntime import RapidOCR


def recognize_image_by_repidocr(images_path):
    engine = RapidOCR()
    result, elapse = engine(images_path)
    result = [x[1] for x in result]  # ['8', '9', '2', '5']
    return "".join(result)


if __name__ == '__main__':
    images_path = r"../yzm2.jpg"
    text = recognize_image_by_repidocr(images_path)
    print(text)
