package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var placeholderPNG = []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0, 31, 21, 196, 137, 0, 0, 0, 13, 73, 68, 65, 84, 120, 156, 98, 0, 1, 0, 0, 5, 0, 1, 13, 10, 45, 180, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130}

//go:embed frontend/templates
var templatesFS embed.FS

type App struct {
	ctx         context.Context
	projectPath string
	gamesRoot   string
	hostServer  *http.Server
	hostMu      sync.Mutex
	hostPort    int

	// FIX: track own process only
	gameProcess *os.Process
	gameCmd     *exec.Cmd
	gameMu      sync.Mutex
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if exePath, err := os.Executable(); err == nil {
		a.gamesRoot = filepath.Join(filepath.Dir(exePath), "games")
		if _, err := os.Stat(a.gamesRoot); os.IsNotExist(err) {
			if cwd, err := os.Getwd(); err == nil {
				candidate := filepath.Join(cwd, "games")
				if _, err := os.Stat(candidate); err == nil {
					a.gamesRoot = candidate
				} else {
					// dev: editor folder
					candidate2 := filepath.Join(cwd, "..", "games")
					if _, err := os.Stat(candidate2); err == nil {
						a.gamesRoot = candidate2
					}
				}
			}
		}
	}
	runtime.LogInfo(a.ctx, "[APP] KimonoGM - Manieczki edition | gamesRoot: "+a.gamesRoot)
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.StopGame()
	a.hostMu.Lock()
	if a.hostServer != nil {
		_ = a.hostServer.Close()
		a.hostServer = nil
	}
	a.hostMu.Unlock()
}

func (a *App) StopGame() error {
	a.gameMu.Lock()
	defer a.gameMu.Unlock()
	if a.gameProcess != nil {
		runtime.LogInfo(a.ctx, "[APP] Stopping own game process")
		_ = a.gameProcess.Kill()
		a.gameProcess = nil
	}
	if a.gameCmd != nil && a.gameCmd.Process != nil {
		_ = a.gameCmd.Process.Kill()
	}
	a.gameCmd = nil
	return nil
}
