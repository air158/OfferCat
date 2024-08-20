// 获取元素
const infoDiv = document.getElementById('info_div');
if (infoDiv.textContent === '录音关闭') {
  console.log('处于录音关闭状态');
  init_record();
}