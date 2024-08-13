from flask import Flask, Response, render_template, request, stream_with_context,session
import json
import requests
import os

app = Flask(__name__)
app.secret_key = 'your_secret_key'

# 从环境变量获取 API 密钥和 Secret
chat_api_key = os.getenv('SPARK_API_KEY')
chat_api_secret = os.getenv('SPARK_API_SECRET')

chat_key = f'{chat_api_key}:{chat_api_secret}'
print(chat_key)
chat_url = 'https://spark-api-open.xf-yun.com/v1/chat/completions'
chat_model = "generalv3.5"

def stream_response(url, headers, data):
    buffer = ""
    response = requests.post(url, headers=headers, data=json.dumps(data), stream=True)
    for line in response.iter_lines():
        if line:
            decoded_line = line.decode('utf-8')
            if decoded_line.strip() == "data: [DONE]":
                yield "[DONE]"
                break
            else:
                try:
                    json_data = json.loads(decoded_line[6:])
                    content = json_data['choices'][0]['delta']['content']
                    buffer += content

                    # 检查是否有完整的问题（以 '@' 结尾）
                    while '@' in buffer:
                        question, buffer = buffer.split('@', 1)
                        question = question.strip()
                        yield f"{question}@"
                except (KeyError, json.JSONDecodeError):
                    pass

@app.route('/stream_questions', methods=['GET'])
def stream_questions():
    job_title = request.args.get('job_title')
    job_description = request.args.get('job_description')
    resume_text = request.args.get('resume_text')
    
    prompt = f"岗位名称：{job_title}\n" \
             f"岗位要求：\n{job_description}\n" \
             f"面试者简历：\n{resume_text}\n" \
             f"\n你是这个 {job_title} 岗位的面试官，请依据 岗位要求 和 面试者简历 为面试者给出 5 道面试题。\n" \
             f"面试题的流程是先让面试者进行自我介绍，然后询问项目经历，接着询问基础知识（八股文），最后出算法题。\n" \
             f"请只给我面试题，不要输出其他无关内容，面试题用口语的形式表达，每个问题是一行，@代表这个问题结束"
    
    print('prompt ', prompt)

    headers = {
        'Authorization': f'Bearer {chat_key}',
        'Content-Type': 'application/json'
    }
    data = {
        "model": chat_model,
        "messages": [
            {"role": "user", "content": prompt}
        ],
        "stream": True
    }

    questions = []

    def generate():
        for chunk in stream_response(chat_url, headers, data):
            if chunk == "[DONE]":
                print("DONE")
                print('session ',session['questions'])
                break
            print(chunk)
            if chunk != '@':
                questions.append(chunk)
            yield f"data: {chunk}\n\n"
    # 保存生成的问题到 session 中
    session['questions'] = questions
    print('session ',session['questions'])
    return Response(stream_with_context(generate()), content_type='text/event-stream')

@app.route('/')
def index():
    return render_template('interview.html')

if __name__ == '__main__':
    app.run(debug=True)