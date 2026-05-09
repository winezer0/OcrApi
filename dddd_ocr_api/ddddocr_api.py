#!/usr/bin/env python
# encoding: utf-8

import base64
import ddddocr
from flask import Flask, request
from datetime import datetime

from PIL import ImageFile

ImageFile.LOAD_TRUNCATED_IMAGES = True

app = Flask(__name__)
app.config['DEBUG'] = True

# 内置有两套ocr模型，默认情况下不会自动切换，需要在初始化ddddocr的时候通过参数进行切换
# ocr = ddddocr.DdddOcr()  # 可选 模型1


ocr = ddddocr.DdddOcr(beta=True)  # 可选 模型2


@app.route('/ping', methods=['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD'])
def ping():
    """心跳检测接口，允许任何请求方法"""
    return 'pong'


@app.route('/base64ocr', methods=["POST"])
def getCode():
    # 获取当前请求的时间
    start_time = datetime.now()
    print(f"Request At: {start_time}")

    # 接受请求数据
    img_b64 = request.get_data()

    # 输出请求数据
    # print(f"Base64 Img Data: {str(img_b64[:20] if len(img_b64) > 20 else img_b64)} ...")
    print(f"Base64 Img Data Length: {len(img_b64)}.")

    if not img_b64:
        print(f"Base64 Img Data Is Null !!!")
        return ""

    # 解码 base64 图像数据
    try:
        img_bin = base64.b64decode(img_b64.strip())
        print(f"Base64 decoded success.")
    except Exception as e:
        # 当传入的数据不是base64的时候直接当作图片二进制处理
        print(f"Error decoding base64: {e}, POST Possible Img Binary...")
        img_bin = img_b64

    # 使用 OCR 进行识别
    try:
        # img_bin = enlarge_captcha_from_binary(img_bin, scale_factor=2) # 进行图片放大
        ocr_result = ocr.classification(img_bin)  # 常规识别
        # ocr_result = ocr.classification(img_bin, png_fix=True) # 使用 png_fix参数 支持部分透明黑色png格式图片
        print(f"OCR result: {ocr_result}")
    except Exception as e:
        print(f"OCR error: {e}")
        return ""

    # 计算请求处理时间
    end_time = datetime.now()
    processing_time = (end_time - start_time).total_seconds() * 1000

    # 输出信息
    print(f"Processed in {processing_time:.2f}ms")
    return ocr_result


# 通过set_ranges方法设置输出字符范围来限定返回的结果。
# 在调用classification方法的时候传参probability=True，此时 classification 将返回全字符表的概率
# 0	纯整数0-9
# 1	纯小写英文a-z
# 2	纯大写英文A-Z
# 3	小写英文a-z + 大写英文A-Z
# 4	小写英文a-z + 整数0-9
# 5	大写英文A-Z + 整数0-9
# 6	小写英文a-z + 大写英文A-Z + 整数0-9
# 7	默认字符库 - 小写英文a-z - 大写英文A-Z - 整数0-9
# 如果为string类型请传入一段不包含空格的文本，其中的每个字符均为一个待选词 如："0123456789+-x/=""
# 定义字符范围的映射关系
range_descriptions = {
    0: "纯整数 (0-9)",
    1: "纯小写英文 (a-z)",
    2: "纯大写英文 (A-Z)",
    3: "小写英文+大写英文 (a-z A-Z)",
    4: "小写英文+整数 (a-z 0-9)",
    5: "大写英文+整数 (A-Z 0-9)",
    6: "小写英文+大写英文+整数 (a-z A-Z 0-9)",
    7: "默认字符库 (a-z A-Z 0-9)"
}


@app.route('/base64ocr/<string:param>', methods=["POST"])
def getCustomCode(param):
    # 获取当前请求的时间
    start_time = datetime.now()
    print(f"Request At: {start_time}")

    # 接受请求数据
    img_b64 = request.get_data()

    # 输出请求数据
    # print(f"Base64 Img Data: {str(img_b64[:20] if len(img_b64) > 20 else img_b64)} ...")
    print(f"Base64 Img Data Length: {len(img_b64)}.")

    if not img_b64:
        print(f"Base64 Img Data Is Null !!!")
        return ""

    # 解码 base64 图像数据
    try:
        img_bin = base64.b64decode(img_b64.strip())
        print(f"Base64 decoded success.")
    except Exception as e:
        # 当传入的数据不是base64的时候直接当作图片二进制处理
        print(f"Error decoding base64: {e}, POST Possible Img Binary...")
        img_bin = img_b64

    # 使用 OCR 进行识别  # 限定结果范围

    try:
        if len(param) == 1 and param.isdigit():
            param = int(param)
            # 检查 param 是否在有效范围内
            if param not in range_descriptions.keys():
                param = 7
            # 打印限定结果类型的描述
            print(f"限定结果类型为内置范围: {param} -> {range_descriptions[param]}")
        else:
            param = str(param)
            print(f"限定结果类型为指定范围: {param}")

        ocr.set_ranges(param)
        ocr_result = ocr.classification(img_bin, probability=True)
        ocr_result = ''.join(ocr_result['charsets'][i.index(max(i))] for i in ocr_result['probability'])
        print(f"OCR result: {ocr_result}")
    except Exception as e:
        print(f"OCR error: {e}")
        return ""

    # 计算请求处理时间
    end_time = datetime.now()
    processing_time = (end_time - start_time).total_seconds() * 1000

    # 输出信息
    print(f"Processed in {processing_time:.2f}ms")
    return ocr_result


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
