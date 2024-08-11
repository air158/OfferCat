function handleWithTimestamp(tmptext, tmptime) {
    if (tmptime == null || tmptime == "undefined" || tmptext.length <= 0) {
        return tmptext;
    }
    tmptext = tmptext.replace(/。|？|，|、|\?|\.|\ /g, ","); // in case there are a lot of "。"
    var words = tmptext.split(",");  // split to Chinese sentence or English words
    var jsontime = JSON.parse(tmptime); //JSON.parse(tmptime.replace(/\]\]\[\[/g, "],[")); // in case there are a lot segments by VAD
    var char_index = 0; // index for timestamp
    var text_withtime = "";
    for (var i = 0; i < words.length; i++) {
        if (words[i] == "undefined" || words[i].length <= 0) {
            continue;
        }

        // addInnerDiv(varArea, jsontime[char_index][0] / 1000, words[i])
        text_withtime=text_withtime+words[i]

        if (/^[a-zA-Z]+$/.test(words[i])) {   // if it is English
            char_index = char_index + 1;  // for English, timestamp unit is about a word
        } else {
            char_index = char_index + words[i].length; // for Chinese, timestamp unit is about a char
        }
    }
    return [text_withtime, jsontime[0][0]/1000];
}

var needNewBub=true;
var nowBub;
var data_time = null;
// 语音识别结果; 对jsonMsg数据解析,将识别结果附加到编辑框中
function getJsonMessage( jsonMsg ) {
    // console.log(jsonMsg);
    // console.log( "message: " + JSON.parse(jsonMsg.data)['text'] );
    var rectxt=""+JSON.parse(jsonMsg.data)['text'];
    var asrmodel=JSON.parse(jsonMsg.data)['mode'];
    var is_final=JSON.parse(jsonMsg.data)['is_final'];
    var timestamp=JSON.parse(jsonMsg.data)['timestamp'];

    if(needNewBub) {
        nowBub = createBubble();
        rec_text="";
        offline_text="";
    }

    if(asrmodel=="2pass-offline" || asrmodel=="offline")
    {
        [text, data_time] = handleWithTimestamp(rectxt,timestamp);
        offline_text=offline_text+text; //rectxt; //.replace(/ +/g,"");
        rec_text=offline_text;
        needNewBub=true;
    }
    else
    {
        rec_text=rec_text+rectxt; //.replace(/ +/g,"");
        needNewBub=false;
    }
	updateBubble(nowBub, rec_text);
    setInterviewerQuesiton(rec_text)

    if (isfilemode==true && is_final==true){
        console.log("call stop ws!");
        wsconnecter.wsStop();
        // info_div.innerHTML="请点击连接";
        // btnStart.disabled = true;
        btnStop.disabled = true;
        btnConnect.disabled=false;
    }
}
// 连接状态响应
function getConnState( connState ) {
	if ( connState === 0 ) { //on open
		info_div.innerHTML='连接成功!请点击开始';
		if (isfilemode==true){
			info_div.innerHTML='请耐心等待,大文件等待时间更长';
			start_file_send();
		}
		else
		{
			// btnStart.disabled = false;
			btnStop.disabled = true;
			btnConnect.disabled=true;
		}
        record();
	} else if ( connState === 1 ) {
		//stop();
	} else if ( connState === 2 ) {
		stop();
		console.log( 'connecttion error' );

        alert("连接地址"+wws_url+"失败,请检查asr地址和端口。或试试界面上手动授权，再连接。");
        wws_url
		 
		// alert("连接地址"+document.getElementById('wssip').value+"失败,请检查asr地址和端口。或试试界面上手动授权，再连接。");
		// btnStart.disabled = true;
		btnStop.disabled = true;
		btnConnect.disabled=false;
 
		info_div.innerHTML='请点击连接';
	}
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