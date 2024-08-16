#下载模型
mkdir -p ../GPT_tmp/models

python ./download_models.py

uvicorn llm_fastapi:app --host 0.0.0.0 --port 8000
uvicorn llm_fastapi_cpu:app --host 0.0.0.0 --port 8000

curl -X POST http://localhost:8000/v1/completions \
-H "Content-Type: application/json" \
-d '{"prompt": "Hello, how are you?", "max_tokens": 50}'

http --stream POST http://localhost:8000/v1/streaming-completions \
Content-Type:application/json \
prompt="Hello, how are you?" \
max_tokens:=50