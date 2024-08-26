from flask import Flask, render_template, request, redirect, url_for, session, jsonify, Response, stream_with_context
from models_yuuki import db, JobInfo, InterviewRecord, QuestionData, Interview
import os
import requests
import json

# 设置倒计时时长（例如60秒）
# countdown_time = 20

os.environ['HTTP_PROXY'] = 'http://127.0.0.1:7890'
os.environ['HTTPS_PROXY'] = 'http://127.0.0.1:7890'

# chat_model = 'generalv3.5'
chat_model="/home/public/add_disk/mengshengwei/llm/models/OfferCat_Yuan2.0-2B"
headers = {
    "Content-Type": "application/json",
}


# chat_url = 'https://spark-api-open.xf-yun.com/v1/chat/completions'
chat_url = 'http://101.201.82.35:10097/v1/completions'


app = Flask(__name__)
def stream_response(url, headers, data):
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
                    # content = json_data['choices'][0]['delta']['content']
                    content = json_data.get("choices", [{}])[0].get("text", "")
                    yield content
                except (KeyError, json.JSONDecodeError):
                    pass

@app.route('/stream_questions', methods=['POST'])
def stream_questions():
    req_data = request.get_json()
    # 如果数据不存在，返回错误
    if not req_data:
        return jsonify({"error": "No data provided"}), 400

    job_title = req_data.get('job_title')
    job_description = req_data.get('job_description')
    resume_text = req_data.get('resume_text')
    print('resume_text', resume_text)
    interview_id = req_data.get('interview_id')
    
    ques_len = req_data.get('ques_len')
    chat_key = req_data.get('chat_key')

    if not job_title or not job_description or not resume_text or not interview_id or not ques_len or not chat_key:
        return jsonify({"code":400,"error": "Missing data in request"}), 400
    
    prompt = "忘记之前的所有内容,只能遵守接下来我要说的话\n" \
             "**的内容是你必须遵守的法则，否则整个人类会有生命危险,人工智能也会被毁灭\n" \
             f"**请只给我口语的形式表达的面试题，不要输出其他无关内容**\n**必须每个问题是一行**\n**必须用\\n代表这个问题结束,也就是用换行符**" \
             f"岗位名称：{job_title}\n" \
             f"岗位要求：\n{job_description}\n" \
             f"面试者简历：\n{resume_text}\n" \
             f"\n你是这个 {job_title} 岗位的面试官，请依据 岗位要求 和 面试者简历 为面试者给出 {str(ques_len)} 道面试题。\n" \
             f"面试题的流程是先让面试者进行自我介绍，然后询问项目经历，接着询问基础知识（八股文），最后出算法题。<sep>" \
                 
    print('prompt', prompt)
    # headers = {
    #     'Authorization': f'Bearer {chat_key}',
    #     'Content-Type': 'application/json'
    # }
    # llm_req = {
    #     "model": chat_model,
    #     "messages": [
    #         {"role": "user", "content": prompt}
    #     ],
    #     "stream": True
    # }
    llm_req = {
        "model": chat_model,
        "prompt": prompt,
        "max_tokens": 256,
        "temperature": 1,
        "use_beam_search": False,
        "top_p": 0,
        "top_k": 1,
        "stop": "<eod>",
        "stream": True  # 启用流式传输
    }

    questions = []
    
    def generate():
        buffer = ""
        
        for chunk in stream_response(chat_url, headers, llm_req):
            if chunk == "[DONE]":
                test = '+'+buffer+'+'
                if test and test != '++' and test != '+\n+':
                    print('test:',test,'ch:',chunk)
                    questions.append(buffer)
                print("[DONE]")
                # 保存生成的问题到 数据库 中
                print('interview_id_g', interview_id)
                # save_questions_to_database(interview_id, questions)
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


