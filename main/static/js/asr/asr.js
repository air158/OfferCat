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


var data_time = null;
// 语音识别结果; 对jsonMsg数据解析,将识别结果附加到编辑框中
function getJsonMessage( jsonMsg ) {

    var rectxt=""+JSON.parse(jsonMsg.data)['text'];
    var asrmodel=JSON.parse(jsonMsg.data)['mode'];
    var is_final=JSON.parse(jsonMsg.data)['is_final'];
    var timestamp=JSON.parse(jsonMsg.data)['timestamp'];


    if(asrmodel=="2pass-offline" || asrmodel=="offline")
    {
        [text, data_time] = handleWithTimestamp(rectxt,timestamp);
        offline_text=offline_text+text; //rectxt; //.replace(/ +/g,"");
        rec_text=offline_text;
    }
    else
    {
        rec_text=rec_text+rectxt; //.replace(/ +/g,"");
    }

    document.getElementById('answer').value = rec_text;


}
// 连接状态响应
function getConnState( connState ) {
	if ( connState === 0 ) { //on open
		info_div.innerHTML='连接成功!请点击开始';
	} else if ( connState === 1 ) {
		//stop();
	} else if ( connState === 2 ) {
		stop();
		console.log( 'connecttion error' );

        alert("连接地址"+wws_url+"失败,请检查asr地址和端口。或试试界面上手动授权，再连接。");
 
		info_div.innerHTML='请点击连接';
	}
}