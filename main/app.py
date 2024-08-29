from datetime import datetime, timedelta
from flask import Flask, render_template, request, redirect, url_for, session, jsonify, Response, stream_with_context
from werkzeug.utils import secure_filename
from models import db, JobInfo, InterviewRecord, QuestionData, Interview, User, RunnigInterview
import time
import uuid  # 用于生成唯一的面试ID
import os
import requests
import json
from functools import wraps

# 设置倒计时时长（例如60秒）
countdown_time = 20

#面试题数量
ques_len = 4

chat_url = 'http://101.201.82.35:10097/v1/completions'
backend_url = 'http://116.198.207.159:12345/api'

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

# 与后端通信
def send_request(func, headers, payload, method="POST"):
    register_url = backend_url + func

    payload = json.dumps(payload)

    response = requests.request(method, register_url, headers=headers, data=payload)
    response_json = response.json() 
    code = response_json.get("code")
    data = response_json.get("data")
    message = response_json.get("massage")
    return code, data, message


# 鉴定验证码
def verify_code(token, redeem_code):

    headers = {
        'Authorization': token,
        'Content-Type': 'application/json'
    }
    payload = {
        "code": redeem_code
    }
    code, data, message = send_request("/redeem-code/verify", headers, payload)
    
    if code == 200:
        # 更新session
        type = data.get("tag")
        user_name = data.get("username")
        interviewPoint = 0
        if type == "interviewPoint:1h":
            interviewPoint = 60
        elif type == "interviewPoint:2h":
            interviewPoint = 120
        print(f'user_name:{user_name}, interviewPoint:{interviewPoint}')
        return True, message
    else:
        return False, message

# 假设这是您实现的获取用户信息的函数
def get_user_info(token):
    # 实现从token获取用户名和面试点数的逻辑
    # 返回 (username, interview_points)

    headers = {
        'Authorization': token,
    }
    payload={}
    code, data, message = send_request("/profile", headers, payload, method="GET")

    user_name = None
    interview_points = 0

    if code == 200:
        # 更新session
        user_name = data.get("username")
        user_id = data.get("uid")
        session['user_id'] = user_id
        session['user_name'] = user_name

        # 确保session被标记为已修改
        session.modified = True
        # interview_points = data.get("interviewPoint")

    return code == 200, user_name, user_id, interview_points, message


# 注册面试
def register_interview(token, job_title="前端工程师", job_description="1. 负责公司各类 Web 应用的前端开发工作，包括网页界面的设计、交互效果的实现以及页面性能的优化。\n2. 运用 HTML5、CSS3 和 JavaScript 等前端技术，构建响应式、跨浏览器兼容的页面，提升用户体验。\n", company="OfferCat", business="Technology", location="Mountain View", progress="Second interview", language="中文", interview_style="结构化面试", interview_role="面试官", time_limit_per_question=300, interview_type="模拟面试"):

    if 'init' in session:
        if 'job_title' in session:
            job_title = session['job_title']
        if 'job_description' in session:
            job_description = session['job_description']
        if 'interview_type' in session:
            interview_type = session['interview_type']

    headers = {
        'Authorization': token,
        'Content-Type': 'application/json'
    }
    payload = {
       "job_title": job_title,
       "job_description": job_description,
       "company": company,
       "business": business,
       "location": location,
       "progress": progress,
       "language": language,
       "interview_style": interview_style,
       "interview_role": interview_role,
       "time_limit_per_question": time_limit_per_question,
       "type": interview_type
    }

    code, data, message = send_request("/interview/register", headers, payload)
    if code == 200:
        print('register_interview data:', data)
        data_interview = data.get("interview")
        interview_id = data_interview.get("id")
        user_id = data_interview.get("user_id")
        closed = data_interview.get("closed")
        running_interview = RunnigInterview(user_id=user_id, interview_id=interview_id)
        db.session.add(running_interview)
        db.session.commit()

        session['job_title'] = job_title
        session['job_description'] = job_description
        session['idx'] = 0
        session['interview_id'] = interview_id

        # 确保session被标记为已修改
        session.modified = True
    else:
        print('注册面试失败: ' + message)
    return code == 200, data, message

