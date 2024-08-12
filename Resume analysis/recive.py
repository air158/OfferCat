import shutil
from paddleocr import PPStructure, save_structure_res
from fastapi import FastAPI, File, UploadFile
from fastapi.responses import JSONResponse
from icecream import ic
from tika import parser
import PyPDF2
import pdfplumber
from pdfminer.pdfinterp import PDFResourceManager, PDFPageInterpreter
from pdfminer.converter import TextConverter
from pdfminer.layout import LAParams
from pdfminer.pdfpage import PDFPage
from io import StringIO
import fitz
import os

# Paddle接口解析必需设置
os.environ["KMP_DUPLICATE_LIB_OK"] = "TRUE"


def method_tika(file):
    parsed = parser.from_buffer(file.file)
    content = parsed['content']
    # 解析失败，使用方式二pdf_reader
    if content is None or len(content.strip()) < 10:
        type = "pdf_reader库"
        pdf_reader = PyPDF2.PdfReader(file.file)
        num_pages = len(pdf_reader.pages)
        content = ""
        ic(num_pages)
        for page_num in range(num_pages):
            page = pdf_reader.pages[page_num]
            content += page.extract_text()
    return content


def method_pdfplumber(file):
    content = ""
    with pdfplumber.open(file.file) as pdf:
        for page in pdf.pages:
            content += page.extract_text()
    return content


def method_pdfminer(file):
    resource_manager = PDFResourceManager()
    # 创建一个StringIO对象，用于存储提取的文本内容
    output = StringIO()
    # 创建一个TextConverter对象
    converter = TextConverter(resource_manager, output, laparams=LAParams())
    # 创建一个PDFPageInterpreter对象
    interpreter = PDFPageInterpreter(resource_manager, converter)
    # 逐页解析文档
    for page in PDFPage.get_pages(file.file):
        interpreter.process_page(page)
    # 获取提取的文本内容
    content = output.getvalue()
    return content


def method_PyMuPDF(save_path):
    doc = fitz.open(save_path)  # 从文件对象读取
    all_content = []
    for i in doc.pages():
        all_content.append(i.get_text())
    content = '\n'.join(all_content)
    return content


def method_paddle(save_path):
    content = ""

    ocr_engine = PPStructure(table=False, ocr=True, show_log=True)

    save_folder = './output'
    img_path = save_path
    result = ocr_engine(img_path)
    for index, res in enumerate(result):
        save_structure_res(res, save_folder, os.path.basename(img_path).split('.')[0], index)

    for res in result:
        for line in res:
            if len(line["res"]) > 0:
                content += line["res"][0]["text"]
    return content


app = FastAPI()


# PDF识别接口
@app.post("/parse-pdf")
async def parse_pdf(file: UploadFile = File(...)):
    file_name = file.filename

    # 设置保存文件的路径（你可以根据需要更改路径）
    save_directory = "uploaded_files"
    os.makedirs(save_directory, exist_ok=True)  # 如果目录不存在则创建

    # 组合保存路径和文件名
    save_path = os.path.join(save_directory, file_name)
    with open(save_path, "wb") as buffer:
        shutil.copyfileobj(file.file, buffer)

    type = "tika库"
    # 方式一，使用tika库进行解析
    content = method_tika(file)

    if content is None or len(content.strip()) < 10:
        type = "pdfplumber方式"
        content = method_pdfplumber(file)
    if content is None or len(content.strip()) < 10:
        type = "pdfminer方式"
        content = method_pdfminer(file)
    if content is None or len(content.strip()) < 10:
        type = "PyMuPDF方式"
        content = method_PyMuPDF(save_path)
    if content is None or len(content.strip()) < 10:
        type = "papermerge方式"
        content = method_paddle(save_path)

    if content is None or len(content.strip()) < 10:
        ic("解析失败！")
    else:
        ic(f"{type}解析成功!")

    return JSONResponse(content=content)


if __name__ == '__main__':
    import uvicorn
    # 广播，端口设置为8848
    uvicorn.run(app, host="0.0.0.0", port=8848)
