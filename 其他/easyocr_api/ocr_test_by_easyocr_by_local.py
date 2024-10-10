import easyocr


def recognize_image_by_easyocr(images_path):
    reader = easyocr.Reader(['ch_sim', 'en'])  # this needs to run only once to load the model into memory
    result = reader.readtext(images_path, detail=0)
    text = ''.join(result)
    return text


if __name__ == '__main__':
    images_path = r"yzm.png"
    text = recognize_image_by_easyocr(images_path)
    print(text)
