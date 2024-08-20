#安装环境
conda install pytorch==2.3.1 torchvision==0.18.1 torchaudio==2.3.1 pytorch-cuda=12.1 -c pytorch -c nvidia

pip install fastapi peft uvicorn

#下载模型
mkdir -p ../GPT_tmp/models

python ./download_models.py

#启动服务
uvicorn llm_fastapi:app --host 0.0.0.0 --port 8000

#测试普通回应
curl -X POST http://localhost:8000/v1/completions -H "Content-Type: application/json" -d "{\"prompt\": \"Hello, how are you?\", \"max_tokens\": 50}"

#测试流式传输
python C:/Users/meng/workspace/OfferCat/llm_fastapi/stream_test.py