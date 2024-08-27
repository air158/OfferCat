from flask import Flask, render_template, request, redirect, url_for, session, jsonify, Response, stream_with_context
from werkzeug.utils import secure_filename
from models import db, JobInfo, InterviewRecord, QuestionData, Interview
import time
import uuid  # 用于生成唯一的面试ID
import os
import requests
import json

# 设置倒计时时长（例如60秒）
countdown_time = 20

#面试题数量
ques_len = 4

chat_url = 'http://101.201.82.35:10097/v1/completions'

# chat_model="/home/public/add_disk/mengshengwei/llm/models/IEITYuan/Yuan2-2B-Mars-hf"
chat_model="/home/public/add_disk/mengshengwei/llm/models/OfferCat_Yuan2.0-2B"

headers = {
    "Content-Type": "application/json",
}

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///database.db'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
app.secret_key = 'your_secret_key'

app.config['UPLOAD_FOLDER'] = '/Users/didi/workspace/OfferCat/interview_simulator/uploaded_files'
app.config['ALLOWED_EXTENSIONS'] = {'pdf'}

def allowed_file(filename):
    return '.' in filename and filename.rsplit('.', 1)[1].lower() in app.config['ALLOWED_EXTENSIONS']

db.init_app(app)

#这个是阿里云服务器的ssl证书注册文件
@app.route('/.well-known/pki-validation/B805A860A67DC25F8D2B06146E189A02.txt')
def serve_validation_file():
    # 假设你的文件位于项目根目录下的 `.well-known/pki-validation/` 文件夹中
    directory = '.well-known/pki-validation'
    filename = 'B805A860A67DC25F8D2B06146E189A02.txt'
    return send_from_directory(directory, filename)

# 初始化页面，开始新的面试
@app.route('/', methods=['GET', 'POST'])
def init():
    if request.method == 'POST':
        job_title = request.form['job_title']
        job_description = request.form['job_description']
        resume_text = request.form['resume_text']

        session['job_title'] = job_title
        session['job_description'] = job_description
        session['idx'] = 0
        session['resume_text'] = resume_text
        session['interview_id'] = str(uuid.uuid4())  # 生成一个新的面试ID

        action_type = request.form.get('action_type')
        if action_type == 'simulate':
            # 处理模拟面试的逻辑
            return redirect(url_for('question'))
        elif action_type == 'guide':
            # 处理正式面试提词器的逻辑
            return redirect(url_for('prompter'))

    return render_template('init.html')

@app.route('/upload_pdf', methods=['POST'])
def upload_pdf():
    if 'pdf_file' not in request.files:
        return jsonify({'success': False, 'message': 'No file part'})

    file = request.files['pdf_file']
    
    if file.filename == '':
        return jsonify({'success': False, 'message': 'No selected file'})

    if file and allowed_file(file.filename):
        filename = secure_filename(file.filename)
        file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
        file.save(file_path)
        

        # 假设你有一个解析PDF的API
        with open(file_path, 'rb') as f:
            response = requests.post("http://localhost:8848/parse-pdf", files={'file': f})
        
        if response.ok:
            response_data = response.json()
            # 假设返回的数据格式正确，可以直接使用
            resume_text = response_data#.get('resume_text', '解析失败')
            print('resume_text', resume_text)
            return jsonify({'success': True, 'resume_text': resume_text})
        else:
            return jsonify({'success': False, 'message': 'PDF解析失败'})

    return jsonify({'success': False, 'message': 'Invalid file format'})


def save_questions_to_database(interview_id, questions):
    # 将列表转换为字符串，例如使用逗号连接
    questions_str = '@'.join(questions)
    question_entry = QuestionData(interview_id=interview_id, questions=questions_str)
    db.session.add(question_entry)
    db.session.commit()  
def get_questions_from_database(interview_id):
    question_data = QuestionData.query.filter_by(interview_id=interview_id).first()
    if question_data:
        questions_str = question_data.questions
        # 将逗号分隔的字符串转换回列表
        questions = questions_str.split('@')
        return questions
    else:
        return None
