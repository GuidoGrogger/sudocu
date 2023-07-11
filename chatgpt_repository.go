package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ServiceWeaver/weaver"
)

type ChatGPTRepository interface {
	ChangeMarkup(ctx context.Context, oldMarkup string, prompt string) ([]byte, error)
}

// Implementation of the PDFGenerator component.
type chatGPTRepository struct {
	weaver.Implements[ChatGPTRepository]
}

type ChatGPTRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Temperature      float64   `json:"temperature"`
	MaxTokens        int       `json:"max_tokens"`
	TopP             float64   `json:"top_p"`
	FrequencyPenalty float64   `json:"frequency_penalty"`
	PresencePenalty  float64   `json:"presence_penalty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTResponse struct {
	Choices []Choice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Choice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

func (c *chatGPTRepository) ChangeMarkup(ctx context.Context, oldMarkup string, prompt string) ([]byte, error) {
	request := ChatGPTRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: "Change this ascii-doc based on the user input:\n" + oldMarkup,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature:      1,
		MaxTokens:        2048,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	c.Logger().Info("Seding request: ", string(requestBody))

	url := "https://api.openai.com/v1/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c.Logger().Info("Got response: ", string(responseBody))

	var chatGPTResponse ChatGPTResponse
	err = json.Unmarshal(responseBody, &chatGPTResponse)
	if err != nil {
		return nil, err
	}

	c.Logger().Info("Total Token Usage", chatGPTResponse.Usage.TotalTokens, ", estimated amount", float64(chatGPTResponse.Usage.TotalTokens)/1000*0.2)

	if len(chatGPTResponse.Choices) > 0 {
		return []byte(chatGPTResponse.Choices[0].Message.Content), nil
	}

	return nil, fmt.Errorf("empty response from ChatGPT")
}
