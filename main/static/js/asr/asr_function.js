var info_div = document.getElementById('info_div');

var rec_text="";  // for online rec asr result
var offline_text=""; // for offline rec asr result

var totalsend=0;

// 连接; 定义socket连接类对象与语音对象
var wsconnecter = new WebSocketConnectMethod({msgHandle:getJsonMessage,stateHandle:getConnState});

function init_record() {
	if(record_flag == false) {
		console.log('start record!')
		start();
		record();
	}
}

// 定义按钮响应事件
record_flag = false;

function record()
{

		 rec.open( function(){
		 rec.start();
		 console.log("开始");
		 });
		 record_flag = true;
 
}

// 识别启动、停止、清空操作
function start() {
	console.log("start!!!")
	// 清除显示
	clear();
	//启动连接
	var ret=wsconnecter.wsStart();
	// 1 is ok, 0 is error
	if(ret==1){
		info_div.innerHTML="正在连接asr服务器，请等待...";
		isRec = true;
 
        return 1;
	}
	else
	{
		info_div.innerHTML="请点击开始";
 
		return 0;
	}
}


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
		//wait 3s for asr result
	  setTimeout(function(){
		console.log("call stop ws!");
		wsconnecter.wsStop();
		info_div.innerHTML="请点击连接";}, 3000 );
	rec.stop(function(blob,duration){
		console.log(blob);
		var audioBlob = Recorder.pcm2wav(data = {sampleRate:16000, bitRate:16, blob:blob},
		function(theblob,duration){
				console.log(theblob);
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

function clear() {
    rec_text="";
	offline_text="";
}
