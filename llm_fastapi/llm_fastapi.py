from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import StreamingResponse, JSONResponse
from transformers import AutoTokenizer, AutoModelForCausalLM
import torch
from peft import PeftModel
import json
import time
import asyncio

# 定义模型路径
model_dir = '/Users/didi/workspace/OfferCat/GPT_tmp/models/IEITYuan/Yuan2-2B-Mars-hf'
lora_path = '/Users/didi/workspace/OfferCat/GPT_tmp/OfferCat_Yuan2.0-2B_lora_bf16/checkpoint-850'

# 定义模型数据类型
torch_dtype = torch.bfloat16  # A10
# torch_dtype = torch.float16  # P100

app = FastAPI()

# 定义一个函数，用于获取模型和tokenizer
def get_model():
    print("Creating tokenizer...")
    tokenizer = AutoTokenizer.from_pretrained(model_dir, add_eos_token=False, add_bos_token=False, eos_token='<eod>')
    tokenizer.add_tokens(['<sep>', '<pad>', '<mask>', '<predict>', '<FIM_SUFFIX>', '<FIM_PREFIX>', '<FIM_MIDDLE>', '<commit_before>', '<commit_msg>', '<commit_after>', '<jupyter_start>', '<jupyter_text>', '<jupyter_code>', '<jupyter_output>', '<empty_output>'], special_tokens=True)

    print("Creating model...")
    model = AutoModelForCausalLM.from_pretrained(model_dir, torch_dtype=torch_dtype, trust_remote_code=True).cuda()
    model = PeftModel.from_pretrained(model, model_id=lora_path)

    return tokenizer, model

# 加载模型和tokenizer
tokenizer, model = get_model()

# 定义一个生成函数用于普通回复
def generate_text(prompt: str, max_length: int = 200):
    inputs = tokenizer(prompt, return_tensors="pt").to(model.device)
    outputs = model.generate(**inputs, max_new_tokens=max_length)
    return tokenizer.decode(outputs[0], skip_special_tokens=True)

# 定义一个流式生成函数用于流式回复
async def stream_generate_text(prompt: str, max_length: int = 200):
    inputs = tokenizer(prompt, return_tensors="pt").to(model.device)
    outputs = model.generate(**inputs, max_new_tokens=max_length, do_sample=True)
    decoded_output = tokenizer.decode(outputs[0], skip_special_tokens=True)
    
    for token in decoded_output.split():
        # 逐步输出每个 token，模拟流式生成
        yield json.dumps({"choices": [{"delta": {"content": token}, "finish_reason": None}]}) + "\n"
        await asyncio.sleep(0.1)  # 模拟生成时间

    # 结束符标志
    yield json.dumps({"choices": [{"delta": {}, "finish_reason": "stop"}]}) + "\n"

# 普通回复 API 端点
@app.post("/v1/completions")
async def completions(request: Request):
    data = await request.json()
    prompt = data.get("prompt", "")
    max_tokens = data.get("max_tokens", 200)
    
    if not prompt:
        raise HTTPException(status_code=400, detail="Prompt is required")

    # 生成文本
    generated_text = generate_text(prompt, max_length=max_tokens)
    
    return JSONResponse(content={
        "id": "cmpl-1",
        "object": "text_completion",
        "created": int(time.time()),
        "model": "Yuan2-2B-Mars-hf",
        "choices": [{"text": generated_text, "finish_reason": "stop"}]
    })

# 流式回复 API 端点
@app.post("/v1/streaming-completions")
async def streaming_completions(request: Request):
    data = await request.json()
    prompt = data.get("prompt", "")
    max_tokens = data.get("max_tokens", 200)

    if not prompt:
        raise HTTPException(status_code=400, detail="Prompt is required")

    # 流式生成文本
    return StreamingResponse(stream_generate_text(prompt, max_length=max_tokens), media_type="text/event-stream")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)