package main

import (
        "context"
        "embed"
        "net/http"
        "os/exec"
        "sync"

        "github.com/wailsapp/wails/v2/pkg/runtime"
)

var placeholderPNG = []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0, 31, 21, 196, 137, 0, 0, 0, 13, 73, 68, 65, 84, 120, 156, 98, 0, 1, 0, 0, 5, 0, 1, 13, 10, 45, 180, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130}

//go:embed frontend/templates
var templatesFS embed.FS

type App struct {
        ctx         context.Context
        projectPath string
        hostServer  *http.Server
        hostMu      sync.Mutex
        hostPort    int
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
        a.ctx = ctx
        runtime.LogInfo(a.ctx, "[APP] KimonoGM - Manieczki edition")
}
func (a *App) shutdown(ctx context.Context) { exec.Command("taskkill", "/F", "/IM", "dotnet.exe").Run() }
