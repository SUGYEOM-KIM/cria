package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"syscall"
)

type OllamaTagResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

func FetchOllamaModels() []string {
	resp, err := http.Get("http://127.0.0.1:11434/api/tags")
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	var tagResp OllamaTagResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagResp); err != nil {
		return []string{}
	}

	var models []string
	for _, m := range tagResp.Models {
		models = append(models, m.Name)
	}
	return models
}

func DownloadOllamaModel(ctx context.Context, modelName string) string {
	cmd := exec.CommandContext(ctx, "ollama", "pull", modelName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, out.String())
	}
	return "Success"
}

func ChatWithOllama(modelName string, prompt string) string {
	reqBody, _ := json.Marshal(GenerateRequest{
		Model:  modelName,
		Prompt: prompt,
		Stream: false,
	})

	resp, err := http.Post("http://127.0.0.1:11434/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	defer resp.Body.Close()

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return fmt.Sprintf("Error decoding response: %v", err)
	}

	return genResp.Response
}

func RemoveOllamaModel(modelName string) string {
	cmd := exec.Command("ollama", "rm", modelName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, out.String())
	}
	return "Success"
}