# 关闭面试
def close_interview(token, interview_id):
    headers = {
        'Authorization': token,
        'Content-Type': 'application/json'
    }
    payload = {
        "interview_id": interview_id
    }
    code, data, message = send_request("/interview/close", headers, payload)
    if code == 200:
        print('close_interview data:', data)
        # TODO: 更新数据库用户点数
        cost_point = data.get("cost_point")
        cost_type = data.get("cost_type")
        left_point = data.get("left_point")
        time_spent = data.get("time_spent")
        print(f'cost_point:{cost_point}, cost_type:{cost_type}, left_point:{left_point}, time_spent:{time_spent}')

        session['left_point'] = left_point
        # 确保session被标记为已修改
        session.modified = True

        running_interview = RunnigInterview.query.filter_by(interview_id=interview_id).first()
        if running_interview:
            db.session.delete(running_interview)
            db.session.commit()
    else:
        print('关闭面试失败: ' + message)
    return code, data, message
# 关闭所有面试
def close_all_interviews(token):
    # 从数据库中获取所有正在进行的面试
    running_interviews = RunnigInterview.query.all()
    #如果没有面试，则返回
    if not running_interviews:
        print('没有正在进行的面试')
        return
    
    for interview in running_interviews:
        # 关闭每个面试
        code, data, message = close_interview(token, interview.interview_id)
        if code != 200:
            print(f'关闭面试 {interview.interview_id} 失败: {message}')
        else:
            print(f'成功关闭面试 {interview.interview_id}')
            # 成功关闭的面试需要删除编号
            db.session.delete(interview)
    
    # 确保所有更改都被提交到数据库
    db.session.commit()
    
    print('所有面试已关闭并删除')



# 用注册面试和关闭面试获取left_point
def get_left_point(token):
    # 注册一个新的面试
    reged, data, message = register_interview(token)
    if not reged:
        print('注册面试失败:', message)
        return None
    
    interview_id = data.get('interview', {}).get('id')
    if not interview_id:
        print('获取面试ID失败')
        return None
    
    # 立即关闭这个面试
    code, close_data, close_message = close_interview(token, interview_id)
    if code != 200:
        print('关闭面试失败:', close_message)
        return None
    
    left_point = close_data.get('left_point')
    return left_point


def clear_session():
    session['job_title'] = None
    session['job_description'] = None
    session['idx'] = 0
    session['resume_text'] = None
    session['init'] = False
    session['interview_type'] = None
    session['return_to'] = None
    # 确保session被标记为已修改
    session.modified = True

def update_session_with_token(token):
    if token:
        # 如果有 token，更新 session
        session['token'] = token
        # 获取并更新 left_point
        left_point = get_left_point(token)
        session['left_point'] = left_point
        print('Updated session[left_point]:', session['left_point'])
        
        # 获取用户信息并更新 session
        loged, username, user_id, points, message = get_user_info(token)
        if loged:
            session['username'] = username
            session['user_id'] = user_id
        
        # 确保 session 被标记为已修改
        session.modified = True

# 修改装饰器来检查面试点数
def check_interview_points(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):

        token = request.args.get('question')
        update_session_with_token(token)
        
        if 'token' not in session:
            return redirect(url_for('login'))
        
        token = session['token']
        
        loged, username, user_id, points, message = get_user_info(token)
        if not loged:
            return redirect(url_for('login'))
        # 如果当前用户没有正在进行的面试则注册面试
        running_interview = RunnigInterview.query.filter_by(user_id=user_id).first()
        if not running_interview:
            reged, data, message = register_interview(token=token)
            # 注册失败，需要充值
            if not reged:
                session['return_to'] = request.url  # 保存当前URL
                return redirect(url_for('recharge'))
        else:
            # 获取面试编号
            interview_id = running_interview.interview_id
            session['interview_id'] = interview_id
            print('interview_id:', interview_id)

        # 确保session被标记为已修改
        session.modified = True
        
        return f(*args, **kwargs)
    return decorated_function

