import requests

url = "http://ip:8848/parse-pdf"
file_path = r"简历(1).pdf"


with open(file_path, 'rb') as file:
    files = {'file': file}
    response = requests.post(url, files=files)

# 打印服务器返回的响应
print(response.json())
