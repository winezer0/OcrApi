#!/usr/bin/env python
# encoding: utf-8
import os

import ddddocr


def recognize_image_by_ddddocr(images_path):
    ocr = ddddocr.DdddOcr()
    with open(images_path, 'rb') as f:
        img_bytes = f.read()
    text = ocr.classification(img_bytes)
    return text


if __name__ == '__main__':
    image_dir = r"..\yzm_num"
    image_names = [name for name in os.listdir(image_dir) if name.lower().endswith((".png", ".jpg", ".jpeg"))]

    image_names.sort()
    for image_name in image_names:
        image_path = os.path.join(image_dir, image_name)
        result = recognize_image_by_ddddocr(image_path)
        print(f"{image_name} => {result}")