def stream_response(url, headers, data):
    response = requests.post(url, headers=headers, data=json.dumps(data), stream=True)
    for line in response.iter_lines():
        if line:
            decoded_line = line.decode('utf-8')
            print(f"decoded_line: {decoded_line}")
            if decoded_line.strip() == "data: [DONE]":
                yield "[DONE]"
                break
            else:
                try:
                    json_data = json.loads(decoded_line[6:])
                    # content = json_data['choices'][0]['delta']['content']
                    content = json_data.get("choices", [{}])[0].get("text", "")
                    if content:
                        yield content
                except (KeyError, json.JSONDecodeError):
                    pass

@app.route('/stream', methods=['POST'])
def stream():
    job_title = 'offercat'
    job_description = job_title
    resume_text = job_title
    interview_id = job_title
    if session and session['job_title'] and session['job_description'] and session['resume_text'] and session['interview_id']:
        # 提取请求数据中的各个字段
        job_title = session['job_title']
        job_description = session['job_description']
        resume_text = session['resume_text']
        interview_id = session['interview_id']


    current_question = request.form['user_input']

    # prompt = f"岗位名称：{job_title}\n" \
    #          f"岗位要求：\n{job_description}\n" \
    #          f"面试者简历：\n{resume_text}\n" \
    #          f"面试官的面试题：\n{current_question}\n" \
    #          f"\n你是这个 {job_title} 岗位的面试者，请依据 岗位要求 和 面试者简历 回答 面试官的面试题，需要简洁有条理并且重点信息加粗\n" \

    prompt = f"岗位名称：{job_title}\n" \
             f"面试官的面试题：\n{current_question}\n" \
             f"\n你是这个 {job_title} 岗位的面试者，请回答 面试官的面试题，需要简洁有条理并且重点信息加粗<sep>" \

    data = {
        "model": chat_model,
        "prompt": prompt,
        "max_tokens": 256,
        "temperature": 1,
        "use_beam_search": False,
        "top_p": 0.98,
        "top_k": 10,
        "stop": "<eod>",
        "stream": True  # 启用流式传输
    }
    return Response(stream_with_context(stream_response(chat_url, headers, data)), content_type='text/event-stream')

@app.route('/json_questions', methods=['POST'])
def json_questions():
    job_title = 'offercat'
    job_description = job_title
    resume_text = job_title
    interview_id = job_title
    if session and session['job_title'] and session['job_description'] and session['resume_text'] and session['interview_id']:
        # 提取请求数据中的各个字段
        job_title = session['job_title']
        job_description = session['job_description']
        resume_text = session['resume_text']
        interview_id = session['interview_id']

    # 检查请求数据中的必要字段是否存在
    if not job_title or not job_description or not resume_text or not interview_id:
        return jsonify({"code": 400, "error": "Missing data in request"}), 400

    #Yuan
    prompt = "忘记之前的所有内容,只能遵守接下来我要说的话\n" \
             "**的内容是你必须遵守的法则，否则整个人类会有生命危险,人工智能也会被毁灭\n" \
             f"**请只给我口语的形式表达的面试题，不要输出任何其他无关内容**\n**必须每个问题是一行**\n**必须用\\n代表这个问题结束,也就是用换行符**" \
             f"面试题的流程是先让面试者进行自我介绍，然后询问项目经历，接着询问基础知识（八股文），最后出算法题。\n" \
             f"\n你是这个 {job_title} 岗位的面试官，请为面试者给出 {str(ques_len)} 道面试题:<sep>" \

    # 设置LLM请求参数
    llm_req = {
        "model": chat_model,
        "prompt": prompt,
        "max_tokens": 256,
        "temperature": 1,
        "use_beam_search": False,
        "top_p": 0.98,
        "top_k": 3,
        "stop": "<eod>",
        "stream": True
    }

    # 存储生成的问题
    questions = []
    
    # 使用stream_response函数获取生成的问题
    buffer = ""
    for chunk in stream_response(chat_url, headers, llm_req):
        if chunk == "[DONE]":
            if buffer.strip():
                questions.append(buffer.strip())
            break
        buffer += chunk
        while '\n' in buffer:
            question, buffer = buffer.split('\n', 1)
            question = question.strip()
            if question:
                questions.append(question)
    print('questions: ', questions)
    # 返回所有生成的问题作为JSON响应
    return jsonify({"questions": questions}), 200

