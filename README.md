# OfferCat
背八股和刷算法题是一个费时费力而且很痛苦的事，这是每个找实习或找工作的计算机专业学生必须要经历的事。
<br>
但是大模型压缩了这些知识，如果能在面试的时候使用大模型这个知识库，则能省去背八股和刷算法题这个痛苦的流程。
<br>
面试中不方便打字和大模型交互，而现有的语音识别软件也没有将识别结果询问大模型的功能。
<br>
所以我们的项目能实时语音识别面试官的问题，并直接传给大模型来回答，帮助用户回答八股题或写算法题，这将彻底改变传统的面试流程。

# Roadmap
- [x] 实时语音识别
- [x] 大模型回复
- [x] 服务器，反向代理，内网穿透
## 
- [ ] 登录注册界面
- [ ] 域名
- [ ] 用户购买兑换码来使用服务（初期方案，方便实现，但是只能小范围使用）
- [ ] 在校内论坛进行宣传和测试
## 
- [ ] 接入第三方聚合支付（方便但是可能不可靠，不是长久之计）
- [ ] 接入支付宝官方接口（可能需要注册公司并办理营业执照）
- [ ] 接入微信支付官方接口（可能需要注册公司并办理营业执照）
## 
- [ ] 大模型识别面试官问题
- [ ] 自定义提问prompt
- [ ] 上传简历并RAG解析
## 
- [ ] 优化前端页面
- [ ] 优化后端代码，转为Go或Java
- [ ] 对代码和通信进行加密保证安全性
##
- [ ] 拍摄视频，在抖音，微信视频进行宣传引流
- [ ] 与抖音博主和B站UP合作，在视频中插入广告
##
- [ ] 构建面试数据集
- [ ] RAG面试题知识库
- [ ] 用数据集进行LoRA微调
##
- [ ] 测试音频大模型Qwen-Audio提升语音识别质量和回答速度
- [ ] 直接获取电脑音频（比如腾讯会议）以提升语音识别准确率
##

# 安装
## clone仓库
```
git clone https://github.com/air158/OfferCat.git
```
## 环境依赖（如果没有conda也可以直接pip3安装环境，然后python3运行服务器）
```
conda create -n offercat python=3.11
conda activate offercat
pip install requests Flask
```
## docker安装
如果您已安装docker，忽略本步骤！!
通过下述命令在服务器上安装docker：
```shell
curl -O https://isv-data.oss-cn-hangzhou.aliyuncs.com/ics/MaaS/ASR/shell/install_docker.sh

sudo bash install_docker.sh
```
docker安装失败请参考 [Docker Installation](https://alibaba-damo-academy.github.io/FunASR/en/installation/docker.html)

## 环境变量
使用了讯飞星火大模型api，需要在运行服务器的终端输入api key：
```
export SPARK_API_KEY="填入对应的 api key"
export SPARK_API_SECRET="填入对应的 api secret"
```
使用了https，需要填入ssl证书密码：
```
export SSL_PW="填入对应的ssl证书密码"
```

# 语音ASR服务
## 镜像启动
通过下述命令拉取并启动FunASR软件包的docker镜像，服务会启动到10096端口,：
```shell
sudo docker pull registry.cn-hangzhou.aliyuncs.com/funasr_repo/funasr:funasr-runtime-sdk-online-cpu-0.1.10

mkdir -p ./funasr-runtime-resources/models

sudo docker run -p 10096:10095 -it --privileged=true -v $PWD/funasr-runtime-resources/models:/workspace/models registry.cn-hangzhou.aliyuncs.com/funasr_repo/funasr:funasr-runtime-sdk-online-cpu-0.1.10
```
## 服务端启动
docker启动之后会进入docker界面，启动 funasr-wss-server-2pass服务程序：
```shell
cd FunASR/runtime
nohup bash run_server_2pass.sh \
  --download-model-dir /workspace/models \
  --vad-dir damo/speech_fsmn_vad_zh-cn-16k-common-onnx \
  --model-dir damo/speech_paraformer-large-vad-punc_asr_nat-zh-cn-16k-common-vocab8404-onnx  \
  --online-model-dir damo/speech_paraformer-large_asr_nat-zh-cn-16k-common-vocab8404-online-onnx  \
  --punc-dir damo/punc_ct-transformer_zh-cn-common-vad_realtime-vocab272727-onnx \
  --lm-dir damo/speech_ngram_lm_zh-cn-ai-wesp-fst \
  --itn-dir thuduj12/fst_itn_zh \
  --hotword /workspace/models/hotwords.txt > log.txt 2>&1 &

# 如果您想关闭ssl（使用http)，增加参数：--certfile 0
# 如果您想使用时间戳或者nn热词模型进行部署，请设置--model-dir为对应模型：
#   damo/speech_paraformer-large-vad-punc_asr_nat-zh-cn-16k-common-vocab8404-onnx（时间戳）
#   damo/speech_paraformer-large-contextual_asr_nat-zh-cn-16k-common-vocab8404-onnx（nn热词）
# 如果您想在服务端加载热词，请在宿主机文件./funasr-runtime-resources/models/hotwords.txt配置热词（docker映射地址为/workspace/models/hotwords.txt）:
#   每行一个热词，格式(热词 权重)：你好 20（注：热词理论上无限制，但为了兼顾性能和效果，建议热词长度不超过10，个数不超过1k，权重1~100）
```
# 使用
## 启动服务器
启动服务器后，本机通过https://127.0.0.1,其他机器可以通过https://{本机ip}访问
```
cd OfferCat/src
python ./test_spark.py
```
## 使用流程
1. 首先需要授权asr服务器，可以点击链接也可以直接访问https://{本机ip}:10096
2. 查看页面提示后，点击箭头或暗处可关闭提示
3. 点击左下角的按钮开始录音：
4. 录音 -> 开始录音
5. 语音识别后点击泡泡（可多选）后再点击右下角的提问，固定的泡泡是语音识别的最新一句话，也可以查看语音识别的历史记录泡泡
   
# 文件目录
主要代码在src中，ssl中存放的是ssl证书，static是html的静态资源，template中事html主页面，test_spark.py则是服务器代码。
# 结构
```
├── README.md
├── requirements.txt
└── src
    ├── ssl
    │   ├── cert.pem
    │   └── key.pem
    ├── static
    │   ├── css
    │   │   ├── highlight.css
    │   │   ├── setting.css
    │   │   └── styles.css
    │   └── js
    │       ├── asr
    │       │   ├── asr.js
    │       │   ├── asr_function.js
    │       │   ├── filemanager.js
    │       │   ├── pcm.js
    │       │   ├── recorder-core.js
    │       │   ├── wav.js
    │       │   ├── wsconnecter.js
    │       │   └── wss_config.js
    │       ├── llm.js
    │       └── ui
    │           ├── bubbles.js
    │           ├── highlight.min.js
    │           ├── iphone_padding.js
    │           ├── marked.min.js
    │           └── setting.js
    ├── templates
    │   └── index.html
    └── test_spark.py
```