const iphone_padding = document.getElementById('iphone-padding');
// 检测是否是iPhone
function isiPhone() {
    return /iPhone|iPod/.test(navigator.userAgent) && !window.MSStream;
}

// 如果是iPhone，则在<body>元素上添加一个类
if (isiPhone()) {
    iphone_padding.style.display = 'block';
    // iphonePadding.style.paddingBottom = "90px";
}