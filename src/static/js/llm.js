function AIResponse(input) {
    lowerSection.style.display = 'block';
    responseContent.innerHTML = '';
    startStream(input);
}
var collectedText = ''; // 用于存储所有流式输入内容
var activeReader;
function startStream(userInput) {
    collectedText = ''

    if (activeReader) {
        activeReader.cancel(); // 取消当前正在进行的流
    }

    fetch('/stream', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({
            'user_input': userInput
        })
    }).then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        const reader = response.body.getReader();
        activeReader = reader; // 更新当前的reader对象
        const decoder = new TextDecoder();

        function readStream() {
            reader.read().then(({ done, value }) => {
                if (done) {
                    activeReader = null; // 流结束后清空reader对象
                    reader.cancel(); 
                    return;
                }
                const text = decoder.decode(value);
                collectedText += text; // 收集流式输入内容
                // 移除最后的 "[DONE]"
                if (collectedText.includes("[DONE]")) {
                    collectedText = collectedText.replace("[DONE]", "");
                }
                // 解析Markdown并更新responseContent
                responseContent.innerHTML = marked.parse(collectedText);

                // 使用highlight.js对代码块进行高亮处理
                document.querySelectorAll('pre code').forEach((block) => {
                    hljs.highlightElement(block);
                });

                if (text.includes("[DONE]")) {
                    activeReader = null; // 流结束后清空reader对象
                    reader.cancel(); // Stop reading the stream when [DONE] is received
                } else {
                    readStream();
                }
            }).catch(error => {
                console.error('Error reading the stream:', error);
            });
        }

        readStream();
    }).catch(error => {
        console.error('There was a problem with the fetch operation:', error);
    });
}