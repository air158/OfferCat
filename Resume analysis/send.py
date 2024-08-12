import requests

url = "http://192.168.11.4:8848/parse-pdf"
file_path = r"pdf"
# file_path = r"pdf"
# file_path = r"pdf"

with open(file_path, 'rb') as file:
    files = {'file': file}
    response = requests.post(url, files=files)

# 打印服务器返回的响应
print(response.json())
