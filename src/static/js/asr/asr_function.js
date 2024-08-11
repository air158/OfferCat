// var btnStart = document.getElementById('btnStart');
var btnStop = document.getElementById('btnStop');
var btnConnect = document.getElementById('btnConnect');
var info_div = document.getElementById('info_div');
var upfile = document.getElementById('upfile');

var rec_text="";  // for online rec asr result
var offline_text=""; // for offline rec asr result

var totalsend=0;

// 连接; 定义socket连接类对象与语音对象
var wsconnecter = new WebSocketConnectMethod({msgHandle:getJsonMessage,stateHandle:getConnState});

function init_record() {
	if(btnStop.disabled == false ){
		console.log("record already started")
		return;
	}
	start();
	record();

	// btnStart.disabled = true;
	btnStop.disabled = false;
	btnConnect.disabled=true;
}

// 定义按钮响应事件

// btnStart.onclick = record;
function record()
{
		 rec.open( function(){
		 rec.start();
		 console.log("开始");
			// btnStart.disabled = true;
			btnStop.disabled = false;
			btnConnect.disabled=true;
		 });
 
}
// btnStart.disabled = true;

btnConnect.onclick = start;
// 识别启动、停止、清空操作
function start() {
	if(btnStop.disabled == false) {
		console.log("already recording");
		return;
	}
	// 清除显示
	clear();
	//控件状态更新
 	console.log("isfilemode "+isfilemode);
	//启动连接
	var ret=wsconnecter.wsStart();
	// 1 is ok, 0 is error
	if(ret==1){
		info_div.innerHTML="正在连接asr服务器，请等待...";
		isRec = true;
		// btnStart.disabled = true;
		btnStop.disabled = true;
		btnConnect.disabled=true;
 
        return 1;
	}
	else
	{
		info_div.innerHTML="请点击开始";
		// btnStart.disabled = true;
		btnStop.disabled = true;
		btnConnect.disabled=false;
 
		return 0;
	}
}

btnStop.onclick = stop;
function stop() {
		var chunk_size = new Array( 5, 10, 5 );
		var request = {
			"chunk_size": chunk_size,
			"wav_name":  "h5",
			"is_speaking":  false,
			"chunk_interval":10,
			"mode":getAsrMode(),
		};
		console.log(request);
		if(sampleBuf.length>0){
		wsconnecter.wsSend(sampleBuf);
		console.log("sampleBuf.length"+sampleBuf.length);
		sampleBuf=new Int16Array();
		}
	   wsconnecter.wsSend( JSON.stringify(request) );
	// 控件状态更新
	isRec = false;
    info_div.innerHTML="发送完数据,请等候,正在识别...";
   if(isfilemode==false){
	    btnStop.disabled = true;
		// btnStart.disabled = true;
		btnConnect.disabled=true;
		//wait 3s for asr result
	  setTimeout(function(){
		console.log("call stop ws!");
		wsconnecter.wsStop();
		btnConnect.disabled=false;
		info_div.innerHTML="请点击连接";}, 3000 );
	rec.stop(function(blob,duration){
		console.log(blob);
		var audioBlob = Recorder.pcm2wav(data = {sampleRate:16000, bitRate:16, blob:blob},
		function(theblob,duration){
				console.log(theblob);
		var audio_record = document.getElementById('audio_record');
		audio_record.src =  (window.URL||webkitURL).createObjectURL(theblob); 
        audio_record.controls=true;
		//audio_record.play(); 
	}   ,function(msg){
		 console.log(msg);
	}
		);
	},function(errMsg){
		console.log("errMsg: " + errMsg);
	});
   }
    // 停止连接
}
btnStop.disabled = true;

function clear() {
	clearAllBubbles();
    rec_text="";
	offline_text="";
}

// 模拟语音识别API，增加更多示例数据
function simulateAPI() {
    const sentences = [
        "欢迎使用OfferCat!",
        "你了解transformer的结构吗",
        "用类写出注意力机制的代码",
        "如何提高多模态大模型的OCR能力",
        "写最长回文子串",
        "这是一个垂直布局的示例",
        "每个泡泡代表一句语音识别的结果",
        "您可以上下滑动来查看更多的语音识别结果",
        "新的泡泡会自动滚动到底部",
        "如果您向上滚动，新泡泡就不会自动滚动了",
        "当您自己滚动到底部时，自动滚动功能会恢复",
        "点击可以选中或取消选中某个泡泡",
        "拖动可以批量选择或取消多个泡泡",
        "选中后点击右下角的绿色按钮提交内容",
        "AI的回答将在下方的弹出框中显示",
        "这个泡泡用于测试滚动效果",
        "继续向下滑动可以看到更多的内容",
        "我们正在测试智能滚动",
        "滑动是否流畅？自动滚动功能是否正常？",
        "您可以尝试选择多个泡泡然后提交",
        "AI会根据您选择的内容给出回应",
        "这是最后一个泡泡，感谢您的耐心测试"
    ];

    sentences.forEach((sentence, index) => {
        setTimeout(() => {
            currentBubble = createBubble();
            typeSentence(currentBubble, sentence, 0);
        }, index * (sentence.length * 50 + 1000));
    });
}

function typeSentence(bubble, sentence, charIndex) {
    if (charIndex < sentence.length) {
        updateBubble(bubble, sentence.slice(0, charIndex + 1));
        setTimeout(() => {
            typeSentence(bubble, sentence, charIndex + 1);
        }, 50);  // 逐字显示的时间间隔
    } else {
        setTimeout(() => {
            updateBubble(bubble, sentence + " (已矫正)");
        }, 500);  // 完整句子显示后 500 毫秒添加 "已矫正"
    }
}