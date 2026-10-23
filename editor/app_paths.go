package main

import (
        "fmt"
        "os"
        "os/exec"
        "path/filepath"
        "strings"
        goruntime "runtime"
        "github.com/wailsapp/wails/v2/pkg/runtime"
)



func (a *App) SetProjectPath(path string) {
        clean := filepath.Clean(path)
        clean = strings.ReplaceAll(clean, "/", "\\")
        a.projectPath = clean
}


func (a *App) GetProjectPath() string { return a.projectPath }


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
                return fmt.Errorf("brak sciezki projektu")
        }
        cleanPath := filepath.Clean(path)
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


func (a *App) SelectFolder() (string, error) {
        return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
                Title: "Wybierz folder projektu", CanCreateDirectories: true,
        })
}

// ---------- HELPERS for assets.go, io.go, audio.go - FIX dla wails ----------

