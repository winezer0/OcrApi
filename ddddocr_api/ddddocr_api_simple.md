## 为ddddocr提供web接口

> 环境：python 3.8.5
>
> 依赖：flask, ddddocr （均可通过pip安装）

配好环境，安装依赖。

执行 python3 ddddocr_api_simple.py，启动flask web服务。

+ 接口URL就是web服务的地址，默认是`http://127.0.0.1:5000`。

+ 请求模板如下：

  ```http
  POST / HTTP/1.1
  
  iVBORw0KGgoAAAANSUhEUgAAAFAAAAArCAIAAABglpj4AAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAABApSURBVGhDXdl36Pbj
  
  返回文本即为验证码
  
  POST /base64ocr HTTP/1.1
  
  iVBORw0KGgoAAAANSUhEUgAAAFAAAAArCAIAAABglpj4AAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAABApSURBVGhDXdl36Pbj
  ```
