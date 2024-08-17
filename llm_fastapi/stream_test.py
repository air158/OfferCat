import requests
import json

url = "http://localhost:8000/v1/streaming-completions"
headers = {
    "Content-Type": "application/json"
}
data = {
    "prompt": "Hello, how are you?",
    "max_tokens": 50
}

print('start')

response = requests.post(url, headers=headers, data=json.dumps(data), stream=True)

print('response ', response)

for line in response.iter_lines():
    if line:
        decoded_line = line.decode('utf-8')
        print(decoded_line)