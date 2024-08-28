from flask_sqlalchemy import SQLAlchemy
from datetime import datetime

db = SQLAlchemy()

class JobInfo(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    job_title = db.Column(db.String(100), nullable=False)
    job_description = db.Column(db.Text, nullable=False)
    resume_text = db.Column(db.Text, nullable=False)

class InterviewRecord(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    job_title = db.Column(db.String(100), nullable=False)
    question = db.Column(db.Text, nullable=False)
    answer = db.Column(db.Text, nullable=False)
    duration = db.Column(db.Float, nullable=False)
    timestamp = db.Column(db.Float, nullable=False)
    interview_id = db.Column(db.String(100), nullable=False)  # 关联每次面试的唯一ID

# 定义存储问题的模型
class QuestionData(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    interview_id = db.Column(db.String(100), nullable=False)
    questions = db.Column(db.Text)

class Interview(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    interview_id = db.Column(db.String(100), unique=True, nullable=False)
    job_title = db.Column(db.String(128), nullable=False)
    interview_feedback = db.Column(db.Text, nullable=True)  # 存储面试的评价

class RunnigInterview(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, nullable=False)
    interview_id = db.Column(db.Integer, nullable=False)

class User(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    user_name = db.Column(db.String(128), nullable=False)
    user_id = db.Column(db.Integer, nullable=False)
    interview_points = db.Column(db.Integer, nullable=False)
