from flask import Flask, render_template, request, redirect, url_for, session
from models import db, JobInfo, InterviewRecord
import time
import uuid  # 用于生成唯一的面试ID

# 设置倒计时时长（例如60秒）
countdown_time = 60

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
        # 模拟大模型生成问题
        questions = [
            "Can you explain a challenging project you've worked on?",
            "How do you handle tight deadlines?"
        ]
        session['questions'] = questions
        return redirect(url_for('interview'))

    return render_template('init.html')

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
    # 模拟大模型生成改进建议
    suggestions = "Consider providing more specific examples and quantify your achievements when possible."

    return render_template('result.html', records=records, suggestions=suggestions)

# 查看旧的历史记录
@app.route('/history', methods=['GET'])
def history():
    interviews = InterviewRecord.query.with_entities(InterviewRecord.interview_id).distinct().all()
    return render_template('history.html', interviews=interviews)

@app.route('/history/<interview_id>', methods=['GET'])
def view_history(interview_id):
    records = InterviewRecord.query.filter_by(interview_id=interview_id).all()
    return render_template('view_history.html', records=records)

if __name__ == '__main__':
    with app.app_context():
        db.create_all()
    app.run(debug=True)