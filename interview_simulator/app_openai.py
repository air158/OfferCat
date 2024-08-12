from flask import Flask, render_template, request, redirect, url_for, session
from models import db, JobInfo, InterviewRecord
from datetime import datetime
import requests  # 使用 requests 库来发送 HTTP 请求

import os

chat_api_key = os.getenv('SPARK_API_KEY')
chat_api_secret = os.getenv('SPARK_API_SECRET')

chat_key = f'{chat_api_key}:{chat_api_secret}'

chat_url = 'https://spark-api-open.xf-yun.com/v1/chat/completions'

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///database.db'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
app.secret_key = 'your_secret_key'

db.init_app(app)

# 初始化数据库
@app.before_first_request
def create_tables():
    db.create_all()

# 初始化页面
@app.route('/', methods=['GET', 'POST'])
def init():
    if request.method == 'POST':
        job_title = request.form['job_title']
        job_description = request.form['job_description']
        resume_text = request.form['resume_text']

        session['job_title'] = job_title
        session['job_description'] = job_description
        session['resume_text'] = resume_text

        # 调用大模型生成问题
        questions = generate_interview_questions(job_title, job_description, resume_text)
        session['questions'] = questions
        return redirect(url_for('interview'))

    return render_template('init.html')

def generate_interview_questions(job_title, job_description, resume_text):
    # 使用星火大模型生成面试问题
    prompt = f"Generate interview questions for the following job position: {job_title}. \
              Job description: {job_description}. Candidate's resume: {resume_text}."

    headers = {
        'Authorization': f'Bearer {chat_key}',
        'Content-Type': 'application/json'
    }
    data = {
        "model": "generalv3.5",  # 指定模型版本
        "messages": [
            {"role": "user", "content": prompt}
        ]
    }
    
    response = requests.post(chat_url, headers=headers, json=data)
    result = response.json()

    # 解析生成的问题
    questions = result.get('choices', [])[0].get('message', {}).get('content', "").strip().split('\n')
    return [q for q in questions if q]

# 面试页面
@app.route('/interview', methods=['GET', 'POST'])
def interview():
    if 'questions' not in session:
        return redirect(url_for('init'))

    if request.method == 'POST':
        answer = request.form['answer']
        question = session['questions'].pop(0)
        start_time = session.get('start_time', datetime.now())
        end_time = datetime.now()
        duration = (end_time - start_time).total_seconds()

        # 保存历史记录
        record = InterviewRecord(
            job_title=session['job_title'],
            question=question,
            answer=answer,
            duration=duration,
            timestamp=end_time
        )
        db.session.add(record)
        db.session.commit()

        session['start_time'] = datetime.now()

        if not session['questions']:
            return redirect(url_for('result'))

    current_question = session['questions'][0] if session['questions'] else None
    history = InterviewRecord.query.filter_by(job_title=session['job_title']).all()
    return render_template('interview.html', question=current_question, history=history)

# 结果页面
@app.route('/result', methods=['GET'])
def result():
    records = InterviewRecord.query.filter_by(job_title=session.get('job_title')).all()
    
    # 调用大模型生成改进建议
    suggestions = generate_improvement_suggestions(records)
    
    return render_template('result.html', records=records, suggestions=suggestions)

def generate_improvement_suggestions(records):
    # 整理历史记录文本
    history_text = "\n".join([f"Q: {rec.question}\nA: {rec.answer}\n" for rec in records])
    
    # 星火大模型 API 生成建议
    prompt = f"Based on the following interview transcript, provide feedback and suggestions for improvement:\n{history_text}"
    
    headers = {
        'Authorization': f'Bearer {chat_key}',
        'Content-Type': 'application/json'
    }
    data = {
        "model": "generalv3.5",  # 指定模型版本
        "messages": [
            {"role": "user", "content": prompt}
        ]
    }
    
    response = requests.post(chat_url, headers=headers, json=data)
    result = response.json()

    suggestions = result.get('choices', [])[0].get('message', {}).get('content', "").strip()
    return suggestions

if __name__ == '__main__':
    app.run(debug=True)