package ollama

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

func EnsureInstalledAndRun() (*exec.Cmd, error) {
	exePath := getOllamaPath()

	if exePath == "" {
		fmt.Println("Ollama engine not found. Starting background installation...")
		if err := installOllama(); err != nil {
			return nil, fmt.Errorf("installation failed: %w", err)
		}
		fmt.Println("Installation completed successfully.")

		exePath = getOllamaPath()
		if exePath == "" {
			return nil, fmt.Errorf("installed but could not find the executable path")
		}
	}

	fmt.Println("Starting Ollama engine in the background...")
	cmd := exec.Command(exePath, "serve")

	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	return cmd, nil
}

func getOllamaPath() string {
	if runtime.GOOS == "windows" {
		path := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Ollama", "ollama.exe")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		return ""
	}

	if path, err := exec.LookPath("ollama"); err == nil {
		return path
	}
	if _, err := os.Stat("/usr/local/bin/ollama"); err == nil {
		return "/usr/local/bin/ollama"
	}
	if _, err := os.Stat("/usr/bin/ollama"); err == nil {
		return "/usr/bin/ollama"
	}

	return ""
}

func installOllama() error {
	if runtime.GOOS == "windows" {
		psScript := `
			$installerPath = "$env:TEMP\OllamaSetup.exe"
			Invoke-WebRequest -Uri "https://ollama.com/download/OllamaSetup.exe" -OutFile $installerPath
			Start-Process -FilePath $installerPath -ArgumentList "/silent" -Wait
		`
		cmd := exec.Command("powershell", "-Command", psScript)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		return cmd.Run()
	}

	cmd := exec.Command("sh", "-c", "curl -fsSL https://ollama.com/install.sh | sh")
	return cmd.Run()
}
