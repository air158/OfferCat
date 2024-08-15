requirements如下：

```python
fastapi==0.112.0
fitz==0.0.1.dev2
icecream==2.1.3
paddleocr==2.8.1
paddleocr.egg==info
pdfminer==20191125
pdfminer.six==20220524
pdfplumber==0.7.4
PyPDF2==3.0.1
Requests==2.32.3
tika==2.6.0
uvicorn==0.30.5
```

避坑：

其中paddle框架请参考如下链接进行安装

https://www.paddlepaddle.org.cn/install/quick

OCR解析所需支撑库参考下面的链接进行安装

https://paddlepaddle.github.io/PaddleOCR/ppstructure/quick_start.html

fitz库无需安装，请直接安装PyMuPDF库，避坑参考

https://github.com/pymupdf/PyMuPDF/issues/660

https://stackoverflow.com/questions/69160152/pymupdf-attributeerror-module-fitz-has-no-attribute-open