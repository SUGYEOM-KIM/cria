package main

import (
	"context"
	"cria/internal/ollama"
	"fmt"
	"os/exec"
)

type App struct {
	ctx       context.Context
	serverCmd *exec.Cmd
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("Starting Cria agent...")

	go func() {
		cmd, err := ollama.EnsureInstalledAndRun()
		if err != nil {
			fmt.Printf("Fatal error: %v\n", err)
			return
		}
		a.serverCmd = cmd
	}()
}

func (a *App) shutdown(ctx context.Context) {
	fmt.Println("Shutting down Cria...")
	if a.serverCmd != nil && a.serverCmd.Process != nil {
		_ = a.serverCmd.Process.Kill()
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