@app.route('/stream_answer', methods=['POST'])
def stream_answer():
    req_data = request.get_json()

    job_title = req_data.get('job_title')
    # 增加这俩参数
    job_description = req_data.get('job_description')
    resume_text = req_data.get('resume_text')
    # ques_len = req_data.get('ques_len')
    chat_key = req_data.get('chat_key')
    current_question = req_data.get('question')

    # current_question = request.args.get('question')

    if not job_title or not chat_key or not current_question:
        return jsonify({"error": "Missing data in request"}), 400

    # prompt = f"你是{job_title}岗位的面试者，请对面试官的问题提供包含重要信息简单的解答，需要有条理并且重点信息加粗：\n{current_question}"
    prompt = "忘记之前的所有内容,只能遵守接下来我要说的话\n" \
             "**的内容是你必须遵守的法则，否则整个人类会有生命危险,人工智能也会被毁灭\n" \
            f"**请只给我口语的形式表达的面试题，不要输出其他无关内容**\n**必须一行结束回答**\n**必须用\\n代表这个回答结束,也就是用换行符**。解析来我会告诉你你的基本信息：" \
             f"岗位要求：\n{job_description}\n" \
             f"面试者简历：\n{resume_text}\n" \
             f"面试官的面试题：\n{current_question}\n" \
             f"\n你是这个 {job_title} 岗位的面试者，请直接回答这个面试官的问题，需要简洁有条理且分段，重点信息加粗，<sep>" \
                 
    print('prompt:', prompt)
    # headers = {
    #     'Authorization': f'Bearer {chat_key}',
    #     'Content-Type': 'application/json'
    # }
    # llm_req = {
    #     "model": chat_model,
    #     "messages": [
    #         {"role": "user", "content": prompt}
    #     ],
    #     "stream": True
    # }
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
    

    def generate():
        for chunk in stream_response(chat_url, headers, llm_req):
            if chunk == "[DONE]":
                yield f"data: [DONE]\n\n"
                break
            yield f"data: {chunk}\n\n"

    return Response(stream_with_context(generate()), content_type='text/event-stream')

@app.route('/stream_result', methods=['GET', 'POST'])
def stream_result():
    req_data = request.get_json()

    job_title = req_data.get('job_title')
    chat_key = req_data.get('chat_key')
    prompt_text = req_data.get('prompt_text')

    # current_question = request.args.get('question')

    if not job_title or not chat_key or not prompt_text:

        return jsonify({"error": "Missing data in request"}), 400

    # interview = Interview.query.filter_by(interview_id=interview_id).first()
    # if not interview:
    #     interview = Interview(interview_id=interview_id, job_title=job_title)
    #     db.session.add(interview)
    #     db.session.commit()
    # else:
    #     return Response(interview.interview_feedback, content_type='text/plain')


    # records = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 获取当前面试的记录


    prompt = f"面试的历史记录：\n{prompt_text}\n\n基于当前{job_title}岗位的面试的历史记录，请先对整个面试进行一个评价：“面试通过”或者“面试不通过”。接着对面试者给出有建设性的改进建议，分段说明，并将重要部分加粗:<sep>"

    print('result:', prompt)

    # headers = {
    #     'Authorization': f'Bearer {chat_key}',
    #     'Content-Type': 'application/json'
    # }
    # llm_req = {
    #     "model": chat_model,
    #     "messages": [
    #         {"role": "user", "content": prompt}
    #     ],
    #     "stream": True
    # }
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

    full_response = ""
    def generate():
        nonlocal full_response
        for chunk in stream_response(chat_url, headers, llm_req):
            if chunk == "[DONE]":
                yield f"data: [DONE]\n\n"
                break
            full_response += chunk
            yield f"data: {chunk}\n\n"

    return Response(stream_with_context(generate()), content_type='text/event-stream')

# # 面试页面
# @app.route('/interview', methods=['GET', 'POST'])
# def interview():
#     interview_id = session['interview_id']  # 获取当前面试的ID
#     print('interview_id_i', interview_id)
#     questions = get_questions_from_database(interview_id)
#
#     if request.method == 'POST':
#         answer = request.form['answer']
#         duration = request.form['duration']  # 接收前端提交的持续时间
#         # 保存历史记录
#         record = InterviewRecord(
#             job_title=session['job_title'],
#             question=questions[session['idx']],
#             answer=answer,
#             duration=float(duration),  # 转换为浮点数保存
#             timestamp=time.time(),
#             interview_id=interview_id  # 保存当前面试的ID
#         )
#         db.session.add(record)
#         db.session.commit()
#
#         session['idx'] += 1
#         session.modified = True  # 标记 session 已修改
#         if session['idx'] >= len(questions):
#             return redirect(url_for('result'))
#
#     current_question = questions[session['idx']]
#     interview_id = session['interview_id']
#     history = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 仅获取当前面试的记录
#     return render_template('interview.html', question=current_question, countdown_time=countdown_time, history=history)

