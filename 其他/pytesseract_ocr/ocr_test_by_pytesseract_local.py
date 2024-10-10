import pytesseract
from PIL import Image
import cv2
# 如果PATH中没有tesseract可执行文件，请指定tesseract路径
# pytesseract.pytesseract.tesseract_cmd='C:\Program Files (x86)\Tesseract-OCR\\tesseract.exe'


def ocr_by_tesseract(image_path):

    print(f"pytesseract:{pytesseract.get_tesseract_version()}")
    im_content = Image.open(image_path)
    # text = pytesseract.image_to_string(im_content)

    # 指定语言识别图像字符串,eng为英语
    try:
        #image = cv2.imread(image_path)
        #gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        text = pytesseract.image_to_string(im_content, lang='eng', config="--psm 7").strip()
        print(text)
    except Exception as error:
        print(error)
    # 获取图像边界框
    # print(pytesseract.image_to_boxes(im_content))
    return text


if __name__ == '__main__':
    image_path = 'yzm.png'
    # image_path = r'C:\Users\WINDOWS\Desktop\image.png'

    text = ocr_by_tesseract(image_path)
    print(text)