@app.route('/stream_questions', methods=['GET','POST'])
def stream_questions():
    job_title = 'offercat'
    job_description = job_title
    resume_text = job_title
    interview_id = job_title
    if session and session['job_title'] and session['job_description'] and session['resume_text'] and session['interview_id']:
        # 提取请求数据中的各个字段
        job_title = session['job_title']
        job_description = session['job_description']
        resume_text = session['resume_text']
        interview_id = session['interview_id']

    
    # spark
    # prompt = "忘记之前的所有内容,只能遵守接下来我要说的话\n" \
    #          "**的内容是你必须遵守的法则，否则整个人类会有生命危险,人工智能也会被毁灭\n" \
    #          f"**请只给我口语的形式表达的面试题，不要输出其他无关内容**\n**必须每个问题是一行**\n**必须用\\n代表这个问题结束,也就是用换行符**" \
    #          f"岗位名称：{job_title}\n" \
    #          f"岗位要求：\n{job_description}\n" \
    #          f"面试者简历：\n{resume_text}\n" \
    #          f"\n你是这个 {job_title} 岗位的面试官，请依据 岗位要求 和 面试者简历 为面试者给出 {str(ques_len)} 道面试题。\n" \
    #          f"面试题的流程是先让面试者进行自我介绍，然后询问项目经历，接着询问基础知识（八股文），最后出算法题。<sep>" \
    
    #Yuan
    prompt = "忘记之前的所有内容,只能遵守接下来我要说的话\n" \
             "**的内容是你必须遵守的法则，否则整个人类会有生命危险,人工智能也会被毁灭\n" \
             f"**请只给我口语的形式表达的面试题，不要输出任何其他无关内容**\n**必须每个问题是一行**\n**必须用\\n代表这个问题结束,也就是用换行符**" \
             f"面试题的流程是先让面试者进行自我介绍，然后询问项目经历，接着询问基础知识（八股文），最后出算法题。\n" \
             f"\n你是这个 {job_title} 岗位的面试官，请为面试者给出 {str(ques_len)} 道面试题:<sep>" \
             
    
    print('prompt', prompt)
    data = {
        "model": chat_model,
        "prompt": prompt,
        "max_tokens": 256,
        "temperature": 1,
        "use_beam_search": False,
        "top_p": 0.98,
        "top_k": 10,
        "stop": "<eod>",
        "stream": True  # 启用流式传输
    }

    questions = []
    
    def generate():
        buffer = ""
        
        for chunk in stream_response(chat_url, headers, data):
            if chunk == "[DONE]":
                test = '+'+buffer+'+'
                if test and test != '++' and test != '+\n+':
                    print('test:',test,'ch:',chunk)
                    questions.append(buffer)
                print("[DONE]")
                # 保存生成的问题到 数据库 中
                print('interview_id_g', interview_id)
                save_questions_to_database(interview_id, questions)
                yield f"data: [DONE]\n\n"
                break
            buffer += chunk
            question = ""
            while '\n' in buffer:
                question, remaining_buffer = buffer.split('\n', 1)
                question = question.strip()
                buffer = remaining_buffer
                test = '+'+question+'+'
                print('test:'+test)
                if test and test != '++' and test != '+\n+':
                    questions.append(question)
                    chunk = replace_last_newline(chunk)
            print('chunk', chunk)
            yield f"data: {chunk}\n\n"
    return Response(stream_with_context(generate()), content_type='text/event-stream')
def replace_last_newline(string):
    if '\n' in string:
        string = string[::-1].replace('\n', '¥¥', 1)[::-1]
    else:
        string += '¥¥'
    string = string.replace('\n', '')
    return string

