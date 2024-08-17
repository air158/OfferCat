
var isfilemode=false;  // if it is in file mode

function getHotwords(){

    var obj = document.getElementById("varHot");

    if(typeof(obj) == 'undefined' || obj==null || obj.value.length<=0){
        return null;
    }
    let val = obj.value.toString();
    
    console.log("hotwords="+val);
    let items = val.split(/[(\r\n)\r\n]+/);  //split by \r\n
    var jsonresult = {};
    const regexNum = /^[0-9]*$/; // test number
    for (item of items) {
    
        let result = item.split(" ");
        if(result.length>=2 && regexNum.test(result[result.length-1]))
        { 
            var wordstr="";
            for(var i=0;i<result.length-1;i++)
                wordstr=wordstr+result[i]+" ";
    
            jsonresult[wordstr.trim()]= parseInt(result[result.length-1]);
        }
    }
    console.log("jsonresult="+JSON.stringify(jsonresult));
    return  JSON.stringify(jsonresult);

}


var sampleBuf=new Int16Array();
var audioBlob;
// 录音; 定义录音对象,wav格式
var rec = Recorder({
	type:"pcm",
	bitRate:16,
	sampleRate:16000,
	onProcess:recProcess
});
function recProcess( buffer, powerLevel, bufferDuration, bufferSampleRate,newBufferIdx,asyncEnd ) {
	if ( isRec === true ) {
		var data_48k = buffer[buffer.length-1];  
		var  array_48k = new Array(data_48k);
		var data_16k=Recorder.SampleData(array_48k,bufferSampleRate,16000).data;
		sampleBuf = Int16Array.from([...sampleBuf, ...data_16k]);
		var chunk_size=960; // for asr chunk_size [5, 10, 5]
		info_div.innerHTML=""+bufferDuration/1000+"s";
		while(sampleBuf.length>=chunk_size){
		    sendBuf=sampleBuf.slice(0,chunk_size);
			sampleBuf=sampleBuf.slice(chunk_size,sampleBuf.length);
			wsconnecter.wsSend(sendBuf);
		}
	}
}

function getUseITN(){
    return true
}
function getAsrMode(){
   return "2pass";
}