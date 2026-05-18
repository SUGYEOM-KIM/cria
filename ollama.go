package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatResponse struct {
	Message ChatMessage `json:"message"`
	Error   string      `json:"error,omitempty"`
}

type OllamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func fetchOllamaModels() []string {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{}
	}

	var tagsResp OllamaTagsResponse
	if err := json.Unmarshal(bodyBytes, &tagsResp); err != nil {
		return []string{}
	}

	var models []string
	for _, m := range tagsResp.Models {
		models = append(models, m.Name)
	}

	return models
}

func chatWithOllama(modelName string, prompt string) string {
	reqData := ChatRequest{
		Model: modelName,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "Error: Could not process request data."
	}

	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "Error: Could not connect to Ollama. Make sure it is running."
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Error: Could not read response from Ollama."
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "Error: Could not parse response from Ollama."
	}

	if chatResp.Error != "" {
		return fmt.Sprintf("Ollama Error: %s", chatResp.Error)
	}

	return chatResp.Message.Content
}

func downloadOllamaModel(ctx context.Context, modelName string) string {
	cmd := exec.Command("ollama", "pull", modelName)

	if currentPath := os.Getenv("OLLAMA_MODELS"); currentPath != "" {
		cmd.Env = append(os.Environ(), "OLLAMA_MODELS="+currentPath)
	}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("Error starting: %v", err)
	}

	go processOutput(ctx, stdout, modelName)
	go processOutput(ctx, stderr, modelName)

	err = cmd.Wait()
	if err != nil {
		return fmt.Sprintf("Error finishing: %v", err)
	}

	runtime.EventsEmit(ctx, "download-progress-"+modelName, "100%")
	return "Success"
}

func processOutput(ctx context.Context, pipe io.ReadCloser, modelName string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "%") {
			parts := strings.Split(line, " ")
			for _, part := range parts {
				if strings.Contains(part, "%") {
					runtime.EventsEmit(ctx, "download-progress-"+modelName, part)
					break
				}
			}
		} else {
			runtime.EventsEmit(ctx, "download-progress-"+modelName, line)
		}
	}
}
