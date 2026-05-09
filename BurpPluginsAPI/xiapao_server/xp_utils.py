import base64
import os
import re
import time
from collections import deque
from io import BytesIO
from PIL import Image
import requests
from PIL import ImageFile
ImageFile.LOAD_TRUNCATED_IMAGES = True


def save_latest_entries(img_bytes, ocr_text, img_time, xp_type, count=50, log_file="tmp.txt"):
    log_entry_format = '<tr align=center><td><img src="data:image/png;base64,%s"/></td><td>%s</td><td>%s</td><td>%s</td></tr>\n'

    # 确保日志文件所在的目录存在
    os.makedirs(os.path.dirname(log_file), exist_ok=True)

    # 使用deque来保持最新的count个条目
    entries = deque(maxlen=count)

    # 尝试读取现有日志内容，并将每条记录加入到deque中
    if os.path.exists(log_file):
        with open(log_file, 'r') as f:
            for line in f:
                if line.strip():  # 忽略空行
                    entries.append(line)

    # 构建新的日志条目
    new_entry = log_entry_format % (
        base64.b64encode(img_bytes).decode("utf-8"),
        ocr_text,
        time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(img_time))),
        xp_type
    )

    # 添加新条目到队列开头
    entries.appendleft(new_entry)

    # 写入更新后的日志文件
    with open(log_file, 'w') as f:
        f.writelines(entries)


def get_log_content(log_file="tmp.txt"):
    content = ""

    if os.path.exists(log_file):
        with open(log_file, 'r') as f:
            content = f.read()

    content = '''
    <title>xp_CAPTCHA</title>
    <body style="text-align:center">
        <h1>验证码识别：xp_CAPTCHA V4.3</h1>
        <a href="http://www.nmd5.com">author:算命縖子</a>
        <p>
            <TABLE style="BORDER-RIGHT: #ff6600 2px dotted; BORDER-TOP: #ff6600 2px dotted; BORDER-LEFT: #ff6600 2px dotted; BORDER-BOTTOM: #ff6600 2px dotted; BORDER-COLLAPSE: collapse" borderColor=#ff6600 height=40 cellPadding=1 align=center border=2>
                <tr align=center>
                    <td>验证码</td>
                    <td>识别结果</td>
                    <td>时间</td>
                    <td>验证码模块</td>
                </tr>
                {}
            </table>
        </p>
    </body>
    '''.format(content)
    return content


def guess_captcha_format(captcha_data):
    img_is_bin = True
    captcha_base64 = None
    if re.findall('"\s*:\s*.?"', captcha_data):
        print("img data is [base64 json] format")
        captcha_data = captcha_data.split('"')
        captcha_data.sort(key=lambda i: len(i), reverse=True)
        captcha_data = captcha_data[0].split(',')
        captcha_data.sort(key=lambda i: len(i), reverse=True)
        captcha_base64 = captcha_data[0]
        img_is_bin = False
    elif re.findall('data:image/\D*;base64,', captcha_data):
        print("img data is [base64] format")
        captcha_data = captcha_data.split(',')
        captcha_data.sort(key=lambda i: len(i), reverse=True)
        captcha_base64 = captcha_data[0]
        img_is_bin = False
    else:
        print("img data is [bin] format")
    return img_is_bin, captcha_base64


def parse_http_package(http_package):
    def find_body_index(http_package_):
        # 假设空行（CRLF或LF）分隔headers和body
        empty_line_regex = r'\r\n\r\n|\n\n|\r\r'
        match_ = re.search(empty_line_regex, http_package_)
        return match_.start() if match_ else -1

    def parse_headers(http_package_, body_index_):
        # 提取header部分并解析为字典
        header_part = http_package_[:body_index_] if body_index_ >= 0 else http_package_
        lines = header_part.splitlines()
        headers_ = {}
        for line in lines[1:]:  # 忽略第一行，即请求行
            if ':' in line:
                key, value = line.split(':', 1)
                headers_[key.strip()] = value.strip()
        return headers_

    # 正则表达式匹配HTTP方法，不区分大小写
    match_method = re.match(r'^(?P<method>[A-Z]+)', http_package, re.IGNORECASE)
    if not match_method:
        raise ValueError("无法识别的请求类型")

    method = match_method.group('method').upper()
    body_index = find_body_index(http_package)
    headers = parse_headers(http_package, body_index)

    # 分离body部分，如果存在的话
    body = http_package[body_index + 2:].strip() if body_index >= 0 else None

    return method, headers, body


def decode_b64(base64_data):
    try:
        if len(base64_data) > 0:
            base64_data = base64.b64decode(base64_data).decode("utf-8")
    except Exception as error:
        print(f"base64 decode [{base64_data}] -> Error:{error}")
    return base64_data


def send_http_package(url, http_package):
    # method, headers, body = parse_http_package(http_package)
    method, headers, body = parse_http_package(http_package)
    # 使用method参数来指定请求方法，并且仅在需要的时候（如POST）提供data参数
    if method in ["GET", "POST"]:
        response = requests.request(method, url, headers=headers, data=body if body else None, timeout=3, verify=False)
        return response
    else:
        raise ValueError("暂不支持重放对应请求")


def store_img_content(img_content, memory=True, img_folder='image', img_name=None, enlarge=False):
    if enlarge:
        # 从二进制数据加载图像，按比例放大，并返回新的二进制数据。 好像没啥用
        img_content  = enlarge_captcha_from_binary(img_content)

    if memory:
        # 将字节数据转换为图像对象
        img_path = Image.open(BytesIO(img_content))
        # 存在错误 提示没有 open方法
    else:
        # 创建目录存储图片
        if not os.path.exists(img_folder):
            os.makedirs(img_folder)
        # 设置文件名称
        if not img_name:
            img_name = f"{time.strftime('%Y%m%d%H%M', time.localtime(time.time()))}.jpg"
        # 最终物理路径设置
        img_path = f"{img_folder}/{img_name}"
        with open(img_path, 'wb') as file_open:
            file_open.write(img_content)
    return img_path


def enlarge_captcha_from_binary(binary_data, scale_factor=2):
    # 将二进制数据加载到内存文件对象中
    input_image_file = BytesIO(binary_data)

    # 使用Pillow打开图像
    with Image.open(input_image_file) as img:
        # 获取原始尺寸
        original_size = img.size
        # 计算新的尺寸
        new_size = (original_size[0] * scale_factor, original_size[1] * scale_factor)
        # 调整图像大小，使用NEAREST模式避免抗锯齿影响验证码识别
        enlarged_img = img.resize(new_size, Image.NEAREST)

        # 创建一个新的内存文件对象以保存调整后的图像
        output_image_file = BytesIO()
        # 保存调整后的图像到内存文件对象
        enlarged_img.save(output_image_file, format=img.format)
        # 移动指针到文件开头以便读取
        output_image_file.seek(0)
        # 获取二进制数据
        enlarged_binary_data = output_image_file.getvalue()

    return enlarged_binary_data


def extract_tuple_in_list(data, need_index):
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