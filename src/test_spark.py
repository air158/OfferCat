import os
import requests
from flask import Flask, Response, request, send_from_directory, render_template, stream_with_context
import json
import ssl

app = Flask(__name__)

api_key = os.getenv('SPARK_API_KEY')
api_secret = os.getenv('SPARK_API_SECRET')

url = 'https://spark-api-open.xf-yun.com/v1/chat/completions'
headers = {
    'Authorization': f'Bearer {api_key}:{api_secret}',
    'Content-Type': 'application/json'
}

#这个是阿里云服务器的ssl证书注册文件
@app.route('/.well-known/pki-validation/B805A860A67DC25F8D2B06146E189A02.txt')
def serve_validation_file():
    # 假设你的文件位于项目根目录下的 `.well-known/pki-validation/` 文件夹中
    directory = '.well-known/pki-validation'
    filename = 'B805A860A67DC25F8D2B06146E189A02.txt'
    return send_from_directory(directory, filename)



@app.route('/')
def index():
    return render_template('index.html')

def stream_response(url, headers, data):
    response = requests.post(url, headers=headers, data=json.dumps(data), stream=True)
    for line in response.iter_lines():
        if line:
            decoded_line = line.decode('utf-8')
            print(f"decoded_line: {decoded_line}")
            if decoded_line == "data: [DONE]":
                yield "[DONE]"
                break
            else:
                try:
                    json_data = json.loads(decoded_line[6:])
                    content = json_data['choices'][0]['delta']['content']
                    yield f"{content}"
                except (KeyError, json.JSONDecodeError):
                    pass
            
@app.route('/stream', methods=['POST'])
def stream():
    user_input = request.form['user_input']
    data = {
        "model": "generalv3.5",
        "messages": [
            {
                "role": "user",
                "content": user_input
            }
        ],
        "stream": True
    }
    return Response(stream_with_context(stream_response(url, headers, data)), content_type='text/event-stream')

if __name__ == '__main__':

    # #https

    # 这里设置SSL证书和密钥文件的路径
    context = ('/etc/ssl/certificate.crt', '/etc/ssl/private/private.key')
    app.run(host='0.0.0.0', port=443, ssl_context=context)

    # context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
    # context.load_cert_chain(certfile='ssl/cert.pem', keyfile='ssl/key.pem', password=ssl_password)  # 替换为您的密码
    # app.run(host='0.0.0.0', port=443, debug=True, ssl_context=context)

    #http
    # app.run(host='0.0.0.0', port=80, debug=True)