# 登录路由
@app.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        email = request.form['email']
        password = request.form['password']
        
        # 准备请求数据
        login_url = backend_url + "/login"
        payload = json.dumps({
            "email": email,
            "password": password
        })
        headers = {
            'Content-Type': 'application/json'
        }
        
        try:
            response = requests.post(login_url, headers=headers, data=payload)
            response.raise_for_status()  # 检查HTTP错误
            
            if response.text:  # 检查响应是否为空
                response_data = response.json()
                if response_data['code'] == 200:
                    session['token'] = response_data['data']['token']
                    session['username'] = response_data['data']['username']
                    left_point = get_left_point(session['token'])
                    session['left_point'] = left_point
                    print('session[left_point]:', session['left_point'])
                    # 确保session被标记为已修改
                    session.modified = True
                    close_all_interviews(session['token'])
                    return redirect(url_for('init'))
                else:
                    return render_template('login.html', error=response_data.get('massage', '登录失败'))
            else:
                return render_template('login.html', error='服务器返回空响应')
        except requests.exceptions.RequestException as e:
            print(f"请求错误: {e}")
            print(f"响应内容: {response.text}")  # 打印响应内容以进行调试
            return render_template('login.html', error='服务器连接错误')
        except ValueError as e:  # JSON解码错误
            print(f"JSON解码错误: {e}")
            print(f"响应内容: {response.text}")  # 打印响应内容以进行调试
            return render_template('login.html', error='服务器返回无效数据')
    
    return render_template('login.html')


# 充值路由
@app.route('/recharge', methods=['GET', 'POST'])
def recharge():
    if 'token' not in session:
        return redirect(url_for('login'))
    token = session['token']
    if request.method == 'POST':
        redeem_code = request.form['redeem_code']
        success, message = verify_code(token, redeem_code)
        if success:
            return_to = session.pop('return_to', url_for('init'))
            return redirect(return_to)
        else:
            return render_template('recharge.html', error=message)
    return render_template('recharge.html')


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
    # 清除session
    clear_session()
    if 'token' in session and session['token']:
        close_all_interviews(session['token'])

    if request.method == 'POST':
        job_title = request.form['job_title']
        job_description = request.form['job_description']
        resume_text = request.form['resume_text']

        session['job_title'] = job_title
        session['job_description'] = job_description
        session['resume_text'] = resume_text
        session['init'] = True
        # session['interview_id'] = str(uuid.uuid4())  # 生成一个新的面试ID

        action_type = request.form.get('action_type')
        session['interview_type'] = action_type

        # 确保session被标记为已修改
        session.modified = True
        if action_type == 'simulate':
            # 处理模拟面试的逻辑
            return redirect(url_for('question'))
        elif action_type == 'guide':
            # 处理正式面试提词器的逻辑
            return redirect(url_for('prompter'))

    if 'left_point' in session:
        left_point = session['left_point']
    else:
        left_point = 0
    print('left_point:', left_point)
    return render_template('init.html', left_point=left_point)

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
    job_title = session['job_title']
    # job_description = session['job_description']
    # resume_text = session['resume_text']

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

@app.route('/stream_questions', methods=['GET','POST'])
def stream_questions():
    job_title = session['job_title']
    # job_description = session['job_description']
    # resume_text = session['resume_text']
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
             "## 要求\n" \
             "**的内容是你必须遵守的法则，否则整个人类会有生命危险,人工智能也会被毁灭\n" \
             "**请只给我口语的形式表达的面试题，不要输出任何其他无关内容**\n" \
             "**必须每个问题是一行**\n" \
             "**必须用\\n代表这个问题结束,也就是用换行符**" \
             f"面试题的流程是先让面试者进行自我介绍，然后询问项目经历，接着询问基础知识（八股文），最后出算法题。\n" \
             "## 例子：\n" \
             "请简单介绍一下你自己,以及你为什么选择应聘我们公司的前端工程师岗位?\\n" \
             "能否详细描述一下你最近参与的一个前端项目?包括项目背景、你的具体职责、遇到的技术难点以及如何解决的。\\n" \
             "请解释一下 JavaScript 中的闭包(Closure)是什么,以及它的作用和使用场景。\\n" \
             "假设我们需要实现一个函数,找出一个无序数组中的第K大元素,你会如何设计这个算法?请简要描述你的思路。\\n" \
             "## 指令：\n" \
             f"你是这个 {job_title} 岗位的面试官，请为面试者给出 {str(ques_len)} 道面试题:<sep>" \
             
    
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
    job_title = session['job_title']
    # job_description = session['job_description']
    # resume_text = session['resume_text']

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
    job_title = session['job_title']
    interview_id = session['interview_id']

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

    prompt = f"## 面试的历史记录：\n{record_txt}\n\n" \
             f"## 要求\n 你是当前{job_title}岗位的面试面试官，基于面试的历史记录，请先对面试进行评价：“面试通过”或者“面试不通过”,接着给出有建设性的改进建议\n\n" \
             f"## 例子1\n```面试通过```\n回答的很好，面试者基础知识掌握的很好，项目经验也很丰富，可以录用\n" \
             f"## 例子2\n```面试不通过```\n回答的不好，面试者基础知识掌握的不好，项目经验也很一般，建议不通过\n" \
             f"## 指令：\n你是当前{job_title}岗位的面试面试官,请基于面试的历史记录对面试进行评价,接着给出有建设性的改进建议:<sep>"

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
@check_interview_points
def interview():
    if 'interview_id' not in session:
        return redirect(url_for('login'))
    interview_id = session['interview_id']  # 获取当前面试的ID
    print('interview_id_i', interview_id)
    questions = get_questions_from_database(interview_id)

    if request.method == 'POST':
        answer = request.form['answer']
        duration = request.form['duration']  # 接收前端提交的持续时间
        
        save_record(session['job_title'], questions[session['idx']], answer, float(duration), interview_id)

        session['idx'] += 1
        session.modified = True  # 标记 session 已修改
        if session['idx'] >= len(questions):
            return redirect(url_for('result'))

    current_question = questions[session['idx']]
    history = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 仅获取当前面试的记录
    return render_template('interview.html', question=current_question, countdown_time=countdown_time, history=history)

