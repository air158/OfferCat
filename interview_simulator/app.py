from flask import Flask, render_template, request, redirect, url_for, session, jsonify, Response, stream_with_context
from models import db, JobInfo, InterviewRecord
import time
import uuid  # 用于生成唯一的面试ID
import os
import requests
import json

# 设置倒计时时长（例如60秒）
countdown_time = 60

#面试题数量
ques_len = "5"

chat_api_key = os.getenv('SPARK_API_KEY')
chat_api_secret = os.getenv('SPARK_API_SECRET')

chat_key = f'{chat_api_key}:{chat_api_secret}'
chat_model = 'generalv3.5'

chat_url = 'https://spark-api-open.xf-yun.com/v1/chat/completions'

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///database.db'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
app.secret_key = 'your_secret_key'

db.init_app(app)

# 初始化页面，开始新的面试
@app.route('/', methods=['GET', 'POST'])
def init():
    if request.method == 'POST':
        job_title = request.form['job_title']
        job_description = request.form['job_description']
        resume_text = request.form['resume_text']

        session['job_title'] = job_title
        session['job_description'] = job_description
        session['resume_text'] = resume_text
        session['interview_id'] = str(uuid.uuid4())  # 生成一个新的面试ID

        # 调用大模型生成问题
        # questions = generate_interview_questions(job_title, job_description, resume_text)
        # # 模拟大模型生成问题
        # questions = [
        #     "Can you explain a challenging project you've worked on?",
        #     "How do you handle tight deadlines?"
        # ]
        # session['questions'] = questions
        return redirect(url_for('question'))

    return render_template('init.html')

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
    job_title = session['job_title']
    job_description = session['job_description']
    resume_text = session['resume_text']
    
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

# 面试页面
@app.route('/interview', methods=['GET', 'POST'])
def interview():
    if 'questions' not in session:
        return redirect(url_for('init'))

    if request.method == 'POST':
        answer = request.form['answer']
        duration = request.form['duration']  # 接收前端提交的持续时间
        question = session['questions'].pop(0)
        session.modified = True  # 标记 session 已修改
        interview_id = session['interview_id']  # 获取当前面试的ID

        # 保存历史记录
        record = InterviewRecord(
            job_title=session['job_title'],
            question=question,
            answer=answer,
            duration=float(duration),  # 转换为浮点数保存
            timestamp=time.time(),
            interview_id=interview_id  # 保存当前面试的ID
        )
        db.session.add(record)
        db.session.commit()

        if not session['questions']:
            return redirect(url_for('result'))

    current_question = session['questions'][0] if session['questions'] else None
    interview_id = session['interview_id']
    history = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 仅获取当前面试的记录
    return render_template('interview.html', question=current_question, countdown_time=countdown_time, history=history)

# 结果页面
@app.route('/result', methods=['GET'])
def result():
    interview_id = session.get('interview_id')
    records = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 获取当前面试的记录
    # # 模拟大模型生成改进建议
    # suggestions = "Consider providing more specific examples and quantify your achievements when possible."
    # 调用大模型生成改进建议
    suggestions = generate_improvement_suggestions(records)

    return render_template('result.html', records=records, suggestions=suggestions)

def generate_improvement_suggestions(records):
    # 整理历史记录文本
    history_text = "\n".join([f"Q: {rec.question}\nA: {rec.answer}\n" for rec in records])
    
    # 大模型 API 生成建议
    prompt = f"基于面试的历史记录，请给出有建设性的改进建议，分段说明，并将重要部分加粗:\n{history_text}"
    
    headers = {
        'Authorization': f'Bearer {chat_key}',
        'Content-Type': 'application/json'
    }
    data = {
        "model": chat_model,  # 指定模型版本
        "messages": [
            {"role": "user", "content": prompt}
        ]
    }
    
    response = requests.post(chat_url, headers=headers, json=data)
    result = response.json()

    suggestions = result.get('choices', [])[0].get('message', {}).get('content', "").strip()
    return suggestions

# 查看旧的历史记录
@app.route('/history', methods=['GET'])
def history():
    interviews = InterviewRecord.query.with_entities(InterviewRecord.interview_id).distinct().all()
    return render_template('history.html', interviews=interviews)

@app.route('/history/<interview_id>', methods=['GET'])
def view_history(interview_id):
    records = InterviewRecord.query.filter_by(interview_id=interview_id).all()
    return render_template('view_history.html', records=records)

# 返回到上一个页面的路由
@app.route('/go_back_result')
def go_back_result():
    return redirect(url_for('result'))
@app.route('/go_back_init')
def go_back_init():
    return redirect(url_for('init'))

# 模拟的数据
data = {
    "job_title": "前端工程师",
    "job_description": "1. 负责公司各类 Web 应用的前端开发工作，包括网页界面的设计、交互效果的实现以及页面性能的优化。\n2. 与产品团队、设计团队和后端开发团队紧密合作，确保前端开发工作与项目整体进度和需求保持一致。\n3. 运用 HTML5、CSS3 和 JavaScript 等前端技术，构建响应式、跨浏览器兼容的页面，提升用户体验。\n4. 参与前端技术选型和框架搭建，持续关注前端技术发展趋势，引入新的技术和工具提升开发效率。\n5. 对前端代码进行优化和维护，解决各种前端技术难题和浏览器兼容性问题，确保网站的稳定性和可靠性。",
    "resume_text": "具备扎实的前端开发技术，熟练掌握 HTML、CSS、JavaScript 等基础技术。精通 Vue.js 框架，曾参与多个基于 Vue.js 开发的大型项目，熟悉其组件化开发模式和生命周期管理。熟悉前端工程化，熟练使用 Webpack 等构建工具。有良好的代码规范意识，注重代码的可读性和可维护性。能够独立解决各种前端技术问题，如浏览器兼容性、性能优化等。在工作中注重团队协作，能够与不同岗位的人员高效沟通合作。具有较强的学习能力，持续关注前端技术的发展动态，不断提升自己的技术水平。"
}

@app.route('/get_data')
def get_data():
    return jsonify(data)

if __name__ == '__main__':
    with app.app_context():
        db.create_all()
    app.run(debug=True)