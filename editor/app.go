package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed frontend/templates
var templatesFS embed.FS

type App struct {
	ctx         context.Context
	projectPath string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.LogInfo(a.ctx, "[APP] Janusz Wails wystartował")
}

func (a *App) SetProjectPath(path string) {
	clean := filepath.Clean(path)
	clean = strings.ReplaceAll(clean, "/", "\\")
	a.projectPath = clean
	runtime.LogInfo(a.ctx, "[APP] Ustawiono projectPath: "+clean)
}

func (a *App) GetProjectPath() string {
	return a.projectPath
}

// FIX: Prywatny folder poza repo - AppData/Roaming/KimonoGM/games
func (a *App) GetDefaultProjectPath() string {
	if configDir, err := os.UserConfigDir(); err == nil {
		p := filepath.Join(configDir, "KimonoGM", "games")
		_ = os.MkdirAll(p, 0755)
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, "Documents", "KimonoGM", "games")
		_ = os.MkdirAll(p, 0755)
		return p
	}
	exe, _ := os.Executable()
	root := filepath.Join(filepath.Dir(exe), "..", "..", "..")
	games := filepath.Join(root, "games")
	_ = os.MkdirAll(games, 0755)
	return games
}

func (a *App) OpenProjectFolder(path string) error {
	if path == "" {
		path = a.projectPath
	}
	if path == "" {
		return fmt.Errorf("brak ścieżki projektu")
	}
	cleanPath := filepath.Clean(path)
	cleanPath = strings.ReplaceAll(cleanPath, "/", "\\")
	runtime.LogInfo(a.ctx, fmt.Sprintf("[APP] Otwieram folder: %s", cleanPath))
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", cleanPath)
	case "darwin":
		cmd = exec.Command("open", cleanPath)
	default:
		cmd = exec.Command("xdg-open", cleanPath)
	}
	return cmd.Start()
}
