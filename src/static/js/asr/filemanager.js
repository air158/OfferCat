
var upfile = document.getElementById('upfile');
var file_ext="";
var file_sample_rate=16000; //for wav file sample rate
var file_data_array;  // array to save file data
var isfilemode=false;  // if it is in file mode


upfile.onclick=function()
{
        btnStart.disabled = true;
        btnStop.disabled = true;
        btnConnect.disabled=false;
        isfilemode=true
}

// from https://github.com/xiangyuecn/Recorder/tree/master
var readWavInfo=function(bytes){
    //读取wav文件头，统一成44字节的头
    if(bytes.byteLength<44){
        return null;
    };
    var wavView=bytes;
    var eq=function(p,s){
        for(var i=0;i<s.length;i++){
            if(wavView[p+i]!=s.charCodeAt(i)){
                return false;
            };
        };
        return true;
    };
    
    if(eq(0,"RIFF")&&eq(8,"WAVEfmt ")){

        var numCh=wavView[22];
        if(wavView[20]==1 && (numCh==1||numCh==2)){//raw pcm 单或双声道
            var sampleRate=wavView[24]+(wavView[25]<<8)+(wavView[26]<<16)+(wavView[27]<<24);
            var bitRate=wavView[34]+(wavView[35]<<8);
            var heads=[wavView.subarray(0,12)],headSize=12;//head只保留必要的块
            //搜索data块的位置
            var dataPos=0; // 44 或有更多块
            for(var i=12,iL=wavView.length-8;i<iL;){
                if(wavView[i]==100&&wavView[i+1]==97&&wavView[i+2]==116&&wavView[i+3]==97){//eq(i,"data")
                    heads.push(wavView.subarray(i,i+8));
                    headSize+=8;
                    dataPos=i+8;break;
                }
                var i0=i;
                i+=4;
                i+=4+wavView[i]+(wavView[i+1]<<8)+(wavView[i+2]<<16)+(wavView[i+3]<<24);
                if(i0==12){//fmt 
                    heads.push(wavView.subarray(i0,i));
                    headSize+=i-i0;
                }
            }
            if(dataPos){
                var wavHead=new Uint8Array(headSize);
                for(var i=0,n=0;i<heads.length;i++){
                    wavHead.set(heads[i],n);n+=heads[i].length;
                }
                return {
                    sampleRate:sampleRate
                    ,bitRate:bitRate
                    ,numChannels:numCh
                    ,wavHead44:wavHead
                    ,dataPos:dataPos
                };
            };
        };
    };
    return null;
};

upfile.onchange = function () {
　　　　　　var len = this.files.length;  
            for(let i = 0; i < len; i++) {

                let fileAudio = new FileReader();
                fileAudio.readAsArrayBuffer(this.files[i]);  

                file_ext=this.files[i].name.split('.').pop().toLowerCase();
                var audioblob;
                fileAudio.onload = function() {
                audioblob = fileAudio.result;

                
                file_data_array=audioblob;

                
                info_div.innerHTML='请点击连接进行识别';

                }

　　　　　　　　　　fileAudio.onerror = function(e) {
　　　　　　　　　　　　console.log('error' + e);
　　　　　　　　　　}
            }
            // for wav file, we  get the sample rate
            if(file_ext=="wav")
            for(let i = 0; i < len; i++) {

                let fileAudio = new FileReader();
                fileAudio.readAsArrayBuffer(this.files[i]);  
                fileAudio.onload = function() {
                audioblob = new Uint8Array(fileAudio.result);

                // for wav file, we can get the sample rate
                var info=readWavInfo(audioblob);
                console.log(info);
                file_sample_rate=info.sampleRate;
    

                }

　　　　　　 
            }

        }

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

function start_file_send()
{
        sampleBuf=new Uint8Array( file_data_array );
        var chunk_size=960; // for asr chunk_size [5, 10, 5]
        while(sampleBuf.length>=chunk_size){
            sendBuf=sampleBuf.slice(0,chunk_size);
            totalsend=totalsend+sampleBuf.length;
            sampleBuf=sampleBuf.slice(chunk_size,sampleBuf.length);
            wsconnecter.wsSend(sendBuf);
        }
        stop();
}
var audio_record = document.getElementById('audio_record');
var mic_mode_div = document.getElementById("mic_mode_div");
var rec_mode_div = document.getElementById("rec_mode_div");
function on_recoder_mode_change(){
    var item = null;
    var obj = document.getElementsByName("recoder_mode");
    for (var i = 0; i < obj.length; i++) { //遍历Radio 
        if (obj[i].checked) {
            item = obj[i].value;  
            break;
        }
    }
    if(item=="mic")
    {
        mic_mode_div.style.display = 'block';
        rec_mode_div.style.display = 'none';


        btnStart.disabled = true;
        btnStop.disabled = true;
        btnConnect.disabled=false;
        isfilemode=false;
    }
    else
    {
        mic_mode_div.style.display = 'none';
        rec_mode_div.style.display = 'block';

        btnStart.disabled = true;
        btnStop.disabled = true;
        btnConnect.disabled=true;
        isfilemode=true;
        info_div.innerHTML='请点击选择文件';
    }
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