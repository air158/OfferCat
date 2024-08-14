package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"strings"
)

type ChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Stream bool `json:"stream"`
}

func CallSparkAPI(userInput string) (string, error) {
	apiKey := viper.GetString("apiKey")
	apiSecret := viper.GetString("apiSecret")

	if apiKey == "" || apiSecret == "" {
		return "", fmt.Errorf("API key or secret is not set")
	}

	url := "https://spark-api-open.xf-yun.com/v1/chat/completions"
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s:%s", apiKey, apiSecret),
		"Content-Type":  "application/json",
	}

	data := ChatRequest{
		Model: "generalv3.5",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: userInput,
			},
		},
		Stream: true,
	}

	requestBody, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	log.Printf("Request body: %s", requestBody)

	// Channel for receiving the result
	resultChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		client := &http.Client{}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			errorChan <- fmt.Errorf("error creating request: %w", err)
			return
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		log.Println("Sending request to API...")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error making request: %v", err)
			errorChan <- fmt.Errorf("error making request: %w", err)
			return
		}
		defer resp.Body.Close()

		log.Printf("Received response with status code: %d", resp.StatusCode)

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("Unexpected status code response body: %s", string(bodyBytes))
			errorChan <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			return
		}

		var fullResponse strings.Builder
		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				log.Println("End of response stream.")
				break
			} else if err != nil {
				log.Printf("Error reading response: %v", err)
				errorChan <- fmt.Errorf("error reading response: %w", err)
				return
			}

			log.Printf("Raw response line: %s", line)

			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "data: ") {
				line = strings.TrimPrefix(line, "data: ")
				if line == "[DONE]" {
					log.Println("Received DONE signal.")
					break
				}

				var decodedLine map[string]interface{}
				if err := json.Unmarshal([]byte(line), &decodedLine); err != nil {
					log.Printf("Error decoding response line: %v", err)
					errorChan <- fmt.Errorf("error decoding response line: %w", err)
					return
				}

				log.Printf("Decoded line: %v", decodedLine)

				if choices, ok := decodedLine["choices"].([]interface{}); ok {
					if choice, ok := choices[0].(map[string]interface{}); ok {
						if delta, ok := choice["delta"].(map[string]interface{}); ok {
							if content, ok := delta["content"].(string); ok {
								fullResponse.WriteString(content)
								log.Printf("Appended content: %s", content)
							}
						}
					}
				}
			}
		}

		resultChan <- fullResponse.String()
		log.Println("Finished reading response.")
	}()

	select {
	case result := <-resultChan:
		log.Println("Returning result from CallSparkAPI.")
		return result, nil
	case err := <-errorChan:
		log.Printf("Returning error from CallSparkAPI: %v", err)
		return "", err
	}
}