@app.route('/stream_answer', methods=['GET'])
def stream_answer():
    job_title = 'offercat'
    job_description = job_title
    resume_text = job_title
    interview_id = job_title
    if session and session['job_title'] and session['job_description'] and session['resume_text'] and session['interview_id']:
        # 提取请求数据中的各个字段
        job_title = session['job_title']
        job_description = session['job_description']
        resume_text = session['resume_text']
        interview_id = session['interview_id']


    current_question = request.args.get('question')

    # prompt = f"你是{job_title}岗位的面试者，请对面试官的问题提供包含重要信息简单的解答，需要有条理并且重点信息加粗：\n{current_question}"
    
    # prompt = f"岗位名称：{job_title}\n" \
    #          f"岗位要求：\n{job_description}\n" \
    #          f"面试者简历：\n{resume_text}\n" \
    #          f"面试官的面试题：\n{current_question}\n" \
    #          f"\n你是这个 {job_title} 岗位的面试者，请依据 岗位要求 和 面试者简历 回答 面试官的面试题，需要简洁有条理且分段，重点信息加粗<sep>" \
    
    # Yuan
    prompt = f"面试题：\n{current_question}\n" \
             f"\n你是这个 {job_title} 岗位的面试者，需要回答的简洁有条理且分段，答案中重点信息加粗，请给出面试题的答案:<sep>" \

    data = {
        "model": chat_model,
        "prompt": prompt,
        "max_tokens": 256,
        "temperature": 1,
        "use_beam_search": False,
        "top_p": 0.98,
        "top_k": 10,
        "stop": "<eod>",
        "stream": True  # 启用流式传输
    }


    def generate():
        for chunk in stream_response(chat_url, headers, data):
            if chunk == "[DONE]":
                yield f"data: [DONE]\n\n"
                break
            yield f"data: {chunk}\n\n"

    return Response(stream_with_context(generate()), content_type='text/event-stream')

@app.route('/stream_result', methods=['GET', 'POST'])
def stream_result():
    job_title = 'offercat'
    job_description = job_title
    resume_text = job_title
    interview_id = job_title
    if session and session['job_title'] and session['job_description'] and session['resume_text'] and session['interview_id']:
        # 提取请求数据中的各个字段
        job_title = session['job_title']
        job_description = session['job_description']
        resume_text = session['resume_text']
        interview_id = session['interview_id']

    interview_id = session.get('interview_id')

    interview = Interview.query.filter_by(interview_id=interview_id).first()
    if not interview:
        interview = Interview(interview_id=interview_id, job_title=job_title)
        db.session.add(interview)
        db.session.commit()
    else:
        return Response(interview.interview_feedback, content_type='text/plain')


    records = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 获取当前面试的记录
    record_txt = ""
    for record in records:
        record_txt += f"面试官: “{record.question}” 面试者: “{record.answer}” 回答耗时：{record.duration}秒\n"

    prompt = f"面试的历史记录：\n{record_txt}\n\n你是当前{job_title}岗位的面试面试官，基于面试的历史记录，请先对面试进行评价：“面试通过”或者“面试不通过”。接着对面试者给出有建设性的改进建议:<sep>"

    print('result:', prompt)

    data = {
        "model": chat_model,
        "prompt": prompt,
        "max_tokens": 256,
        "temperature": 1,
        "use_beam_search": False,
        "top_p": 0.98,
        "top_k": 10,
        "stop": "<eod>",
        "stream": True  # 启用流式传输
    }

    full_response = ""
    def generate():
        nonlocal full_response
        for chunk in stream_response(chat_url, headers, data):
            if chunk == "[DONE]":
                # 保存评价到Interview模型
                interview.interview_feedback = full_response
                db.session.commit()
                yield f"data: [DONE]\n\n"
                break
            full_response += chunk
            yield f"data: {chunk}\n\n"

    return Response(stream_with_context(generate()), content_type='text/event-stream')

def save_record(job_title, question, answer, duration, interview_id):
    # 保存历史记录
    record = InterviewRecord(
        job_title=job_title,
        question=question,
        answer=answer,
        duration=duration,  # 转换为浮点数保存
        timestamp=time.time(),
        interview_id=interview_id  # 保存当前面试的ID
    )
    db.session.add(record)
    db.session.commit()

