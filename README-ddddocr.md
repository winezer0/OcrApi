# ddddocr for python 

## **ddddocr_api 打包独立程序**
```sh
1、修改 ddddocr_api.spec 内的 pathex 为当前主机上的Python路径
2、复制 onnxruntime_providers_shared.dll 和 common.onnx 到  ddddocr_api.spec 所在目录，
     PS: 目前代码内好像没有用到 common_old.onnx 和  common_det.onnx
3、执行pyinstaller ddddocr_api.spec
```

## **ddddocr库安装**

**1. 从pypi安装** 
```sh
pip install ddddocr
```

**2. 从源码安装**
```sh
git clone https://github.com/sml2h3/ddddocr.git
cd ddddocr
python setup.py
```

**3. 从源码安装 Python 3.13**
```sh
pip install setuptools==79.0.1 -i http://mirrors.aliyun.com/pypi/simple/ --trusted-host mirrors.aliyun.com
git clone https://github.com/winezer0/ddddocr/ --depth 1
cd ddddocr
python setup.py
```

提示: 
```
setuptools 80.0及以上已经弃用 setup.py 安裝.
Python3.13 安装时需要限定 pip install onnxruntime==1.20.0 高版本会提示动态链接库错误.
```

## 联系方式
如需获取更多信息、技术支持或定制服务，请通过以下方式联系我们： NOVASEC微信公众号或通过社交信息联系开发者【酒零】
![NOVASEC0](https://raw.githubusercontent.com/winezer0/mypics/refs/heads/main/NOVASEC0.jpg)