@app.route('/prompter')
@check_interview_points
def prompter():
    return render_template('prompter.html')

# 结果页面
@app.route('/result', methods=['GET'])
def result():
    if 'interview_id' not in session:
        return redirect(url_for('login'))
    interview_id = session['interview_id']
    token = session['token']
    # 关闭面试
    code, data, message = close_interview(token, interview_id)
    if code != 200:
        #需要重试
        print(f'{interview_id} 需要重试 {code} {data} {message}')
        # return redirect(url_for('interview'))
    else:
        print('关闭面试成功: ' + message)
        # 更新用户点数信息
        cost_point = data.get('cost_point')
        left_point = data.get('left_point')
        # session['interview_points'] = left_point
        print(f'面试已结束。本次消耗{cost_point}点，剩余{left_point}点。')
    records = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 获取当前面试的记录
    # 获取或创建 Interview 对象
    interview = Interview.query.filter_by(interview_id=interview_id).first()

    if interview and interview.interview_feedback:
        # 如果已经有评价，则直接返回
        return render_template('result.html', records=records, interview_feedback=interview.interview_feedback, message=message)
    return render_template('result.html', records=records)

# 查看旧的历史记录
@app.route('/history', methods=['GET'])
def history():
    interviews = InterviewRecord.query.with_entities(InterviewRecord.interview_id).distinct().all()
    return render_template('history.html', interviews=interviews)

@app.route('/history/<interview_id>', methods=['GET'])
def view_history(interview_id):
    records = InterviewRecord.query.filter_by(interview_id=interview_id).all()
    suggestions = Interview.query.filter_by(interview_id=interview_id).all()
    print('suggestions ', suggestions)
    return render_template('view_history.html', records=records, suggestions=suggestions)

@app.route('/question')
@check_interview_points
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
@app.route('/go_back_login')
def go_back_login():
    return redirect(url_for('login'))
@app.route('/go_back_recharge')
def go_back_recharge():
    return redirect(url_for('recharge'))
@app.route('/go_back_interview')
def go_back_interview():
    return redirect(url_for('interview'))
@app.route('/go_back_init')
def go_back_init():
    # return redirect('http://117.72.35.68:3200/')
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
    # 清除所有 session
    first_request = True
    @app.before_request
    def clear_session_on_first_request():
        global first_request
        if first_request:
            session.clear()
            first_request = False
    # app.run(debug=True, host='0.0.0.0', port=12345)
    
    # ssl
    context = ('/home/public/add_disk/mengshengwei/llm/ssl/ip/certificate.crt', '/home/public/add_disk/mengshengwei/llm/ssl/ip/private.key')
    # context = ('/home/public/add_disk/mengshengwei/llm/ssl/url/cert.pem', '/home/public/add_disk/mengshengwei/llm/ssl/url/cert.key')
    app.run(host='0.0.0.0', port=12345, ssl_context=context)