import os
import time
from io import BytesIO

from PIL import Image


def store_img(img_content, memory=True, img_folder='image', img_name=None):
    # 假设 img_content 是你已经解码得到的字节数据
    if memory:
        # 将字节数据转换为图像对象  # 存在bug
        img_path = Image.open(BytesIO(img_content))
    else:
        # 创建目录存储图片
        if not os.path.exists(img_folder):
            os.makedirs(img_folder)
        # 设置文件名称
        if not img_name:
            img_name = f"{time.strftime('%Y%m%d%H%M', time.localtime(time.time()))}.png"
        # 最终物理路径设置
        img_path = f"{img_folder}/{img_name}"
        with open(img_path, 'wb') as file_open:
            file_open.write(img_content)
    return img_path


def extract_result_tuple(data, need_index):
    """
    递归遍历嵌套的列表和元组结构，当遇到元组时提取并返回第一个元素。
    参数:
        data (list or tuple): 嵌套的列表或元组结构。
    返回:
        list: 包含所有从元组中提取的第一个元素的列表。
    """
    extracted = []

    def recurse(item):
        if isinstance(item, list):
            for sub_item in item:
                recurse(sub_item)
        elif isinstance(item, tuple):
            # 如果是元组，添加第一个元素到结果列表
            if len(item) > need_index:
                extracted.append(item[need_index])
        else:
            # 对于其他类型的元素，可以决定如何处理，这里选择忽略它们
            pass
    recurse(data)
    return extracted