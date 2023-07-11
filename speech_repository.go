package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/ServiceWeaver/weaver"
)

type SpeechRepository interface {
	SpeechToText(ctx context.Context, audio []byte) (string, error)
}

// Implementation of the PDFGenerator component.
type speechRepository struct {
	weaver.Implements[SpeechRepository]
}

func (s *speechRepository) SpeechToText(ctx context.Context, audio []byte) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	// Create a multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "openai.mp3")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}
	part.Write(audio)

	// Add other form fields
	_ = writer.WriteField("model", "whisper-1")

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %v", err)
	}

	s.Logger().Info("Sending request to Whisper API")

	// Create an HTTP request with the Whisper API endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request and get the response
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	s.Logger().Info("Received response from Whisper API: " + string(respBody))

	// Check for errors in the response
	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Error string `json:"error"`
		}
		err := json.Unmarshal(respBody, &errorResponse)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal error response: %v", err)
		}
		return "", fmt.Errorf("API request failed: %s", errorResponse.Error)
	}

	// Parse the response to get the transcribed text
	var response struct {
		Text string `json:"text"`
	}
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	s.Logger().Info("Received text from Whisper API: " + response.Text)

	return response.Text, nil
}
