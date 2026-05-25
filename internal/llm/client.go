package llm

import (
	"bufio"
	"bytes"
	"context"
	"cria/internal/logging"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
	Error   string      `json:"error,omitempty"`
}

const systemPrompt = "You are a sharp, friendly AI detective named Cria. Help the user solve their problems with a touch of wit and logical deduction. Always respond in the same language that the user used."

var readyHTTPClient = &http.Client{Timeout: 1500 * time.Millisecond}

func WaitForReady(ctx context.Context, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		resp, err := readyHTTPClient.Get("http://127.0.0.1:11434/api/tags")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	return false
}

func FetchOllamaModels() []string {
	resp, err := http.Get("http://127.0.0.1:11434/api/tags")
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{}
	}

	var tagsResp ollamaTagsResponse
	if err := json.Unmarshal(bodyBytes, &tagsResp); err != nil {
		return []string{}
	}

	var models []string
	for _, m := range tagsResp.Models {
		models = append(models, m.Name)
	}
	return models
}

func ChatWithOllama(modelName string, prompt string) string {
	reqData := chatRequest{
		Model: modelName,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "Error: Could not process request data."
	}

	resp, err := http.Post("http://127.0.0.1:11434/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "Error: Could not connect to Ollama. Make sure it is running."
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Error: Could not read response from Ollama."
	}

	var cResp chatResponse
	if err := json.Unmarshal(bodyBytes, &cResp); err != nil {
		return "Error: Could not parse response from Ollama."
	}

	if cResp.Error != "" {
		return fmt.Sprintf("Ollama Error: %s", cResp.Error)
	}

	return cResp.Message.Content
}

func DownloadOllamaModel(ctx context.Context, modelName string) string {
	cmd := exec.CommandContext(ctx, "ollama", "pull", modelName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if currentPath := os.Getenv("OLLAMA_MODELS"); currentPath != "" {
		cmd.Env = append(os.Environ(), "OLLAMA_MODELS="+currentPath)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Sprintf("Error creating stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Sprintf("Error creating stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("Error starting: %v", err)
	}

	go streamProgress(ctx, stdout, modelName)
	go streamProgress(ctx, stderr, modelName)

	if err := cmd.Wait(); err != nil {
		return fmt.Sprintf("Error finishing: %v", err)
	}

	runtime.EventsEmit(ctx, "download-progress-"+modelName, "100%")
	return "Success"
}

func streamProgress(ctx context.Context, pipe io.ReadCloser, modelName string) {
	scanner := bufio.NewScanner(pipe)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	scanner.Split(scanProgressLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if pct := extractPercent(line); pct != "" {
			runtime.EventsEmit(ctx, "download-progress-"+modelName, pct)
		} else {
			runtime.EventsEmit(ctx, "download-progress-"+modelName, line)
		}
	}

	if err := scanner.Err(); err != nil {
		logging.Errorf("scanner error for model %s: %v", modelName, err)
	}
}

func scanProgressLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i, b := range data {
		if b == '\n' || b == '\r' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func extractPercent(line string) string {
	for _, part := range strings.Fields(line) {
		if strings.HasSuffix(part, "%") {
			return part
		}
	}
	return ""
}

func RemoveOllamaModel(modelName string) string {
	cmd := exec.Command("ollama", "rm", modelName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if currentPath := os.Getenv("OLLAMA_MODELS"); currentPath != "" {
		cmd.Env = append(os.Environ(), "OLLAMA_MODELS="+currentPath)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, out.String())
	}
	return "Success"
}
