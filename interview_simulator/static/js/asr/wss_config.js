// nowip = "10.29.173.59"
nowip = "127.0.0.1"


//这里的port是在阿里云服务器上用了内网穿透和反向代理，如果是本机运行的话用10096端口
nowport = "8443/ws/"
// nowport = "10096"

//wss是https，ws是http
wss_type = "wss://"
// wss_type = "ws://"


var wws_url=wss_type + nowip + ":" + nowport+"/"
var ip_url = wws_url

function processUri(now_ipaddress) {
    now_ipaddress=now_ipaddress.replace("localhost", "127.0.0.1");
    now_ipaddress=now_ipaddress.replace("/#","");
    now_ipaddress=now_ipaddress.replace("5/","5");
    now_ipaddress=now_ipaddress.replace("http://", wss_type);
    now_ipaddress=now_ipaddress.replace("https://", wss_type);
    now_ipaddress=now_ipaddress.replace("static/index.html","");
    now_ipaddress=now_ipaddress.replace("templates/index.html","");
    return now_ipaddress
}
function initWss() {
    var now_ipaddress=window.location.href;
    now_ipaddress = processUri(now_ipaddress)

    // var localport=window.location.port;
    // now_ipaddress=now_ipaddress.replace(localport,nowport);
    
    // wws_url=now_ipaddress+":"+nowport;
    wws_url="wss://101.201.82.35:8443/ws/"
    console.log("wsip " + wws_url)
    // window.open(wws_url.replace(wss_type, "https://"), '_blank');
    ip_url=wws_url.replace(wss_type, "https://")
}
initWss();
