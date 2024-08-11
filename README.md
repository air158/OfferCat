# OfferCat
背八股和刷算法题是一个费时费力而且很痛苦的事，这是每个找实习或找工作的计算机专业学生必须要经历的事。
<br>
但是大模型压缩了这些知识，我们可以用大模型作为面试官模拟面试，大模型可以针对面试表现给用户提供改进建议，并且如果能在正式面试的时候使用大模型回答的问题作为题词器，则能省去背八股和刷算法题这个痛苦的流程。
<br>
我们的项目能支持大模型模拟面试，大模型面试官依据岗位和简历信息题面试问题，用户作为面试者可以用文字或语音的方式回答。
<br>
我们还支持正式面试中的大模型提词器，实时语音识别面试官的问题和面试者，并传给大模型来回答，帮助用户回答八股题或写算法题
<br>
最后支持整理面经和提供面试建议
<br>
这将彻底改变传统的面试流程。

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