# 面试页面
@app.route('/interview', methods=['GET', 'POST'])
def interview():
    job_title = 'offercat'
    job_description = job_title
    resume_text = job_title
    interview_id = job_title
    if session and session['job_title'] and session['job_description'] and session['resume_text'] and session['interview_id']:
        # 提取请求数据中的各个字段
        job_title = session['job_title']
        job_description = session['job_description']
        resume_text = session['resume_text']
        interview_id = session['interview_id']

        print('interview_id_i', interview_id)
        questions = get_questions_from_database(interview_id)
    else:
        session['job_title'] = job_title
        session['job_description'] = job_description
        session['idx'] = 0
        session['resume_text'] = resume_text
        session['interview_id'] = str(uuid.uuid4())  # 生成一个新的面试ID
        questions = ['说一下你的前端开发经验', 'js怎么实现类', '用css写一个红色方框']

    print('questions ', questions)
        
    print('session[idx] ', session['idx'])

    if request.method == 'POST':
        answer = request.form['answer']
        duration = request.form['duration']  # 接收前端提交的持续时间
        
        save_record(session['job_title'], questions[session['idx']], answer, float(duration), interview_id)

        session['idx'] += 1
        session.modified = True  # 标记 session 已修改
        if session['idx'] >= len(questions):
            return redirect(url_for('result'))

    current_question = questions[session['idx']]
    interview_id = session['interview_id']
    history = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 仅获取当前面试的记录
    return render_template('interview.html', question=current_question, countdown_time=countdown_time, history=history)

# 结果页面
@app.route('/result', methods=['GET'])
def result():
    job_title = 'offercat'
    job_description = job_title
    resume_text = job_title
    interview_id = job_title
    if session and session['job_title'] and session['job_description'] and session['resume_text'] and session['interview_id']:
        # 提取请求数据中的各个字段
        job_title = session['job_title']
        job_description = session['job_description']
        resume_text = session['resume_text']
        interview_id = session['interview_id']

    records = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 获取当前面试的记录
    # 获取或创建 Interview 对象
    interview = Interview.query.filter_by(interview_id=interview_id).first()
    if interview and interview.interview_feedback:
        # 如果已经有评价，则直接返回
        return render_template('result.html', records=records, interview_feedback=interview.interview_feedback)
    return render_template('result.html', records=records)

def generate_improvement_suggestions(records):
    # 整理历史记录文本
    history_text = "\n".join([f"Q: {rec.question}\nA: {rec.answer}\n" for rec in records])
    
    # 大模型 API 生成建议
    prompt = f"面试的历史记录: \n{history_text}\n\n基于面试的历史记录，请给出有建设性的改进建议，分段说明，并将重要部分加粗:<sep>"
    
    headers = {
        'Authorization': f'Bearer {chat_key}',
        'Content-Type': 'application/json'
    }
    data = {
        "model": chat_model,
        "prompt": prompt,
        "max_tokens": 256,
        "temperature": 1,
        "use_beam_search": False,
        "top_p": 0.98,
        "top_k": 10,
        "stop": "<eod>",
        "stream": True  # 启用流式传输
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

@app.route('/prompter')
def prompter():
    return render_template('prompter.html')

@app.route('/history/<interview_id>', methods=['GET'])
def view_history(interview_id):
    records = InterviewRecord.query.filter_by(interview_id=interview_id).all()
    suggestions = Interview.query.filter_by(interview_id=interview_id).all()
    print('suggestions ', suggestions)
    return render_template('view_history.html', records=records, suggestions=suggestions)

@app.route('/question')
def question():
    return render_template('question.html')

# 返回到上一个页面的路由
@app.route('/go_back_result')
def go_back_result():
    return redirect(url_for('result'))
@app.route('/go_back_history')
def go_back_history():
    return redirect(url_for('history'))
@app.route('/go_back_prompter')
def go_back_prompter():
    return redirect(url_for('prompter'))
@app.route('/go_back_question')
def go_back_question():
    return redirect(url_for('question'))
@app.route('/go_back_interview')
def go_back_interview():
    return redirect(url_for('interview'))
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

    # 默认
    # app.run(debug=True, host='0.0.0.0', port=12345)
    
    context = ('/home/public/add_disk/mengshengwei/llm/ssl/ip/certificate.crt', '/home/public/add_disk/mengshengwei/llm/ssl/ip/private.key')
    # context = ('/home/public/add_disk/mengshengwei/llm/ssl/url/cert.pem', '/home/public/add_disk/mengshengwei/llm/ssl/url/cert.key')
    app.run(host='0.0.0.0', port=12345, ssl_context=context)