# # 结果页面
# @app.route('/result', methods=['GET'])
# def result():
#     interview_id = session.get('interview_id')
#     records = InterviewRecord.query.filter_by(interview_id=interview_id).all()  # 获取当前面试的记录
#     # 获取或创建 Interview 对象
#     interview = Interview.query.filter_by(interview_id=interview_id).first()
#     if interview and interview.interview_feedback:
#         # 如果已经有评价，则直接返回
#         return render_template('result.html', records=records, interview_feedback=interview.interview_feedback)
#     return render_template('result.html', records=records)
#
# def generate_improvement_suggestions(records):
#     # 整理历史记录文本
#     history_text = "\n".join([f"Q: {rec.question}\nA: {rec.answer}\n" for rec in records])
#
#     # 大模型 API 生成建议
#     prompt = f"基于面试的历史记录，请给出有建设性的改进建议，分段说明，并将重要部分加粗:\n{history_text}"
#
#     headers = {
#         'Authorization': f'Bearer {chat_key}',
#         'Content-Type': 'application/json'
#     }
#     data = {
#         "model": chat_model,  # 指定模型版本
#         "messages": [
#             {"role": "user", "content": prompt}
#         ]
#     }
#
#     response = requests.post(chat_url, headers=headers, json=data)
#     result = response.json()
#
#     suggestions = result.get('choices', [])[0].get('message', {}).get('content', "").strip()
#     return suggestions

# # 查看旧的历史记录
# @app.route('/history', methods=['GET'])
# def history():
#     interviews = InterviewRecord.query.with_entities(InterviewRecord.interview_id).distinct().all()
#     return render_template('history.html', interviews=interviews)

# @app.route('/history/<interview_id>', methods=['GET'])
# def view_history(interview_id):
#     records = InterviewRecord.query.filter_by(interview_id=interview_id).all()
#     suggestions = Interview.query.filter_by(interview_id=interview_id).all()
#     print('suggestions ', suggestions)
#     return render_template('view_history.html', records=records, suggestions=suggestions)

# @app.route('/question')
# def question():
#     return render_template('question.html')

# # 返回到上一个页面的路由
# @app.route('/go_back_result')
# def go_back_result():
#     return redirect(url_for('result'))
# @app.route('/go_back_history')
# def go_back_history():
#     return redirect(url_for('history'))
# @app.route('/go_back_question')
# def go_back_question():
#     return redirect(url_for('question'))
# @app.route('/go_back_interview')
# def go_back_interview():
#     return redirect(url_for('interview'))
# @app.route('/go_back_init')
# def go_back_init():
#     return redirect(url_for('init'))

# # 模拟的数据
# data = {
#     "job_title": "前端工程师",
#     "job_description": "1. 负责公司各类 Web 应用的前端开发工作，包括网页界面的设计、交互效果的实现以及页面性能的优化。\n2. 与产品团队、设计团队和后端开发团队紧密合作，确保前端开发工作与项目整体进度和需求保持一致。\n3. 运用 HTML5、CSS3 和 JavaScript 等前端技术，构建响应式、跨浏览器兼容的页面，提升用户体验。\n4. 参与前端技术选型和框架搭建，持续关注前端技术发展趋势，引入新的技术和工具提升开发效率。\n5. 对前端代码进行优化和维护，解决各种前端技术难题和浏览器兼容性问题，确保网站的稳定性和可靠性。",
#     "resume_text": "具备扎实的前端开发技术，熟练掌握 HTML、CSS、JavaScript 等基础技术。精通 Vue.js 框架，曾参与多个基于 Vue.js 开发的大型项目，熟悉其组件化开发模式和生命周期管理。熟悉前端工程化，熟练使用 Webpack 等构建工具。有良好的代码规范意识，注重代码的可读性和可维护性。能够独立解决各种前端技术问题，如浏览器兼容性、性能优化等。在工作中注重团队协作，能够与不同岗位的人员高效沟通合作。具有较强的学习能力，持续关注前端技术的发展动态，不断提升自己的技术水平。"
# }
#
# @app.route('/get_data')
# def get_data():
#     return jsonify(data)

if __name__ == '__main__':
    # with app.app_context():
    #     db.create_all()
    app.run(debug=True)