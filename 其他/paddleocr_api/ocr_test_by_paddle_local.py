from paddleocr import PaddleOCR

def recognize_image_by_paddleocr(images_path):
    ocr = PaddleOCR(use_angle_cls=True, lang="ch", show_log=False)  # need to run only once to download and load model into memory
    result = ocr.ocr(images_path, cls=True)
    text = result[-1][-1][0]
    return text


if __name__ == '__main__':
    images_path = r"C:\Users\WINDOWS\Desktop\wujiyzm.png"
    text = recognize_image_by_paddleocr(images_path)
    print(text)
